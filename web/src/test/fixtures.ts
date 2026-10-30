// Datos de prueba con la forma de /api/workspace. Las referencias usan números
// de dos cifras: el gate de hex de CA-26 contaría «#» seguido de tres.
import type { Categoria, Detalle, Nodo, Trabajo, Workspace } from '@/lib/api'

const refs = { issue: null, grill: null, spec: null, pr: null }

function nodo(etapa: Nodo['etapa'], subEstado: string, extra: Partial<Nodo> = {}): Nodo {
  return { etapa, subEstado, ts: '2026-10-27T10:00:00Z', refs: { ...refs, issue: '#57' }, url: null, ...extra }
}

export const stack: Trabajo = {
  id: 'o/a/issue/57',
  repo: 'o/a',
  titulo: 'Sombras suaves',
  ancla: { tipo: 'issue', ref: '#57' },
  etapa: 'mergeado',
  subEstado: '2/3',
  ultimaActividad: '2026-10-27T11:00:00Z',
  terminado: false,
  sinHistoria: false,
  categoria: 'en-curso',
  ramas: [
    {
      id: 'principal',
      estado: 'viva',
      nodos: [
        nodo('grill', 'finalizado', { refs: { ...refs, issue: '#57', grill: '2026-10-01-sombras' } }),
        nodo('spec', 'implementada'),
        nodo('run', 'con PR'),
        ...[21, 22, 23].map((n, i) => nodo('pr', i < 2 ? 'mergeado' : 'validado', { refs: { ...refs, pr: n }, capa: `${i + 1}/3`, url: `https://github.com/o/a/pull/${n}` })),
        ...[21, 22].map((n, i) => nodo('mergeado', 'mergeado', { refs: { ...refs, pr: n }, capa: `${i + 1}/3` })),
      ],
    },
    { id: 'run-1', estado: 'fallada', nodos: [nodo('run', 'fallado (interrupted)')] },
  ],
}

export const enPR: Trabajo = {
  id: 'o/a/issue/58',
  repo: 'o/a',
  titulo: 'Exportar tablero',
  ancla: { tipo: 'issue', ref: '#58' },
  etapa: 'pr',
  subEstado: 'checks rojos',
  ultimaActividad: '2026-10-27T09:00:00Z',
  terminado: false,
  sinHistoria: false,
  categoria: 'en-curso',
  ramas: [{ id: 'principal', estado: 'viva', nodos: [nodo('spec', 'aprobada'), nodo('pr', 'checks rojos', { refs: { ...refs, pr: 31 }, capa: '1/1' })] }],
}

export const idea: Trabajo = {
  id: 'o/b/issue/12',
  repo: 'o/b',
  titulo: 'Idea sin artefactos',
  ancla: { tipo: 'issue', ref: '#12' },
  etapa: 'idea',
  subEstado: 'abierto',
  ultimaActividad: '2026-10-26T08:00:00Z',
  terminado: false,
  sinHistoria: false,
  categoria: 'idea',
  ramas: [{ id: 'principal', estado: 'viva', nodos: [nodo('idea', 'abierto', { stale: true })] }],
}

// Grill viejo sin ningún evento: el visor lo marca «sin traza».
export const sinTraza: Trabajo = {
  id: 'o/b/grill/2026-09-01-viejo',
  repo: 'o/b',
  titulo: 'Grill viejo',
  ancla: { tipo: 'grill', ref: '2026-09-01-viejo' },
  etapa: 'grill',
  subEstado: 'pausado',
  ultimaActividad: '2026-09-01T10:00:00Z',
  terminado: false,
  sinHistoria: true,
  categoria: 'sin-traza',
  ramas: [{ id: 'principal', estado: 'viva', nodos: [nodo('grill', 'pausado', { refs: { ...refs, grill: '2026-09-01-viejo' }, ts: '2026-09-01T10:00:00Z' })] }],
}

// Grill con traza en events.jsonl, todavía abierto: categoría en-curso.
export const grillVivo: Trabajo = { ...sinTraza, id: 'o/b/grill/2026-10-20-norte', titulo: 'Visor del flujo', ancla: { tipo: 'grill', ref: '2026-10-20-norte' }, sinHistoria: false, categoria: 'en-curso' }

