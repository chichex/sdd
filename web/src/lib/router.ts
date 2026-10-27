// Ruteo mínimo del visor (sin librería): la ruta sale de location y los
// filtros viven en la query (?repo=&filtro=) para que una recarga o un link
// compartido los conserven.
export type Filtro = 'activos' | 'todos'

export interface Filters {
  repo: string | null
  filtro: Filtro
}

export function readFilters(search: string): Filters {
  const params = new URLSearchParams(search)
  const filtro = params.get('filtro') === 'todos' ? 'todos' : 'activos'
  return { repo: params.get('repo') || null, filtro }
}

export function filtersQuery(f: Filters): string {
  const params = new URLSearchParams()
  if (f.repo) params.set('repo', f.repo)
  if (f.filtro !== 'activos') params.set('filtro', f.filtro)
  const query = params.toString()
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
