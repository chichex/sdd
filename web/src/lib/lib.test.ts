import { describe, expect, it, vi } from 'vitest'
import { edadCorta } from './time'
import { filtersQuery, readFilters, trabajoPath, writeFilters } from './router'
import { aplicarTema, guardarTema, leerTema } from './theme'

describe('edadCorta', () => {
  const now = new Date('2026-10-27T12:00:00Z')
  it.each([
    ['2026-10-27T11:59:30Z', 'ahora'],
    ['2026-10-27T11:40:00Z', '20 min'],
    ['2026-10-27T09:00:00Z', '3 h'],
    ['2026-10-26T11:00:00Z', '1 d'],
    ['2026-10-27T12:05:00Z', 'ahora'],
  ])('%s → %s', (ts, want) => {
    expect(edadCorta(ts, now)).toBe(want)
  })
  it('sin fecha devuelve un guion', () => {
    expect(edadCorta(null, now)).toBe('—')
  })
})

describe('filtros en la URL', () => {
  const DEFAULT = ['en-curso', 'idea']
  it('lee repo y mostrar, con en-curso e idea por default', () => {
    expect(readFilters('?repo=chichex%2Fwaica&mostrar=terminado')).toEqual({ repo: 'chichex/waica', mostrar: ['terminado'], etapas: [], q: '' })
    expect(readFilters('')).toEqual({ repo: null, mostrar: DEFAULT, etapas: [], q: '' })
  })
  it('lee mostrar en el orden canónico, sin repetidos e ignorando ids desconocidos', () => {
    expect(readFilters('?mostrar=terminado,en-curso,zzz').mostrar).toEqual(['en-curso', 'terminado'])
    expect(readFilters('?mostrar=sin-traza,idea,idea').mostrar).toEqual(['idea', 'sin-traza'])
    expect(readFilters('?mostrar=sin-traza,terminado,idea,en-curso').mostrar).toEqual(['en-curso', 'idea', 'terminado', 'sin-traza'])
  })
  it.each(['?mostrar=zzz', '?mostrar=', '?mostrar=,', '?otro=1'])('%s cae al default', (search) => {
    expect(readFilters(search).mostrar).toEqual(DEFAULT)
  })
  it('?filtro= de un link viejo se ignora', () => {
    expect(readFilters('?filtro=con-traza').mostrar).toEqual(DEFAULT)
    expect(readFilters('?filtro=todos&mostrar=terminado').mostrar).toEqual(['terminado'])
  })
  it('lee etapa en el orden de las etapas, sin repetidos e ignorando ids desconocidos', () => {
    expect(readFilters('?etapa=spec,grill,zzz').etapas).toEqual(['grill', 'spec'])
    expect(readFilters('?etapa=pr,pr,idea').etapas).toEqual(['idea', 'pr'])
    expect(readFilters('?etapa=').etapas).toEqual([])
  })
  it('lee q tal cual', () => {
    expect(readFilters('?q=sombras%20suaves').q).toBe('sombras suaves')
  })
  it('escribe los cuatro filtros en orden, con comas literales y etapa en el orden de las etapas', () => {
    expect(filtersQuery({ repo: 'o/a', mostrar: ['en-curso', 'terminado'], etapas: ['spec', 'grill'], q: 'nandu' })).toBe(
      '?repo=o%2Fa&mostrar=en-curso,terminado&etapa=grill,spec&q=nandu',
    )
  })
  it('omite el default de mostrar, etapa y q vacíos', () => {
    expect(filtersQuery({ repo: null, mostrar: ['en-curso', 'idea'], etapas: [], q: '' })).toBe('')
    expect(filtersQuery({ repo: null, mostrar: ['idea', 'en-curso'], etapas: [], q: '' })).toBe('')
    expect(filtersQuery({ repo: null, mostrar: ['idea', 'terminado', 'en-curso'], etapas: [], q: '' })).toBe('?mostrar=en-curso,idea,terminado')
  })
  it('trabajoPath conserva mostrar', () => {
    expect(trabajoPath('o/a/issue/57', { repo: null, mostrar: ['terminado'], etapas: [], q: '' })).toBe('/trabajos/o%2Fa%2Fissue%2F57?mostrar=terminado')
  })
  it('escribe con replaceState sin sumar historia y omite los defaults', () => {
    window.history.replaceState(null, '', '/?x=1')
    const before = window.history.length
    writeFilters({ repo: 'chichex/waica', mostrar: ['en-curso'], etapas: [], q: '' })
    expect(window.location.search).toBe('?repo=chichex%2Fwaica&mostrar=en-curso')
    expect(window.history.length).toBe(before)
    writeFilters({ repo: null, mostrar: ['en-curso', 'idea'], etapas: [], q: '' })
    expect(window.location.search).toBe('')
  })
})

describe('tema', () => {
  it('oscuro y claro fijan data-theme; sistema lo quita', () => {
    aplicarTema('oscuro')
    expect(document.documentElement.dataset.theme).toBe('dark')
    aplicarTema('claro')
    expect(document.documentElement.dataset.theme).toBe('light')
    aplicarTema('sistema')
    expect(document.documentElement.hasAttribute('data-theme')).toBe(false)
  })
  it('persiste en localStorage y vuelve a sistema sin dato', () => {
    expect(leerTema()).toBe('sistema')
    guardarTema('oscuro')
    expect(window.localStorage.getItem('sdd-tema')).toBe('oscuro')
    expect(leerTema()).toBe('oscuro')
  })
  it('sin storage no rompe', () => {
    const spy = vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('bloqueado')
    })
    expect(leerTema()).toBe('sistema')
    spy.mockRestore()
  })
})
