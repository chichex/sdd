import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import App from './App'
import { mockApi, mockEventSource } from './test/api-mock'
import { detalleStack, workspace } from './test/fixtures'

function abrir(path: string) {
  window.history.replaceState(null, '', path)
  const api = mockApi(workspace())
  api.detalles.set(detalleStack.id, detalleStack)
  mockEventSource()
  render(<App />)
  return api
}

describe('detalle de un trabajo', () => {
  it('la ruta /trabajos/<id> abre el detalle sobre el board con título, chips y grafo', async () => {
    const api = abrir(`/trabajos/${encodeURIComponent(detalleStack.id)}?filtro=todos`)
    const dialog = await screen.findByRole('dialog', { name: 'Sombras suaves' })
    expect(api.calls.some((c) => c.url === `/api/trabajos/${encodeURIComponent(detalleStack.id)}`)).toBe(true)
    expect(within(dialog).getByText('o/a')).toBeTruthy()
    expect(within(dialog).getByRole('link', { name: '#57' }).getAttribute('href')).toBe('https://github.com/o/a/issues/57')
    const grafo = within(dialog).getByRole('img', { name: 'Grafo' })
    expect(grafo.querySelectorAll('[data-grafo-nodo]')).toHaveLength(9)
    expect(grafo.querySelectorAll('[data-activo]')).toHaveLength(1)
    expect(grafo.querySelectorAll('[data-rama="run-1"]')).toHaveLength(1)
    expect(within(dialog).getByRole('link', { name: 'PR 23' }).getAttribute('href')).toBe('https://github.com/o/a/pull/23')
    expect(within(dialog).getByText('/ws/o/a/.sdd/specs/sombras.md')).toBeTruthy()
    expect(screen.getByRole('region', { name: 'Tablero', hidden: true })).toBeTruthy() // debajo del modal, con aria-hidden
  })

  it('el Gantt dibuja una barra por etapa y rama con la línea de hoy', async () => {
    abrir(`/trabajos/${encodeURIComponent(detalleStack.id)}`)
    const dialog = await screen.findByRole('dialog', { name: 'Sombras suaves' })
    const gantt = within(dialog).getByRole('img', { name: 'Gantt' })
    expect(gantt.querySelectorAll('[data-barra]')).toHaveLength(4)
    expect(gantt.querySelectorAll('[data-hoy]')).toHaveLength(1)
  })

  it('Esc cierra y vuelve a / conservando los filtros', async () => {
    abrir(`/trabajos/${encodeURIComponent(detalleStack.id)}?repo=o%2Fa&filtro=todos`)
    const dialog = await screen.findByRole('dialog', { name: 'Sombras suaves' })
    fireEvent.keyDown(dialog, { key: 'Escape' })
    await waitFor(() => expect(window.location.pathname).toBe('/'))
    expect(window.location.search).toBe('?repo=o%2Fa&filtro=todos')
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())
  })

  it('un id inexistente muestra «Trabajo no encontrado»', async () => {
    abrir(`/trabajos/${encodeURIComponent('o/a/issue/99')}`)
    expect(await screen.findByText('Trabajo no encontrado')).toBeTruthy()
  })

  it('en celular el detalle abre como Sheet desde abajo', async () => {
    vi.stubGlobal('matchMedia', (q: string) => ({ matches: false, media: q, addEventListener: () => {}, removeEventListener: () => {} }))
    abrir(`/trabajos/${encodeURIComponent(detalleStack.id)}`)
    const sheet = await screen.findByRole('dialog', { name: 'Sombras suaves' })
    expect(sheet.dataset.slot).toBe('sheet-content')
  })

  it('tocar una fila de la lista abre el detalle', async () => {
    vi.stubGlobal('matchMedia', (q: string) => ({ matches: false, media: q, addEventListener: () => {}, removeEventListener: () => {} }))
    abrir('/')
    const lista = await screen.findByRole('list', { name: 'Trabajos' })
    fireEvent.click(within(lista).getByRole('button', { name: /Sombras suaves/ }))
    expect(await screen.findByRole('dialog', { name: 'Sombras suaves' })).toBeTruthy()
  })
})
