import type { Detalle } from '@/lib/api'
import { fechaLarga } from '@/lib/time'
import { GANTT_BARRA, GANTT_EJE, GANTT_LABEL, ganttLayout } from './gantt'

const FILL: Record<string, string> = {
  idea: 'fill-stage-idea',
  grill: 'fill-stage-grill',
  spec: 'fill-stage-spec',
  run: 'fill-stage-run',
  pr: 'fill-stage-pr',
  mergeado: 'fill-stage-merged',
  publicado: 'fill-stage-published',
}

const DIA = 86_400_000

/** Gantt SVG: barras por etapa y rama, eje de fechas en mono y línea de hoy; scroll horizontal propio. */
export function Gantt({ detalle, hoy }: { detalle: Detalle; hoy: Date }) {
  if (detalle.gantt.length === 0) {
    return <p className="text-sm text-fg-muted">Sin historia en el log: este trabajo es anterior a events-v1.</p>
  }
  const t0 = Math.min(...detalle.gantt.map((b) => Date.parse(b.inicio)))
  const dias = Math.max((hoy.getTime() - t0) / DIA, 1)
  const layout = ganttLayout(detalle, Math.max(640, GANTT_LABEL + dias * 56), hoy)
  return (
    <div className="overflow-x-auto">
      <svg role="img" aria-label="Gantt" width={layout.ancho} height={layout.alto} viewBox={`0 0 ${layout.ancho} ${layout.alto}`} className="block">
        {layout.ticks.map((t) => (
          <text key={t.x} x={t.x} y={14} textAnchor="middle" className="fill-fg-muted font-mono text-xs tabular-nums">
            {t.label}
          </text>
        ))}
        {layout.barras.map((b) => (
          <g key={`${b.rama}-${b.etapa}`} data-barra>
            <title>{`${b.etiqueta} · ${b.estado}`}</title>
            <text x={8} y={b.y + GANTT_BARRA / 2 + 4} className="fill-fg text-sm">
              {b.etiqueta}
            </text>
            <rect x={b.x} y={b.y} width={b.w} height={GANTT_BARRA} rx={4} className={FILL[b.etapa]} />
            {b.rama === 'principal' && (
              <text x={b.x + b.w + 6} y={b.y + GANTT_BARRA / 2 + 4} className="fill-fg-secondary text-xs">
                {b.estado}
              </text>
            )}
          </g>
        ))}
        <line data-hoy x1={layout.hoyX} x2={layout.hoyX} y1={GANTT_EJE - 4} y2={layout.alto} className="stroke-primary-text" strokeDasharray="4 3">
          <title>{`hoy · ${fechaLarga(hoy.toISOString())}`}</title>
        </line>
      </svg>
    </div>
  )
}
