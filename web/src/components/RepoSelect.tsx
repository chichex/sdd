import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

const TODOS = '__todos__'

export interface RepoCount {
  id: string
  count: number
}

/** Select custom de repo (nunca un select nativo) con el conteo de trabajos. */
export function RepoSelect({ repos, value, total, onChange }: {
  repos: RepoCount[]
  value: string | null
  total: number
  onChange: (repo: string | null) => void
}) {
  return (
    <Select value={value ?? TODOS} onValueChange={(v) => onChange(v === TODOS ? null : v)}>
      <SelectTrigger aria-label="Repo" className="w-56">
        <SelectValue />
      </SelectTrigger>
      <SelectContent position="popper">
        <SelectItem value={TODOS}>
          <span className="truncate">Todos los repos</span> <span className="ml-auto font-mono text-xs text-fg-muted">{total}</span>
        </SelectItem>
        {repos.map((r) => (
          <SelectItem key={r.id} value={r.id}>
            <span className="truncate">{r.id}</span> <span className="ml-auto font-mono text-xs text-fg-muted">{r.count}</span>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
