import { cn } from 'cn'
import type { Trabajo } from '@/lib/api'
import { fechaCorta } from '@/lib/time'
import { etapaInfo } from '../etapas'
import { grafoLayout } from './grafo'

const FILL: Record<string, string> = {
  idea: 'fill-stage-idea',
  grill: 'fill-stage-grill',
  spec: 'fill-stage-spec',
  run: 'fill-stage-run',
  pr: 'fill-stage-pr',
  mergeado: 'fill-stage-merged',
  publicado: 'fill-stage-published',
}

/** Grafo SVG del trabajo: un nodo por etapa alcanzada, ramas cerradas como bifurcaciones. */
export function Grafo({ trabajo }: { trabajo: Trabajo }) {
  const { ancho, alto, nodos, aristas } = grafoLayout(trabajo)
  return (
    <div className="overflow-x-auto">
      <svg role="img" aria-label="Grafo" width={ancho} height={alto} viewBox={`0 0 ${ancho} ${alto}`} className="block">
        {aristas.map((a) => (
          <path
            key={a.key}
            d={a.curva ? `M ${a.x1} ${a.y1} C ${a.x1 + 60} ${a.y1}, ${a.x2 - 60} ${a.y2}, ${a.x2} ${a.y2}` : `M ${a.x1} ${a.y1} L ${a.x2} ${a.y2}`}
            className="fill-none stroke-border-strong"
            strokeWidth={1.5}
          />
        ))}
        {nodos.map((g) => {
          const etiqueta = g.nodo.capa ? `${etapaInfo(g.nodo.etapa).nombre} ${g.nodo.capa}` : etapaInfo(g.nodo.etapa).nombre
          const titulo = `${etiqueta} · ${g.nodo.subEstado}${g.nodo.refs.pr !== null ? ` · PR ${g.nodo.refs.pr}` : ''}${g.nodo.ruta ? ` · ${g.nodo.ruta}` : ''}`
          const cuerpo = (
            <g data-grafo-nodo data-rama={g.rama} data-activo={g.activo ? 'true' : undefined}>
              <title>{titulo}</title>
              <circle cx={g.x} cy={g.y} r={16} className={cn(FILL[g.nodo.etapa], g.activo ? 'stroke-primary-text' : 'stroke-border-strong')} strokeWidth={g.activo ? 3 : 1} />
              <text x={g.x} y={g.y + 32} textAnchor="middle" className={cn('text-xs', g.activo ? 'fill-primary-text font-semibold' : 'fill-fg')}>
                {etiqueta}
              </text>
              <text x={g.x} y={g.y + 46} textAnchor="middle" className="fill-fg-secondary text-xs">
                {g.nodo.subEstado}
              </text>
              <text x={g.x} y={g.y + 60} textAnchor="middle" className="fill-fg-muted font-mono text-xs">
                {fechaCorta(g.nodo.ts)}
              </text>
            </g>
          )
          return g.nodo.url ? (
            <a key={g.key} href={g.nodo.url} target="_blank" rel="noreferrer" aria-label={titulo}>
              {cuerpo}
            </a>
          ) : (
            <g key={g.key}>{cuerpo}</g>
          )
        })}
      </svg>
    </div>
  )
}
