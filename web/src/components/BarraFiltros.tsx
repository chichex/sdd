// Barra de filtros bajo la cabecera: búsqueda por texto y, en desktop, las
// chips de etapa; en celular las etapas viven en el Sidebar.
import { useRef } from 'react'
import { cn } from 'cn'
import { Search, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { Etapa } from '@/lib/api'
import { ETAPA_INFO } from './etapas'

/** Selección de etapas: lo elegido, el conteo por etapa y el alternado de una. */
export interface EtapasProps {
  etapas: readonly Etapa[]
  conteos: ReadonlyMap<Etapa, number>
  onEtapa: (e: Etapa) => void
}

function Busqueda({ q, onQ, desktop }: { q: string; onQ: (q: string) => void; desktop: boolean }) {
  const input = useRef<HTMLInputElement>(null)
  const limpiar = () => {
    onQ('')
    input.current?.focus()
  }
  return (
    <div className={cn('relative', desktop ? 'w-[280px] shrink-0' : 'w-full')}>
      <Search aria-hidden className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-fg-muted" />
      <Input
        ref={input}
        type="search"
        aria-label="Buscar"
        placeholder="Buscar trabajo, #NN o repo"
        value={q}
        onChange={(e) => onQ(e.target.value)}
        className={cn('pr-9 pl-8 [&::-webkit-search-cancel-button]:appearance-none', !desktop && 'h-11 text-lg')}
      />
      {q && (
        <Button variant="ghost" size="icon-sm" aria-label="Limpiar búsqueda" onClick={limpiar} className="absolute top-1/2 right-0.5 -translate-y-1/2">
          <X aria-hidden />
        </Button>
      )}
    </div>
  )
}

/** Chips de etapa de desktop: multi-selección con aria-pressed, nombre y conteo. */
function ChipsEtapa({ etapas, conteos, onEtapa }: EtapasProps) {
  return (
    <div role="group" aria-label="Etapas" className="flex min-w-0 items-center gap-2 overflow-x-auto">
      {ETAPA_INFO.map((e) => {
        const activa = etapas.includes(e.id)
        return (
          <button
            key={e.id}
            type="button"
            aria-pressed={activa}
            onClick={() => onEtapa(e.id)}
            className={cn(
              'inline-flex h-7 shrink-0 cursor-pointer items-center gap-1.5 rounded-md border border-border px-2 text-xs font-medium whitespace-nowrap text-fg-secondary hover:text-fg',
              activa && 'bg-primary-subtle text-primary-text hover:text-primary-text',
            )}
          >
            {e.nombre} <span className="font-mono tabular-nums">{conteos.get(e.id) ?? 0}</span>
          </button>
        )
      })}
    </div>
  )
}

export function BarraFiltros({ desktop, q, onQ, ...etapas }: { desktop: boolean; q: string; onQ: (q: string) => void } & EtapasProps) {
  return (
    <div className={cn('flex items-center gap-3 border-b bg-bg', desktop ? 'h-10 px-6' : 'px-4 py-2')}>
      <Busqueda q={q} onQ={onQ} desktop={desktop} />
      {desktop && <ChipsEtapa {...etapas} />}
    </div>
  )
}
