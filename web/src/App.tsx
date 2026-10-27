import { useMemo, useRef, useState } from 'react'
import { Toaster } from '@/components/ui/sonner'
import { TooltipProvider } from '@/components/ui/tooltip'
import { Banner } from '@/components/Banner'
import { Board, BoardSkeleton } from '@/components/board/Board'
import { EmptyState } from '@/components/EmptyState'
import { ListaTrabajos } from '@/components/ListaTrabajos'
import { NavegacionCelular, NavegacionDesktop } from '@/components/Navegacion'
import { Refresh } from '@/components/Refresh'
import { Sidebar } from '@/components/Sidebar'
import { useHealth, useLocationPath, useMediaQuery, useNow, useRefresh, useTema } from '@/hooks/ui'
import { useWorkspace } from '@/hooks/useWorkspace'
import type { Trabajo, Workspace } from '@/lib/api'
import { navigate, readFilters, trabajoPath, writeFilters, type Filters } from '@/lib/router'

function conteos(ws: Workspace | null) {
  const counts = new Map<string, number>()
  for (const t of ws?.trabajos ?? []) counts.set(t.repo, (counts.get(t.repo) ?? 0) + 1)
  return (ws?.repos ?? []).filter((r) => !r.sinGit).map((r) => ({ id: r.id, count: counts.get(r.id) ?? 0 }))
}

export default function App() {
  const { search } = useLocationPath()
  const filters = readFilters(search)
  const { data, error, reload } = useWorkspace(filters.filtro)
  const health = useHealth()
  const [tema, setTema] = useTema()
  const desktop = useMediaQuery('(min-width: 768px)')
  const now = useNow(30_000)
  const { refreshing, failed, refresh } = useRefresh(reload)
  const [menu, setMenu] = useState(false)
  const menuRef = useRef<HTMLButtonElement>(null)
  // El skeleton dura hasta el primer snapshot con el escaneo inicial terminado.
  const [listo, setListo] = useState(false)
  if (!listo && data && !data.scanning) setListo(true)

  const repos = useMemo(() => conteos(data), [data])
  const trabajos = (data?.trabajos ?? []).filter((t) => !filters.repo || t.repo === filters.repo)
  const set = (patch: Partial<Filters>) => writeFilters({ ...filters, ...patch })
  const abrir = (t: Trabajo) => navigate(trabajoPath(t.id, filters))
  const refreshButton = (
    <Refresh fetchedAt={data?.github.fetchedAt ?? null} now={now} refreshing={refreshing} failed={failed} onRefresh={() => void refresh()} />
  )
  const nav = { repos, total: data?.trabajos.length ?? 0, repo: filters.repo, filtro: filters.filtro, tema, refresh: refreshButton }

  let contenido: React.ReactNode
  if (!listo) contenido = <BoardSkeleton />
  else if (data && data.repos.length === 0)
    contenido = (
      <EmptyState comando="sdd --dir <otra carpeta>">No hay repos con .sdd/project.md bajo {data.root}. Corré el visor en otra carpeta:</EmptyState>
    )
  else if (desktop) contenido = <Board trabajos={trabajos} now={now} onOpen={abrir} />
  else contenido = <ListaTrabajos trabajos={trabajos} now={now} onOpen={abrir} />

  return (
    <TooltipProvider>
      <div className="flex min-h-dvh flex-col bg-bg text-fg">
        {desktop ? (
          <NavegacionDesktop {...nav} onRepo={(repo) => set({ repo })} onFiltro={(filtro) => set({ filtro })} onTema={setTema} />
        ) : (
          <NavegacionCelular repo={filters.repo} onMenu={() => setMenu(true)} refresh={refreshButton} menuRef={menuRef} />
        )}
        {data && !data.github.ok && (
          <Banner tone="warning" onRetry={() => void refresh()}>
            Sin GitHub: mostrando solo lo local
          </Banner>
        )}
        {error && (
          <Banner tone="error" onRetry={() => void reload()}>
            No se pudo leer el workspace ({error}). Se muestra el último dato bueno.
          </Banner>
        )}
        <main className="min-w-0 flex-1">{contenido}</main>
        {!desktop && (
          <Sidebar
            open={menu}
            onOpenChange={setMenu}
            returnFocus={menuRef}
            {...nav}
            version={health?.version ?? null}
            root={health?.root ?? data?.root ?? null}
            onRepo={(repo) => set({ repo })}
            onFiltro={(filtro) => set({ filtro })}
            onTema={setTema}
          />
        )}
        <Toaster position={desktop ? 'bottom-right' : 'bottom-center'} theme={tema === 'sistema' ? 'system' : tema === 'oscuro' ? 'dark' : 'light'} />
      </div>
    </TooltipProvider>
  )
}
