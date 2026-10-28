import { describe, expect, it } from 'vitest'
import type { Etapa, Trabajo } from './api'
import { agrupar, armarVista, coincide } from './filtrar'

function trabajo(id: string, repo: string, etapa: Etapa, ultimaActividad: string, over: Partial<Trabajo> = {}): Trabajo {
  return {
    id,
    repo,
    titulo: id,
    ancla: { tipo: 'issue', ref: '#1' },
    etapa,
    subEstado: '',
    ultimaActividad,
    terminado: false,
    sinHistoria: false,
    ramas: [],
    ...over,
  }
}

describe('agrupar', () => {
  it('arma un grupo por repo, ordenados por id ascendente', () => {
    const grupos = agrupar([
      trabajo('p1', 'pramaestudio/platform', 'spec', '2026-10-27T10:00:00Z'),
      trabajo('w1', 'chichex/waica', 'grill', '2026-10-20T10:00:00Z'),
      trabajo('d1', 'chichex/dale', 'idea', '2026-10-01T10:00:00Z'),
    ])
    expect(grupos.map((g) => g.repo)).toEqual(['chichex/dale', 'chichex/waica', 'pramaestudio/platform'])
  })

  it('dentro del repo va primero la etapa más avanzada y, a igual etapa, la actividad más reciente', () => {
    const [grupo] = agrupar([
      trabajo('idea', 'o/a', 'idea', '2026-10-27T12:00:00Z'),
      trabajo('spec-vieja', 'o/a', 'spec', '2026-10-01T10:00:00Z'),
      trabajo('publicado', 'o/a', 'publicado', '2026-09-01T10:00:00Z'),
      trabajo('spec-nueva', 'o/a', 'spec', '2026-10-25T10:00:00Z'),
      trabajo('pr', 'o/a', 'pr', '2026-10-02T10:00:00Z'),
    ])
    expect(grupo?.trabajos.map((t) => t.id)).toEqual(['publicado', 'pr', 'spec-nueva', 'spec-vieja', 'idea'])
  })

  it('sin trabajos no hay grupos', () => {
    expect(agrupar([])).toEqual([])
  })
})

describe('coincide', () => {
  const t = trabajo('chichex/waica/issue/57', 'chichex/waica', 'pr', '2026-10-27T10:00:00Z', {
    titulo: 'Ñandú en 3D',
    ancla: { tipo: 'issue', ref: '#57' },
  })

  it.each([
    ['nandu', 'sin acentos ni eñe'],
    ['ÑANDÚ', 'sin distinguir mayúsculas'],
    ['en 3d', 'substring del medio del título'],
    ['#57', 'la referencia del ancla'],
    ['chichex/waica', 'el repo owner/repo'],
    ['issue/57', 'el id'],
    ['', 'búsqueda vacía'],
    ['   ', 'búsqueda de solo espacios'],
  ])('«%s» coincide (%s)', (q) => {
    expect(coincide(q, t)).toBe(true)
  })

  it('no coincide con un texto que no está en ningún campo', () => {
    expect(coincide('zzz', t)).toBe(false)
    expect(coincide('#58', t)).toBe(false)
  })
})

describe('armarVista', () => {
  const base = [
    trabajo('a-spec', 'o/a', 'spec', '2026-10-27T10:00:00Z', { titulo: 'Sombras' }),
    trabajo('a-viejo', 'o/a', 'grill', '2026-09-01T10:00:00Z', { titulo: 'Grill viejo', sinHistoria: true }),
    trabajo('b-idea', 'o/b', 'idea', '2026-10-26T10:00:00Z', { titulo: 'Idea' }),
  ]
  const sinFiltros = { etapas: [], q: '' }

  it('con «todos» separa los sin traza al final, agrupados por repo', () => {
    const v = armarVista(base, { filtro: 'todos', ...sinFiltros })
    expect(v.grupos.map((g) => [g.repo, g.trabajos.map((t) => t.id)])).toEqual([['o/a', ['a-spec']], ['o/b', ['b-idea']]])
    expect(v.sinTraza.map((g) => [g.repo, g.trabajos.map((t) => t.id)])).toEqual([['o/a', ['a-viejo']]])
    expect(v.nSinTraza).toBe(1)
  })

  it.each(['activos', 'con-traza'] as const)('con «%s» no hay grupo sin traza', (filtro) => {
    const v = armarVista(base, { filtro, ...sinFiltros })
    expect(v.sinTraza).toEqual([])
    expect(v.nSinTraza).toBe(0)
  })

  it('etapa y búsqueda filtran antes de separar: N cuenta lo que queda', () => {
    expect(armarVista(base, { filtro: 'todos', etapas: ['spec'], q: '' }).nSinTraza).toBe(0)
    expect(armarVista(base, { filtro: 'todos', etapas: [], q: 'viejo' }).nSinTraza).toBe(1)
    expect(armarVista(base, { filtro: 'todos', etapas: ['grill', 'idea'], q: '' }).visibles.map((t) => t.id)).toEqual(['a-viejo', 'b-idea'])
  })

  it('omite los encabezados si la base trae trabajos de un solo repo', () => {
    expect(armarVista(base, { filtro: 'todos', ...sinFiltros }).encabezados).toBe(true)
    expect(armarVista(base.slice(0, 2), { filtro: 'todos', ...sinFiltros }).encabezados).toBe(false)
  })
})
