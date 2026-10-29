// Contrato JSON de la API del visor (/api/*) y su validación en la frontera:
// lo que llega por red se comprueba antes de usarlo.

export type Etapa = 'idea' | 'grill' | 'spec' | 'run' | 'pr' | 'mergeado' | 'publicado'
export const ETAPAS: readonly Etapa[] = ['idea', 'grill', 'spec', 'run', 'pr', 'mergeado', 'publicado']

/** Categoría de un trabajo: exactamente una; el control «Mostrar» elige cuáles se piden. */
export type Categoria = 'en-curso' | 'idea' | 'terminado' | 'sin-traza'
/** Orden canónico de las categorías: el de las chips, el del Sidebar y el de la URL. */
export const CATEGORIAS: readonly Categoria[] = ['en-curso', 'idea', 'terminado', 'sin-traza']
/** Lo que se muestra sin elegir: también es lo que el server devuelve sin ?mostrar=. */
export const MOSTRAR_DEFAULT: readonly Categoria[] = ['en-curso', 'idea']

/** Las categorías conocidas de la lista, en el orden canónico y sin repetidos. */
export function categoriasEnOrden(ids: readonly string[]): Categoria[] {
  return CATEGORIAS.filter((c) => ids.includes(c))
}

/** true si la selección es el default, sin importar el orden. */
export function esMostrarDefault(mostrar: readonly Categoria[]): boolean {
  const orden = categoriasEnOrden(mostrar)
  return orden.length === MOSTRAR_DEFAULT.length && orden.every((c, i) => c === MOSTRAR_DEFAULT[i])
}

export interface Refs {
  issue: string | null
  grill: string | null
  spec: string | null
  pr: number | null
  mergedAt?: string
}
export interface Nodo {
  etapa: Etapa
  subEstado: string
  ts: string
  refs: Refs
  url: string | null
  ruta?: string
  capa?: string
  stale?: boolean
  tags?: string[]
}
export interface Rama {
  id: string
  estado: string
  nodos: Nodo[]
}
export interface Trabajo {
  id: string
  repo: string
  titulo: string
  ancla: { tipo: string; ref: string }
  etapa: Etapa
  subEstado: string
  ultimaActividad: string
  terminado: boolean
  sinHistoria: boolean
  categoria: Categoria
  ramas: Rama[]
}
export interface GitHubEstado {
  ok: boolean
  reason: string | null
  fetchedAt: string | null
}
export interface Repo {
  id: string
  nombre: string
  rutas: string[]
  sinGit: boolean
  remote: string | null
  lineasInvalidas: number
  errores: string[]
  github: GitHubEstado
}
export interface Workspace {
  root: string
  scannedAt: string | null
  scanning: boolean
  github: GitHubEstado
  repos: Repo[]
  trabajos: Trabajo[]
  /** Trabajos por categoría tras aplicar el repo y antes de «Mostrar»; siempre las cuatro. */
  categorias: Record<Categoria, number>
}

export interface Barra {
  rama: string
  etapa: Etapa
  inicio: string
  fin: string | null
  estado: string
}
export interface Links {
  issue: string | null
  prs: Array<{ numero: number; url: string | null }>
  grills: string[]
  specs: string[]
}
export interface Detalle extends Trabajo {
  gantt: Barra[]
  links: Links
}

export class ApiError extends Error {
  status: number | null
  constructor(message: string, status: number | null = null) {
    super(message)
    this.status = status
  }
}

type Obj = Record<string, unknown>

function obj(v: unknown, where: string): Obj {
  if (v === null || typeof v !== 'object' || Array.isArray(v)) throw new ApiError(`${where}: se esperaba un objeto`)
  return v as Obj
}
function str(v: unknown, where: string): string {
  if (typeof v !== 'string') throw new ApiError(`${where}: se esperaba texto`)
  return v
}
function optStr(v: unknown, where: string): string | null {
  return v === null || v === undefined ? null : str(v, where)
}
function bool(v: unknown, where: string): boolean {
  if (typeof v !== 'boolean') throw new ApiError(`${where}: se esperaba booleano`)
  return v
}
function arr<T>(v: unknown, where: string, item: (x: unknown, i: number) => T): T[] {
  if (!Array.isArray(v)) throw new ApiError(`${where}: se esperaba una lista`)
  return v.map(item)
}
function etapa(v: unknown, where: string): Etapa {
  const s = str(v, where)
  if (!(ETAPAS as readonly string[]).includes(s)) throw new ApiError(`${where}: etapa desconocida ${s}`)
  return s as Etapa
}

function categoria(v: unknown, where: string): Categoria {
  const s = str(v, where)
  if (!(CATEGORIAS as readonly string[]).includes(s)) throw new ApiError(`${where}: categoría desconocida ${s}`)
  return s as Categoria
}

function parseCategorias(v: unknown): Record<Categoria, number> {
  const o = obj(v, 'categorias')
  const conteo = (c: Categoria): number => {
    const n = o[c]
    if (typeof n !== 'number' || !Number.isInteger(n) || n < 0) throw new ApiError(`categorias.${c}: se esperaba un conteo`)
    return n
  }
  return { 'en-curso': conteo('en-curso'), idea: conteo('idea'), terminado: conteo('terminado'), 'sin-traza': conteo('sin-traza') }
}

function parseGitHub(v: unknown): GitHubEstado {
  const o = obj(v, 'github')
  return { ok: bool(o.ok, 'github.ok'), reason: optStr(o.reason, 'github.reason'), fetchedAt: optStr(o.fetchedAt, 'github.fetchedAt') }
}

