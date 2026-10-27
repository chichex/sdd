// Etapas del flujo con su nombre, su color de relleno y el par semántico del
// badge de cada sub-estado (DESIGN_SYSTEM.md: Badge y colores de etapa).
import type { Etapa } from '@/lib/api'

export const ETAPA_INFO: ReadonlyArray<{ id: Etapa; nombre: string; dot: string }> = [
  { id: 'idea', nombre: 'Idea', dot: 'bg-stage-idea' },
  { id: 'grill', nombre: 'Grill', dot: 'bg-stage-grill' },
  { id: 'spec', nombre: 'Spec', dot: 'bg-stage-spec' },
  { id: 'run', nombre: 'Run', dot: 'bg-stage-run' },
  { id: 'pr', nombre: 'PR', dot: 'bg-stage-pr' },
  { id: 'mergeado', nombre: 'Mergeado', dot: 'bg-stage-merged' },
  { id: 'publicado', nombre: 'Publicado', dot: 'bg-stage-published' },
]

export function etapaInfo(id: Etapa): { id: Etapa; nombre: string; dot: string } {
  return ETAPA_INFO.find((e) => e.id === id) ?? ETAPA_INFO[0]
}

export type BadgeTone = 'neutral' | 'info' | 'success' | 'warning' | 'error'

const TONES: Record<string, BadgeTone> = {
  validado: 'success',
  mergeado: 'success',
  taggeado: 'success',
  implementada: 'success',
  finalizado: 'success',
  'cambios pedidos': 'warning',
  'checks rojos': 'error',
  conflicto: 'error',
  abierto: 'info',
  corriendo: 'info',
  'con PR': 'info',
  aprobada: 'info',
  commiteado: 'info',
}

/** Par semántico de un sub-estado; «n/N» es éxito solo si está completo. */
export function badgeTone(subEstado: string): BadgeTone {
  if (subEstado.startsWith('fallado')) return 'error'
  const frac = /^(\d+)\/(\d+)$/.exec(subEstado)
  if (frac) return frac[1] === frac[2] ? 'success' : 'info'
  return TONES[subEstado] ?? 'neutral'
}

/** Texto corto de la referencia de un ancla: #NN, el slug del grill o el archivo de la spec. */
export function refCorta(ref: string): string {
  if (ref.startsWith('#') || /#\d+$/.test(ref)) return ref
  const parts = ref.split('/')
  return parts[parts.length - 1] ?? ref
}
