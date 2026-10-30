import { Check } from 'lucide-react'
import { cn } from 'cn'
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import type { Tema } from '@/lib/theme'
import type { EtapasProps, MostrarProps } from './BarraFiltros'
import { ETAPA_INFO } from './etapas'
import { EtapaDot } from './marcas'
import { CATEGORIA_INFO, TEMAS } from './opciones'
import type { RepoCount } from './RepoSelect'
import { Segmented } from './Segmented'

/** Fila de 44 px con tilde; `toggle` la vuelve un botón de alternar (aria-pressed). */
function Fila({ activa, toggle, children, onClick }: { activa: boolean; toggle?: boolean; children: React.ReactNode; onClick: () => void }) {
  return (
    <button
      type="button"
      aria-pressed={toggle ? activa : undefined}
      onClick={onClick}
      className={cn('flex h-11 w-full cursor-pointer items-center gap-2 rounded-md px-3 text-left text-base', activa ? 'bg-primary-subtle text-primary-text' : 'hover:bg-bg-muted')}
    >
      <Check aria-hidden className={cn('size-4 shrink-0', !activa && 'invisible')} />
      {children}
    </button>
  )
}

function Seccion({ titulo, children }: { titulo: string; children: React.ReactNode }) {
  return (
    <section className="flex flex-col gap-1 px-2">
      <h3 className="px-3 pt-2 text-xs font-semibold tracking-wide text-fg-secondary uppercase">{titulo}</h3>
      {children}
    </section>
  )
}

/** Menú lateral de celular: repo, categorías, etapas y tema; elegir un repo lo cierra, tocar una categoría o una etapa no. */
export function Sidebar({ open, onOpenChange, returnFocus, repos, total, repo, tema, version, root, onRepo, onTema, mostrar, etapas }: {
  open: boolean
  onOpenChange: (open: boolean) => void
  returnFocus: React.RefObject<HTMLButtonElement | null>
  repos: RepoCount[]
  total: number
  repo: string | null
  tema: Tema
  version: string | null
  root: string | null
  onRepo: (repo: string | null) => void
  onTema: (t: Tema) => void
  mostrar: MostrarProps
  etapas: EtapasProps
}) {
  const elegir = <T,>(fn: (v: T) => void) => (v: T) => {
    fn(v)
    onOpenChange(false)
  }
  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side="left"
        onCloseAutoFocus={(e) => {
          e.preventDefault()
          returnFocus.current?.focus() // el foco vuelve a la hamburguesa
        }}
        className="w-[280px] max-w-[85%] gap-2 overflow-y-auto pt-[env(safe-area-inset-top)] pb-[env(safe-area-inset-bottom)] pl-[env(safe-area-inset-left)] sm:max-w-[280px]">
        <SheetHeader className="px-4 py-3">
          <SheetTitle>sdd</SheetTitle>
          <SheetDescription className="sr-only">Repo, categorías, etapas y tema del tablero</SheetDescription>
        </SheetHeader>
        <Seccion titulo="Repo">
          <Fila activa={repo === null} onClick={() => elegir(onRepo)(null)}>
            <span className="flex-1 truncate">Todos los repos</span> <span className="font-mono text-xs text-fg-muted">{total}</span>
          </Fila>
          {repos.map((r) => (
            <Fila key={r.id} activa={repo === r.id} onClick={() => elegir(onRepo)(r.id)}>
              <span className="flex-1 truncate">{r.id}</span> <span className="font-mono text-xs text-fg-muted">{r.count}</span>
            </Fila>
          ))}
        </Seccion>
        <Seccion titulo="Mostrar">
          {CATEGORIA_INFO.map((c) => (
            <Fila key={c.id} activa={mostrar.mostrar.includes(c.id)} toggle onClick={() => mostrar.onMostrar(c.id)}>
              <span className="flex-1 truncate">{c.nombre}</span> <span className="font-mono text-xs text-fg-muted">{mostrar.categorias[c.id]}</span>
            </Fila>
          ))}
        </Seccion>
        <Seccion titulo="Etapas">
          {ETAPA_INFO.map((e) => (
            <Fila key={e.id} activa={etapas.etapas.includes(e.id)} toggle onClick={() => etapas.onEtapa(e.id)}>
              <EtapaDot etapa={e.id} />
              <span className="flex-1 truncate">{e.nombre}</span> <span className="font-mono text-xs text-fg-muted">{etapas.conteos.get(e.id) ?? 0}</span>
            </Fila>
          ))}
        </Seccion>
        <Seccion titulo="Tema">
          <Segmented label="Tema" value={tema} options={TEMAS} onChange={onTema} className="w-full" />
        </Seccion>
        <footer className="mt-auto flex flex-col gap-0.5 border-t px-4 py-3 font-mono text-xs text-fg-muted">
          {version && <span>{version}</span>}
          {root && <span className="truncate">{root}</span>}
        </footer>
      </SheetContent>
    </Sheet>
  )
}
