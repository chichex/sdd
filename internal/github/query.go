package github

import "time"

// repoQuery trae en una sola consulta los issues abiertos, los PRs abiertos y
// los 100 PRs más recientes con lo que el gate de sdd-land necesita.
const repoQuery = `query($owner: String!, $name: String!) {
  repository(owner: $owner, name: $name) {
    issues(states: OPEN, first: 100, orderBy: {field: UPDATED_AT, direction: DESC}) {
      totalCount
      nodes { number title url updatedAt }
    }
    openPRs: pullRequests(states: OPEN, first: 100, orderBy: {field: UPDATED_AT, direction: DESC}) { nodes { ...pr } }
    recentPRs: pullRequests(first: 100, orderBy: {field: UPDATED_AT, direction: DESC}) { nodes { ...pr } }
  }
}
fragment pr on PullRequest {
  number title url state isDraft mergeable reviewDecision mergedAt updatedAt body
  mergeCommit { oid }
  reviewThreads(first: 100) { nodes { isResolved } }
  commits(last: 1) { nodes { commit { statusCheckRollup { contexts(first: 100) { nodes {
    __typename
    ... on CheckRun { name status conclusion startedAt completedAt checkSuite { workflowRun { workflow { name } } } }
    ... on StatusContext { context state createdAt }
  } } } } } }
}`

type gqlResponse struct {
	Data struct {
		Repository *gqlRepository `json:"repository"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type gqlRepository struct {
	Issues struct {
		TotalCount int     `json:"totalCount"`
		Nodes      []Issue `json:"nodes"`
	} `json:"issues"`
	OpenPRs   gqlPRs `json:"openPRs"`
	RecentPRs gqlPRs `json:"recentPRs"`
}

type gqlPRs struct {
	Nodes []gqlPR `json:"nodes"`
}

type gqlPR struct {
	Number         int    `json:"number"`
	Title          string `json:"title"`
	URL            string `json:"url"`
	State          string `json:"state"`
	IsDraft        bool   `json:"isDraft"`
	Mergeable      string `json:"mergeable"`
	ReviewDecision string `json:"reviewDecision"`
	MergedAt       string `json:"mergedAt"`
	UpdatedAt      string `json:"updatedAt"`
	Body           string `json:"body"`
	MergeCommit    *struct {
		OID string `json:"oid"`
	} `json:"mergeCommit"`
	ReviewThreads struct {
		Nodes []struct {
			IsResolved bool `json:"isResolved"`
		} `json:"nodes"`
	} `json:"reviewThreads"`
	Commits struct {
		Nodes []struct {
			Commit struct {
				StatusCheckRollup *struct {
					Contexts struct {
						Nodes []gqlContext `json:"nodes"`
					} `json:"contexts"`
				} `json:"statusCheckRollup"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"commits"`
}

type gqlContext struct {
	Typename    string `json:"__typename"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Conclusion  string `json:"conclusion"`
	StartedAt   string `json:"startedAt"`
	CompletedAt string `json:"completedAt"`
	CheckSuite  *struct {
		WorkflowRun *struct {
			Workflow struct {
				Name string `json:"name"`
			} `json:"workflow"`
		} `json:"workflowRun"`
	} `json:"checkSuite"`
	Context   string `json:"context"`
	State     string `json:"state"`
	CreatedAt string `json:"createdAt"`
}

// check lleva un contexto GraphQL a la forma de `gh pr view`: el workflow sale
// del check suite y un StatusContext usa createdAt como startedAt.
func (c gqlContext) check() Check {
	out := Check{Typename: c.Typename, Name: c.Name, Status: c.Status, Conclusion: c.Conclusion,
		StartedAt: c.StartedAt, CompletedAt: c.CompletedAt, Context: c.Context, State: c.State}
	if c.Typename == "StatusContext" {
		out.StartedAt = c.CreatedAt
	}
	if c.CheckSuite != nil && c.CheckSuite.WorkflowRun != nil {
		out.WorkflowName = c.CheckSuite.WorkflowRun.Workflow.Name
	}
	return out
}

func (p gqlPR) pr() PR {
	out := PR{Number: p.Number, Title: p.Title, URL: p.URL, State: p.State, IsDraft: p.IsDraft, Mergeable: p.Mergeable,
		ReviewDecision: p.ReviewDecision, MergedAt: p.MergedAt, Body: p.Body, UpdatedAt: p.UpdatedAt}
	if p.MergeCommit != nil {
		out.MergeCommit = p.MergeCommit.OID
	}
	for _, t := range p.ReviewThreads.Nodes {
		if !t.IsResolved {
			out.UnresolvedThreads++
		}
	}
	for _, n := range p.Commits.Nodes {
		if n.Commit.StatusCheckRollup == nil {
			continue
		}
		for _, c := range n.Commit.StatusCheckRollup.Contexts.Nodes {
			out.Checks = append(out.Checks, c.check())
		}
	}
	return out
}

func (r *gqlRepository) snapshot(repo string, now time.Time) *Snapshot {
	s := &Snapshot{Repo: repo, FetchedAt: now.UTC().Format("2006-01-02T15:04:05Z"), Issues: r.Issues.Nodes,
		IssuesComplete: r.Issues.TotalCount <= len(r.Issues.Nodes)}
	if s.Issues == nil {
		s.Issues = []Issue{}
	}
	seen := map[int]bool{}
	for _, list := range []gqlPRs{r.OpenPRs, r.RecentPRs} {
		for _, p := range list.Nodes {
			if !seen[p.Number] {
				seen[p.Number] = true
				s.PRs = append(s.PRs, p.pr())
			}
		}
	}
	return s
}
