// Filtrado y agrupado del lado del cliente sobre la respuesta de
// /api/workspace: búsqueda por texto, etapas elegidas y grupos por repo.
import { ETAPAS, type Trabajo } from './api'

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
