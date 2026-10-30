// Ruteo mínimo del visor (sin librería): la ruta sale de location y los
// filtros viven en la query (?repo=&mostrar=&etapa=&q=) para que una recarga
// o un link compartido los conserven.
import { categoriasEnOrden, esMostrarDefault, ETAPAS, MOSTRAR_DEFAULT, type Categoria, type Etapa } from './api'

export interface Filters {
  repo: string | null
  /** Categorías elegidas, sin repetidos y en el orden canónico; nunca vacío (default en-curso e ideas). */
  mostrar: Categoria[]
  /** Etapas elegidas, sin repetidos y en el orden del flujo; vacío = todas. */
  etapas: Etapa[]
  /** Búsqueda tal como se escribió; vacío = sin búsqueda. */
  q: string
}

/** Etapas conocidas de la lista, en el orden del flujo y sin repetidos. */
function etapasEnOrden(ids: readonly string[]): Etapa[] {
  return ETAPAS.filter((e) => ids.includes(e))
}

export function readFilters(search: string): Filters {
  const params = new URLSearchParams(search)
  const mostrar = categoriasEnOrden((params.get('mostrar') ?? '').split(','))
  return {
    repo: params.get('repo') || null,
    mostrar: mostrar.length > 0 ? mostrar : [...MOSTRAR_DEFAULT],
    etapas: etapasEnOrden((params.get('etapa') ?? '').split(',')),
    q: params.get('q') ?? '',
  }
}

export function filtersQuery(f: Filters): string {
  const params = new URLSearchParams()
  if (f.repo) params.set('repo', f.repo)
  if (f.mostrar.length > 0 && !esMostrarDefault(f.mostrar)) params.set('mostrar', categoriasEnOrden(f.mostrar).join(','))
  if (f.etapas.length > 0) params.set('etapa', etapasEnOrden(f.etapas).join(','))
  if (f.q) params.set('q', f.q)
  // La coma es válida en una query: se deja literal para que ?etapa=grill,spec se lea.
  const query = params.toString().replaceAll('%2C', ',')
  return query ? `?${query}` : ''
}

export const NAVIGATE_EVENT = 'sdd:navigate'

/** Escribe los filtros en la URL actual sin sumar una entrada de historia. */
export function writeFilters(f: Filters): void {
  window.history.replaceState(window.history.state, '', `${window.location.pathname}${filtersQuery(f)}`)
  window.dispatchEvent(new Event(NAVIGATE_EVENT))
}

/** Navega a una ruta del front sumando una entrada de historia. */
export function navigate(pathAndQuery: string): void {
  window.history.pushState(null, '', pathAndQuery)
  window.dispatchEvent(new Event(NAVIGATE_EVENT))
}

/** Ruta del detalle de un trabajo; el id va codificado como un solo segmento. */
export function trabajoPath(id: string, f: Filters): string {
  return `/trabajos/${encodeURIComponent(id)}${filtersQuery(f)}`
}

/** Id del trabajo si la ruta es la del detalle. */
export function trabajoFromPath(pathname: string): string | null {
  const prefix = '/trabajos/'
  if (!pathname.startsWith(prefix) || pathname.length === prefix.length) return null
  try {
    return decodeURIComponent(pathname.slice(prefix.length))
  } catch {
    return null
  }
}
