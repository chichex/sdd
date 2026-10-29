import { fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import App from './App'
import type { Trabajo } from './lib/api'
import { mockApi, mockEventSource } from './test/api-mock'
import { enPR, idea, sinTraza, specSuelta, workspace } from './test/fixtures'

const ideaA: Trabajo = { ...idea, id: 'o/a/issue/70', repo: 'o/a', titulo: 'Roadmap del orden de grills', ancla: { tipo: 'issue', ref: '#70' }, ultimaActividad: '2026-10-26T10:00:00Z' }
const ideaB: Trabajo = { ...idea, id: 'o/a/issue/71', repo: 'o/a', titulo: 'Unificar el parser de glb', ancla: { tipo: 'issue', ref: '#71' }, ultimaActividad: '2026-10-25T10:00:00Z' }
const grillVivo: Trabajo = { ...sinTraza, id: 'o/b/grill/2026-10-20-norte', titulo: 'Visor del flujo', ancla: { tipo: 'grill', ref: '2026-10-20-norte' }, sinHistoria: false }

async function lista(trabajos: Trabajo[], path = '/') {
  window.history.replaceState(null, '', path)
  vi.stubGlobal('matchMedia', (q: string) => ({ matches: false, media: q, addEventListener: () => {}, removeEventListener: () => {} }))
  mockApi(workspace({ trabajos }))
  mockEventSource()
  render(<App />)
  return screen.findByRole('list', { name: 'Trabajos' })
}

const nombres = (el: HTMLElement) => within(el).getAllByRole('heading', { level: 4 }).map((h) => h.textContent)

describe('lista de celular por etapas', () => {
  it('cada repo parte sus filas en secciones con un h4 «<Etapa> <conteo>», y cada fila va bajo su encabezado', async () => {
    const l = await lista([enPR, ideaA, ideaB, idea])
    expect(nombres(l)).toEqual(['PR 1', 'Idea 2', 'Idea 1'])
    const pr = within(l).getByRole('heading', { level: 4, name: 'PR 1' }).closest('li') as HTMLElement
    expect(within(pr).getByRole('button', { name: /Exportar tablero/ })).toBeTruthy()
    expect(within(pr).queryByRole('button', { name: /Roadmap/ })).toBeNull()
    const ideas = within(l).getByRole('heading', { level: 4, name: 'Idea 2' }).closest('li') as HTMLElement
    expect(within(ideas).getAllByRole('button').map((b) => b.textContent)).toEqual([
      expect.stringContaining('Roadmap del orden de grills'),
      expect.stringContaining('Unificar el parser de glb'),
    ])
  })

  it('con una sola etapa en el repo también dibuja su encabezado', async () => {
    const l = await lista([idea])
    expect(nombres(l)).toEqual(['Idea 1'])
  })

  it('con el filtro de etapas solo aparecen las secciones de las etapas presentes', async () => {
    const l = await lista([enPR, ideaA, ideaB, idea], '/?etapa=idea')
    expect(nombres(l)).toEqual(['Idea 2', 'Idea 1'])
    expect(within(l).queryByRole('heading', { level: 4, name: 'PR 1' })).toBeNull()
  })

  it('el conteo de la sección sigue a la búsqueda', async () => {
    const l = await lista([enPR, ideaA, ideaB, idea], '/?q=roadmap')
    expect(nombres(l)).toEqual(['Idea 1'])
  })

  it('con «Todos», el grupo «Sin traza (N)» abierto también trae sus secciones por etapa', async () => {
    const l = await lista([enPR, ideaA, sinTraza], '/?filtro=todos')
    expect(nombres(l)).toEqual(['PR 1', 'Idea 1'])
    const grupo = within(l).getByRole('button', { name: 'Sin traza (1)' })
    fireEvent.click(grupo)
    const dentro = document.getElementById(grupo.getAttribute('aria-controls') ?? '') as HTMLElement
    expect(nombres(dentro)).toEqual(['Grill 1'])
    expect(within(dentro).getByRole('button', { name: /Grill viejo/ })).toBeTruthy()
  })

  it('el board de desktop no suma encabezados h4', async () => {
    window.history.replaceState(null, '', '/')
    mockApi(workspace({ trabajos: [enPR, ideaA, ideaB, idea] }))
    mockEventSource()
    render(<App />)
    const board = await screen.findByRole('region', { name: 'Tablero' })
    expect(within(board).queryAllByRole('heading', { level: 4 })).toHaveLength(0)
  })

  it('los encabezados de etapa no son controles: solo las filas entran al recorrido con Tab', async () => {
    const l = await lista([enPR, ideaA, ideaB, idea])
    const tabulables = l.querySelectorAll('button, a[href], input, select, textarea, [tabindex]:not([tabindex="-1"])')
    expect(tabulables).toHaveLength(4)
    for (const h of within(l).getAllByRole('heading', { level: 4 })) expect(h.tabIndex).toBe(-1)
  })
})

describe('fila de la lista con la etiqueta de ancla', () => {
  it.each([
    ['issue', enPR, /^issue Exportar tablero/],
    ['grill', grillVivo, /^grill Visor del flujo/],
    ['spec', specSuelta, /^spec Spec suelta/],
  ])('la palabra «%s» es texto visible y abre el nombre accesible de la fila', async (tipo, trabajo, nombre) => {
    const l = await lista([trabajo])
    const fila = within(l).getByRole('button', { name: nombre })
    expect(within(fila).getByText(tipo)).toBeTruthy()
  })

  it('la fila no lleva punto de etapa ni el ícono con tooltip «ancla: <tipo>»', async () => {
    const l = await lista([enPR, grillVivo])
    expect(within(l).queryAllByLabelText(/^ancla:/)).toHaveLength(0)
    // El punto de etapa es aria-hidden: su único rastro observable es el color de etapa.
    for (const fila of within(l).getAllByRole('button')) expect(fila.querySelector('[class*="bg-stage-"]')).toBeNull()
  })
})
