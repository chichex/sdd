import { cn } from 'cn'
import type { Nodo, Trabajo } from '@/lib/api'
import { edadCorta } from '@/lib/time'
import { etapaInfo, refCorta } from '../etapas'
import { ChipRef, EtapaDot, SinTrazaBadge, SubEstadoBadge } from '../marcas'

function etiqueta(t: Trabajo, n: Nodo): string {
  return `${t.titulo} · ${etapaInfo(n.etapa).nombre.toLowerCase()} · ${n.subEstado}`
}

function referencia(t: Trabajo, n: Nodo): string {
  if (n.refs.pr !== null) return `#${n.refs.pr}`
  if (n.etapa === 'grill' && n.refs.grill) return refCorta(n.refs.grill)
  return refCorta(t.ancla.ref)
}

/** Nodo del board: tres filas de una línea (título, referencia y antigüedad, sub-estado). */
export function NodoCard({ trabajo, nodo, now, onOpen }: { trabajo: Trabajo; nodo: Nodo; now: Date; onOpen: () => void }) {
  return (
    <button
      type="button"
      data-nodo
      data-stale={nodo.stale ? 'true' : undefined}
      aria-label={etiqueta(trabajo, nodo)}
      onClick={onOpen}
      className={cn(
        'relative z-10 flex w-full cursor-pointer flex-col gap-1 rounded-md border border-border bg-bg p-3 text-left shadow-sm hover:border-border-strong',
        nodo.stale && 'text-fg-muted',
      )}
    >
      <span className="flex min-w-0 items-center gap-2">
        <EtapaDot etapa={nodo.etapa} />
        <span className="truncate text-sm font-semibold">{trabajo.titulo}</span>
      </span>
      <span className="flex min-w-0 items-center justify-between gap-2">
        <ChipRef ancla={trabajo.ancla.tipo}>{nodo.capa ? `${nodo.capa} ${referencia(trabajo, nodo)}` : referencia(trabajo, nodo)}</ChipRef>
        <span className="shrink-0 font-mono text-xs text-fg-muted tabular-nums">{edadCorta(nodo.ts, now)}</span>
      </span>
      <span className="flex min-w-0 items-center gap-2 overflow-hidden">
        <SubEstadoBadge subEstado={nodo.subEstado} />
        {trabajo.sinHistoria && <SinTrazaBadge />}
        {nodo.tags && nodo.tags.length > 0 && <ChipRef>{nodo.tags.join(' ')}</ChipRef>}
      </span>
    </button>
  )
}

/** Stack de PRs o merges: nodos encadenados n/N de una línea cada uno. */
export function StackCard({ trabajo, nodos, now, onOpen }: { trabajo: Trabajo; nodos: Nodo[]; now: Date; onOpen: () => void }) {
  return (
    <div className="relative z-10 flex flex-col gap-1 before:absolute before:top-4 before:bottom-4 before:left-[15px] before:w-px before:bg-border-strong">
      {nodos.map((n) => (
        <button
          key={`${n.etapa}-${n.refs.pr ?? n.capa}`}
          type="button"
          data-nodo
          data-stale={n.stale ? 'true' : undefined}
          aria-label={etiqueta(trabajo, n)}
          onClick={onOpen}
          className={cn(
            'relative flex h-8 w-full cursor-pointer items-center gap-2 rounded-md border border-border bg-bg px-3 text-left hover:border-border-strong pointer-coarse:h-11',
            n.stale && 'text-fg-muted',
          )}
        >
          <EtapaDot etapa={n.etapa} />
          <span className="font-mono text-xs tabular-nums">{n.capa}</span>
          <span className="truncate font-mono text-xs text-fg-secondary">{referencia(trabajo, n)}</span>
          <span className="ml-auto flex shrink-0 items-center gap-2">
            <SubEstadoBadge subEstado={n.subEstado} />
            <span className="font-mono text-xs text-fg-muted tabular-nums">{edadCorta(n.ts, now)}</span>
          </span>
        </button>
      ))}
    </div>
  )
}
