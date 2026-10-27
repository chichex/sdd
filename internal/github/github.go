// Package github lee GitHub para el visor ejecutando `gh` como subproceso,
// siempre en modo de solo lectura: una consulta GraphQL por repo y por poll,
// con caché en disco, polling y degradado sin red ni login.
package github

// Razones de degradado de la fuente GitHub.
const (
	ReasonSinGH     = "sin-gh"
	ReasonSinLogin  = "sin-login"
	ReasonSinRemote = "sin-remote"
	ReasonRed       = "red"
	ReasonRateLimit = "rate-limit"
)

// Check es un check del rollup con la forma de `gh pr view --json statusCheckRollup`.
type Check struct {
	Typename     string `json:"__typename"`
	Name         string `json:"name,omitempty"`
	WorkflowName string `json:"workflowName,omitempty"`
	Status       string `json:"status,omitempty"`
	Conclusion   string `json:"conclusion,omitempty"`
	StartedAt    string `json:"startedAt,omitempty"`
	CompletedAt  string `json:"completedAt,omitempty"`
	Context      string `json:"context,omitempty"`
	State        string `json:"state,omitempty"`
}

// Issue es un issue abierto.
type Issue struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	UpdatedAt string `json:"updatedAt"`
}

// PR es un pull request con lo que el visor necesita para su sub-estado,
// su merge y su vinculación por body.
type PR struct {
	Number            int     `json:"number"`
	Title             string  `json:"title"`
	URL               string  `json:"url"`
	State             string  `json:"state"`
	IsDraft           bool    `json:"isDraft"`
	Mergeable         string  `json:"mergeable"`
	ReviewDecision    string  `json:"reviewDecision"`
	MergedAt          string  `json:"mergedAt"`
	MergeCommit       string  `json:"mergeCommit"`
	Body              string  `json:"body"`
	UpdatedAt         string  `json:"updatedAt"`
	UnresolvedThreads int     `json:"unresolvedThreads"`
	Checks            []Check `json:"checks"`
}

// Snapshot es lo que se leyó de GitHub para un repo en un poll.
// IssuesComplete dice si la lista de issues abiertos está completa: un issue
// que no figura en una lista completa está cerrado.
type Snapshot struct {
	Repo           string  `json:"repo"`
	FetchedAt      string  `json:"fetchedAt"`
	Issues         []Issue `json:"issues"`
	IssuesComplete bool    `json:"issuesComplete"`
	PRs            []PR    `json:"prs"`
}

// PR devuelve el PR por número.
func (s *Snapshot) PR(number int) (PR, bool) {
	if s == nil {
		return PR{}, false
	}
	for _, pr := range s.PRs {
		if pr.Number == number {
			return pr, true
		}
	}
	return PR{}, false
}
