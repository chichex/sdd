import { describe, expect, it, vi } from 'vitest'
import { edadCorta } from './time'
import { filtersQuery, readFilters, writeFilters } from './router'
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
  it('lee repo y filtro, con activos por default', () => {
    expect(readFilters('?repo=chichex%2Fwaica&filtro=todos')).toEqual({ repo: 'chichex/waica', filtro: 'todos', etapas: [], q: '' })
    expect(readFilters('')).toEqual({ repo: null, filtro: 'activos', etapas: [], q: '' })
    expect(readFilters('?filtro=raro')).toEqual({ repo: null, filtro: 'activos', etapas: [], q: '' })
  })
  it('acepta filtro=con-traza', () => {
    expect(readFilters('?filtro=con-traza').filtro).toBe('con-traza')
  })
  it('lee etapa en el orden de las etapas, sin repetidos e ignorando ids desconocidos', () => {
    expect(readFilters('?etapa=spec,grill,zzz').etapas).toEqual(['grill', 'spec'])
    expect(readFilters('?etapa=pr,pr,idea').etapas).toEqual(['idea', 'pr'])
    expect(readFilters('?etapa=').etapas).toEqual([])
  })
  it('lee q tal cual', () => {
    expect(readFilters('?q=sombras%20suaves').q).toBe('sombras suaves')
  })
  it('escribe los cuatro filtros en orden, etapa con comas y en el orden de las etapas', () => {
    expect(filtersQuery({ repo: 'o/a', filtro: 'con-traza', etapas: ['spec', 'grill'], q: 'nandu' })).toBe('?repo=o%2Fa&filtro=con-traza&etapa=grill,spec&q=nandu')
  })
  it('omite etapa y q vacíos', () => {
    expect(filtersQuery({ repo: null, filtro: 'activos', etapas: [], q: '' })).toBe('')
    expect(filtersQuery({ repo: null, filtro: 'todos', etapas: [], q: '' })).toBe('?filtro=todos')
  })
  it('escribe con replaceState sin sumar historia y omite los defaults', () => {
    window.history.replaceState(null, '', '/?x=1')
    const before = window.history.length
    writeFilters({ repo: 'chichex/waica', filtro: 'todos', etapas: [], q: '' })
    expect(window.location.search).toBe('?repo=chichex%2Fwaica&filtro=todos')
    expect(window.history.length).toBe(before)
    writeFilters({ repo: null, filtro: 'activos', etapas: [], q: '' })
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