// Mergeado con su issue cerrado: categoría terminado.
export const cerrado: Trabajo = {
  id: 'o/a/issue/40',
  repo: 'o/a',
  titulo: 'Modo oscuro',
  ancla: { tipo: 'issue', ref: '#40' },
  etapa: 'mergeado',
  subEstado: '1/1',
  ultimaActividad: '2026-10-20T10:00:00Z',
  terminado: true,
  sinHistoria: false,
  categoria: 'terminado',
  ramas: [
    {
      id: 'principal',
      estado: 'viva',
      nodos: [nodo('spec', 'implementada'), nodo('pr', 'mergeado', { refs: { ...refs, pr: 12 }, capa: '1/1' }), nodo('mergeado', 'mergeado', { refs: { ...refs, pr: 12 }, capa: '1/1' })],
    },
  ],
}

export const specSuelta: Trabajo = {
  id: 'o/a/spec/.sdd%2Fspecs%2Fsuelta.md',
  repo: 'o/a',
  titulo: 'Spec suelta',
  ancla: { tipo: 'spec', ref: '.sdd/specs/suelta.md' },
  etapa: 'spec',
  subEstado: 'aprobada',
  ultimaActividad: '2026-10-25T10:00:00Z',
  terminado: false,
  sinHistoria: false,
  categoria: 'en-curso',
  ramas: [{ id: 'principal', estado: 'viva', nodos: [nodo('spec', 'aprobada', { refs: { ...refs, spec: '.sdd/specs/suelta.md' } })] }],
}

const github = { ok: true, reason: null, fetchedAt: '2026-10-27T11:58:00Z' }

function repo(id: string) {
  return { id, nombre: id.split('/')[1] ?? id, rutas: [`/ws/${id}`], sinGit: false, remote: `https://github.com/${id}.git`, lineasInvalidas: 0, errores: [], github }
}

/** Los conteos que el server manda: trabajos por categoría, siempre las cuatro claves. */
export function contarCategorias(trabajos: readonly Trabajo[]): Record<Categoria, number> {
  const n = (c: Categoria) => trabajos.filter((t) => t.categoria === c).length
  return { 'en-curso': n('en-curso'), idea: n('idea'), terminado: n('terminado'), 'sin-traza': n('sin-traza') }
}

export function workspace(over: Partial<Workspace> = {}): Workspace {
  const trabajos = over.trabajos ?? [stack, enPR, idea]
  return {
    root: '/ws',
    scannedAt: '2026-10-27T11:59:00.000Z',
    scanning: false,
    github,
    repos: [repo('o/a'), repo('o/b')],
    trabajos,
    categorias: contarCategorias(trabajos),
    ...over,
  }
}

export const detalleStack: Detalle = {
  ...stack,
  gantt: [
    { rama: 'principal', etapa: 'run', inicio: '2026-10-17T10:00:00Z', fin: '2026-10-17T12:00:00Z', estado: 'con PR' },
    { rama: 'principal', etapa: 'pr', inicio: '2026-10-17T12:00:00Z', fin: '2026-10-20T10:00:00Z', estado: 'validado' },
    { rama: 'principal', etapa: 'mergeado', inicio: '2026-10-20T10:00:00Z', fin: null, estado: '2/3' },
    { rama: 'run-1', etapa: 'run', inicio: '2026-10-17T04:00:00Z', fin: '2026-10-17T06:00:00Z', estado: 'fallado (interrupted)' },
  ],
  links: {
    issue: 'https://github.com/o/a/issues/57',
    prs: [21, 22, 23].map((n) => ({ numero: n, url: `https://github.com/o/a/pull/${n}` })),
    grills: ['/ws/o/a/.sdd/grills/2026-10-01-sombras.md'],
    specs: ['/ws/o/a/.sdd/specs/sombras.md'],
  },
}

export const detalleSinTraza: Detalle = {
  ...sinTraza,
  gantt: [],
  links: { issue: null, prs: [], grills: ['/ws/o/b/.git/sdd/grills/2026-09-01-viejo.md'], specs: [] },
}
