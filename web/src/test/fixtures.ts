// Datos de prueba con la forma de /api/workspace. Las referencias usan números
// de dos cifras: el gate de hex de CA-26 contaría «#» seguido de tres.
import type { Nodo, Trabajo, Workspace } from '@/lib/api'

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
  ramas: [{ id: 'principal', estado: 'viva', nodos: [nodo('idea', 'abierto', { stale: true })] }],
}

const github = { ok: true, reason: null, fetchedAt: '2026-10-27T11:58:00Z' }

function repo(id: string) {
  return { id, nombre: id.split('/')[1] ?? id, rutas: [`/ws/${id}`], sinGit: false, remote: `https://github.com/${id}.git`, lineasInvalidas: 0, errores: [], github }
}

export function workspace(over: Partial<Workspace> = {}): Workspace {
  return {
    root: '/ws',
    scannedAt: '2026-10-27T11:59:00.000Z',
    scanning: false,
    github,
    repos: [repo('o/a'), repo('o/b')],
    trabajos: [stack, enPR, idea],
    ...over,
  }
}