function parseNodo(v: unknown): Nodo {
  const o = obj(v, 'nodo')
  const r = obj(o.refs, 'nodo.refs')
  const pr = r.pr === null || r.pr === undefined ? null : Number(r.pr)
  const nodo: Nodo = {
    etapa: etapa(o.etapa, 'nodo.etapa'),
    subEstado: str(o.subEstado, 'nodo.subEstado'),
    ts: str(o.ts, 'nodo.ts'),
    refs: { issue: optStr(r.issue, 'refs.issue'), grill: optStr(r.grill, 'refs.grill'), spec: optStr(r.spec, 'refs.spec'), pr },
    url: optStr(o.url, 'nodo.url'),
  }
  if (typeof r.mergedAt === 'string') nodo.refs.mergedAt = r.mergedAt
  if (typeof o.ruta === 'string') nodo.ruta = o.ruta
  if (typeof o.capa === 'string') nodo.capa = o.capa
  if (o.stale === true) nodo.stale = true
  if (Array.isArray(o.tags)) nodo.tags = arr(o.tags, 'nodo.tags', (t) => str(t, 'tag'))
  return nodo
}

export function parseTrabajo(v: unknown): Trabajo {
  const o = obj(v, 'trabajo')
  const ancla = obj(o.ancla, 'trabajo.ancla')
  return {
    id: str(o.id, 'trabajo.id'),
    repo: str(o.repo, 'trabajo.repo'),
    titulo: str(o.titulo, 'trabajo.titulo'),
    ancla: { tipo: str(ancla.tipo, 'ancla.tipo'), ref: str(ancla.ref, 'ancla.ref') },
    etapa: etapa(o.etapa, 'trabajo.etapa'),
    subEstado: str(o.subEstado, 'trabajo.subEstado'),
    ultimaActividad: str(o.ultimaActividad, 'trabajo.ultimaActividad'),
    terminado: bool(o.terminado, 'trabajo.terminado'),
    sinHistoria: bool(o.sinHistoria, 'trabajo.sinHistoria'),
    categoria: categoria(o.categoria, 'trabajo.categoria'),
    ramas: arr(o.ramas, 'trabajo.ramas', (r) => {
      const ro = obj(r, 'rama')
      return { id: str(ro.id, 'rama.id'), estado: str(ro.estado, 'rama.estado'), nodos: arr(ro.nodos, 'rama.nodos', parseNodo) }
    }),
  }
}

export function parseWorkspace(v: unknown): Workspace {
  const o = obj(v, 'workspace')
  return {
    root: str(o.root, 'root'),
    scannedAt: optStr(o.scannedAt, 'scannedAt'),
    scanning: bool(o.scanning, 'scanning'),
    github: parseGitHub(o.github),
    repos: arr(o.repos, 'repos', (r) => {
      const ro = obj(r, 'repo')
      return {
        id: str(ro.id, 'repo.id'),
        nombre: str(ro.nombre, 'repo.nombre'),
        rutas: arr(ro.rutas, 'repo.rutas', (x) => str(x, 'ruta')),
        sinGit: bool(ro.sinGit, 'repo.sinGit'),
        remote: optStr(ro.remote, 'repo.remote'),
        lineasInvalidas: Number(ro.lineasInvalidas ?? 0),
        errores: arr(ro.errores ?? [], 'repo.errores', (x) => str(x, 'error')),
        github: parseGitHub(ro.github ?? { ok: false, reason: null, fetchedAt: null }),
      }
    }),
    trabajos: arr(o.trabajos, 'trabajos', parseTrabajo),
    categorias: parseCategorias(o.categorias),
  }
}

async function getJSON(url: string, init?: RequestInit): Promise<unknown> {
  const res = await fetch(url, init)
  if (!res.ok) throw new ApiError(`${url}: HTTP ${res.status}`, res.status)
  return res.json() as Promise<unknown>
}

/** Pide los trabajos de las categorías elegidas; con el default no manda ?mostrar=. */
export async function fetchWorkspace(mostrar: readonly Categoria[]): Promise<Workspace> {
  const query = esMostrarDefault(mostrar) ? '' : `?mostrar=${categoriasEnOrden(mostrar).join(',')}`
  return parseWorkspace(await getJSON(`/api/workspace${query}`))
}

export async function postRefresh(): Promise<string | null> {
  const body = obj(await getJSON('/api/refresh', { method: 'POST' }), 'refresh')
  return optStr(body.fetchedAt, 'refresh.fetchedAt')
}

export async function fetchHealth(): Promise<{ version: string; root: string }> {
  const o = obj(await getJSON('/api/health'), 'health')
  return { version: str(o.version, 'health.version'), root: str(o.root, 'health.root') }
}

function parseDetalle(v: unknown): Detalle {
  const o = obj(v, 'detalle')
  const links = obj(o.links, 'links')
  return {
    ...parseTrabajo(o),
    gantt: arr(o.gantt, 'gantt', (b) => {
      const bo = obj(b, 'barra')
      return { rama: str(bo.rama, 'barra.rama'), etapa: etapa(bo.etapa, 'barra.etapa'), inicio: str(bo.inicio, 'barra.inicio'), fin: optStr(bo.fin, 'barra.fin'), estado: str(bo.estado, 'barra.estado') }
    }),
    links: {
      issue: optStr(links.issue, 'links.issue'),
      prs: arr(links.prs, 'links.prs', (p) => {
        const po = obj(p, 'pr')
        return { numero: Number(po.numero), url: optStr(po.url, 'pr.url') }
      }),
      grills: arr(links.grills, 'links.grills', (x) => str(x, 'grill')),
      specs: arr(links.specs, 'links.specs', (x) => str(x, 'spec')),
    },
  }
}

/** Detalle de un trabajo; un id desconocido rechaza con ApiError de status 404. */
export async function fetchTrabajo(id: string): Promise<Detalle> {
  return parseDetalle(await getJSON(`/api/trabajos/${encodeURIComponent(id)}`))
}
