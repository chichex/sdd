// Layout del Gantt del detalle: una barra por etapa y por rama sobre un eje de
// fechas. Es puro (sin DOM ni reloj implícito) para poder testearlo.
import type { Detalle, Etapa } from '@/lib/api'
import { fechaCorta } from '@/lib/time'
import { etapaInfo } from '../etapas'

export const GANTT_LABEL = 160
export const GANTT_EJE = 24
export const GANTT_FILA = 32
export const GANTT_BARRA = 20
const MARGEN = 8

export interface BarraLayout {
  rama: string
  etapa: Etapa
  estado: string
  etiqueta: string
  x: number
  y: number
  w: number
  abierta: boolean
}

export interface GanttLayout {
  ancho: number
  alto: number
  barras: BarraLayout[]
  ticks: Array<{ x: number; label: string }>
  hoyX: number
}

function etiquetaRama(rama: string, etapa: Etapa, estado: string): string {
  const nombre = etapaInfo(etapa).nombre
  const m = /^(?:run|spec)-(\d+)$/.exec(rama)
  return m ? `${nombre} ${m[1]} · ${estado}` : nombre
}

/** Posiciona las barras de d.gantt en un ancho dado; las abiertas llegan hasta hoy. */
export function ganttLayout(d: Detalle, ancho: number, hoy: Date): GanttLayout {
  if (d.gantt.length === 0) return { ancho, alto: GANTT_EJE, barras: [], ticks: [], hoyX: ancho - MARGEN }
  const tHoy = hoy.getTime()
  const inicios = d.gantt.map((b) => Date.parse(b.inicio))
  const fines = d.gantt.map((b) => (b.fin ? Date.parse(b.fin) : tHoy))
  const t0 = Math.min(...inicios)
  const t1 = Math.max(tHoy, ...fines)
  const span = Math.max(t1 - t0, 1)
  const area = ancho - GANTT_LABEL - MARGEN
  const x = (ms: number) => GANTT_LABEL + ((ms - t0) / span) * area
  const barras = d.gantt.map((b, i) => {
    const xi = x(inicios[i] ?? t0)
    const xf = x(fines[i] ?? tHoy)
    return {
      rama: b.rama,
      etapa: b.etapa,
      estado: b.estado,
      etiqueta: etiquetaRama(b.rama, b.etapa, b.estado),
      x: xi,
      y: GANTT_EJE + i * GANTT_FILA + (GANTT_FILA - GANTT_BARRA) / 2,
      w: Math.max(xf - xi, 2),
      abierta: b.fin === null,
    }
  })
  const n = 5
  const ticks = Array.from({ length: n }, (_, i) => {
    const ms = t0 + (span * i) / (n - 1)
    return { x: x(ms), label: fechaCorta(new Date(ms).toISOString()) }
  })
  return { ancho, alto: GANTT_EJE + d.gantt.length * GANTT_FILA, barras, ticks, hoyX: x(tHoy) }
}
