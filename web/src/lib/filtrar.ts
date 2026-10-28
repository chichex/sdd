// Filtrado y agrupado del lado del cliente sobre la respuesta de
// /api/workspace: búsqueda por texto, etapas elegidas y grupos por repo.
import { ETAPAS, type Trabajo } from './api'
import type { Filters } from './router'

export interface Grupo {
  repo: string
  trabajos: Trabajo[]
}

function comparar(a: string, b: string): number {
  if (a === b) return 0
  return a < b ? -1 : 1
}

/**
 * Grupos por repo, ordenados por id ascendente; dentro de cada uno, la etapa
 * más avanzada primero y, a igual etapa, la actividad más reciente.
 */
export function agrupar(trabajos: readonly Trabajo[]): Grupo[] {
  const porRepo = new Map<string, Trabajo[]>()
  for (const t of trabajos) porRepo.set(t.repo, [...(porRepo.get(t.repo) ?? []), t])
  return [...porRepo.keys()].sort(comparar).map((repo) => ({
    repo,
    trabajos: (porRepo.get(repo) ?? []).toSorted(
      (a, b) => ETAPAS.indexOf(b.etapa) - ETAPAS.indexOf(a.etapa) || comparar(b.ultimaActividad, a.ultimaActividad),
    ),
  }))
}

/** Minúsculas y sin acentos (NFD sin marcas combinantes): «Ñandú» → «nandu». */
export function normalizar(s: string): string {
  return s.normalize('NFD').replace(/\p{M}/gu, '').toLowerCase()
}

/** La búsqueda es substring de título, repo, referencia del ancla o id; vacía coincide con todo. */
export function coincide(q: string, t: Trabajo): boolean {
  const aguja = normalizar(q).trim()
  if (!aguja) return true
  return [t.titulo, t.repo, t.ancla.ref, t.id].some((campo) => normalizar(campo).includes(aguja))
}

/** Lo que el board y la lista dibujan con los filtros aplicados. */
export interface Vista {
  /** Los trabajos que pasan etapa y búsqueda, incluidos los sin traza. */
  visibles: Trabajo[]
  /** Los grupos por repo que se dibujan abiertos. */
  grupos: Grupo[]
  /** Solo con «todos»: los sin traza, agrupados igual, para el grupo colapsado. */
  sinTraza: Grupo[]
  nSinTraza: number
  /** false si la respuesta trae trabajos de un solo repo: no hacen falta encabezados. */
  encabezados: boolean
}

/** Aplica etapa y búsqueda sobre la respuesta y, con «todos», aparta los sin traza. */
export function armarVista(base: readonly Trabajo[], f: Pick<Filters, 'filtro' | 'etapas' | 'q'>): Vista {
  const visibles = base.filter((t) => (f.etapas.length === 0 || f.etapas.includes(t.etapa)) && coincide(f.q, t))
  const sinTraza = f.filtro === 'todos' ? visibles.filter((t) => t.sinHistoria) : []
  const abiertos = f.filtro === 'todos' ? visibles.filter((t) => !t.sinHistoria) : visibles
  return {
    visibles,
    grupos: agrupar(abiertos),
    sinTraza: agrupar(sinTraza),
    nSinTraza: sinTraza.length,
    encabezados: new Set(base.map((t) => t.repo)).size > 1,
  }
}
