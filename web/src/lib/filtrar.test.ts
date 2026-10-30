import { describe, expect, it } from 'vitest'
import type { Categoria, Etapa, Trabajo } from './api'
import { agrupar, alternarCategoria, armarVista, coincide, seccionarPorEtapa } from './filtrar'

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
    categoria: 'en-curso',
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

describe('seccionarPorEtapa', () => {
  const pr1 = trabajo('pr1', 'o/a', 'pr', '2026-10-27T10:00:00Z')
  const pr2 = trabajo('pr2', 'o/a', 'pr', '2026-10-26T10:00:00Z')
  const idea = trabajo('idea', 'o/a', 'idea', '2026-10-25T10:00:00Z')

  it('parte en secciones consecutivas, una por etapa, sin reordenar ni perder trabajos', () => {
    expect(seccionarPorEtapa([pr1, pr2, idea])).toEqual([
      { etapa: 'pr', trabajos: [pr1, pr2] },
      { etapa: 'idea', trabajos: [idea] },
    ])
  })

  it('sin trabajos no hay secciones', () => {
    expect(seccionarPorEtapa([])).toEqual([])
  })

  it('con una sola etapa hay una sola sección', () => {
    expect(seccionarPorEtapa([pr1, pr2])).toEqual([{ etapa: 'pr', trabajos: [pr1, pr2] }])
  })

  it('sobre la salida de agrupar, las secciones siguen el orden del grupo', () => {
    const [grupo] = agrupar([idea, pr2, trabajo('spec', 'o/a', 'spec', '2026-10-24T10:00:00Z'), pr1])
    const secciones = seccionarPorEtapa(grupo?.trabajos ?? [])
    expect(secciones.map((s) => [s.etapa, s.trabajos.map((t) => t.id)])).toEqual([
      ['pr', ['pr1', 'pr2']],
      ['spec', ['spec']],
      ['idea', ['idea']],
    ])
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
    trabajo('a-viejo', 'o/a', 'grill', '2026-09-01T10:00:00Z', { titulo: 'Grill viejo', sinHistoria: true, categoria: 'sin-traza' }),
    trabajo('b-idea', 'o/b', 'idea', '2026-10-26T10:00:00Z', { titulo: 'Idea', categoria: 'idea' }),
  ]
  const sinFiltros = { etapas: [], q: '' }

  it('separa al final los de la categoría sin-traza, agrupados por repo, y deja las ideas en línea', () => {
    const v = armarVista(base, sinFiltros)
    expect(v.grupos.map((g) => [g.repo, g.trabajos.map((t) => t.id)])).toEqual([['o/a', ['a-spec']], ['o/b', ['b-idea']]])
    expect(v.sinTraza.map((g) => [g.repo, g.trabajos.map((t) => t.id)])).toEqual([['o/a', ['a-viejo']]])
    expect(v.nSinTraza).toBe(1)
  })

  it('sin trabajos sin-traza en la respuesta no hay grupo', () => {
    const v = armarVista(base.filter((t) => t.categoria !== 'sin-traza'), sinFiltros)
    expect(v.sinTraza).toEqual([])
    expect(v.nSinTraza).toBe(0)
  })

  it('el grupo depende de la categoría, no del flag sinHistoria', () => {
    const conFlag = [trabajo('x', 'o/a', 'spec', '2026-10-27T10:00:00Z', { sinHistoria: true, categoria: 'en-curso' })]
    expect(armarVista(conFlag, sinFiltros).nSinTraza).toBe(0)
    const sinFlag = [trabajo('y', 'o/a', 'spec', '2026-10-27T10:00:00Z', { sinHistoria: false, categoria: 'sin-traza' })]
    expect(armarVista(sinFlag, sinFiltros).nSinTraza).toBe(1)
  })

  it('etapa y búsqueda filtran antes de separar: N cuenta lo que queda', () => {
    expect(armarVista(base, { etapas: ['spec'], q: '' }).nSinTraza).toBe(0)
    expect(armarVista(base, { etapas: [], q: 'viejo' }).nSinTraza).toBe(1)
    expect(armarVista(base, { etapas: ['grill', 'idea'], q: '' }).visibles.map((t) => t.id)).toEqual(['a-viejo', 'b-idea'])
  })

  it('omite los encabezados si la base trae trabajos de un solo repo', () => {
    expect(armarVista(base, sinFiltros).encabezados).toBe(true)
    expect(armarVista(base.slice(0, 2), sinFiltros).encabezados).toBe(false)
  })
})

describe('alternarCategoria', () => {
  const t = (mostrar: Categoria[], c: Categoria) => alternarCategoria(mostrar, c)

  it('suma una categoría en el orden canónico', () => {
    expect(t(['en-curso'], 'sin-traza')).toEqual(['en-curso', 'sin-traza'])
    expect(t(['terminado'], 'en-curso')).toEqual(['en-curso', 'terminado'])
    expect(t(['en-curso', 'terminado', 'sin-traza'], 'idea')).toEqual(['en-curso', 'idea', 'terminado', 'sin-traza'])
  })

  it('saca una categoría de la selección', () => {
    expect(t(['en-curso', 'idea'], 'idea')).toEqual(['en-curso'])
    expect(t(['en-curso', 'idea', 'terminado'], 'en-curso')).toEqual(['idea', 'terminado'])
  })

  it('tocar la última activa no la destilda', () => {
    expect(t(['terminado'], 'terminado')).toEqual(['terminado'])
  })
})
