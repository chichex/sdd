import { Menu } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { Tema } from '@/lib/theme'
import { RepoSelect, type RepoCount } from './RepoSelect'
import { TEMAS } from './opciones'
import { Segmented } from './Segmented'

export interface NavProps {
  repos: RepoCount[]
  total: number
  repo: string | null
  tema: Tema
  onRepo: (repo: string | null) => void
  onTema: (t: Tema) => void
  refresh: React.ReactNode
}

/** Cabecera de una sola fila en desktop (≥768 px). */
export function NavegacionDesktop(p: NavProps) {
  return (
    <header className="flex h-12 items-center gap-3 border-b bg-bg px-6">
      <span className="text-base font-semibold">sdd</span>
      <RepoSelect repos={p.repos} total={p.total} value={p.repo} onChange={p.onRepo} />
      <div className="flex-1" />
      {p.refresh}
      <Segmented label="Tema" value={p.tema} options={TEMAS} onChange={p.onTema} />
    </header>
  )
}

/** Cabecera de celular: hamburguesa, «sdd · <repo>» y Refresh, sin segunda fila. */
export function NavegacionCelular({ repo, onMenu, refresh, menuRef }: {
  repo: string | null
  onMenu: () => void
  refresh: React.ReactNode
  menuRef: React.Ref<HTMLButtonElement>
}) {
  return (
    <header className="flex h-12 items-center gap-1 border-b bg-bg pr-2 pl-1 pt-[env(safe-area-inset-top)]">
      <Button ref={menuRef} variant="ghost" size="icon" aria-label="Menú" onClick={onMenu} className="size-11">
        <Menu aria-hidden className="size-5" />
      </Button>
      <span className="min-w-0 flex-1 truncate text-base font-semibold">{`sdd · ${repo ?? 'todos los repos'}`}</span>
      {refresh}
    </header>
  )
}
