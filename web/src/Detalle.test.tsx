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

const celular = () =>
  vi.stubGlobal('matchMedia', (q: string) => ({ matches: false, media: q, addEventListener: () => {}, removeEventListener: () => {} }))

describe('detalle de un trabajo', () => {
  it('la ruta /trabajos/<id> abre el detalle sobre el board con título, chips y grafo', async () => {
    const api = abrir(`/trabajos/${encodeURIComponent(detalleStack.id)}?mostrar=en-curso,terminado`)
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
    abrir(`/trabajos/${encodeURIComponent(detalleStack.id)}?repo=o%2Fa&mostrar=en-curso,terminado`)
    const dialog = await screen.findByRole('dialog', { name: 'Sombras suaves' })
    fireEvent.keyDown(dialog, { key: 'Escape' })
    await waitFor(() => expect(window.location.pathname).toBe('/'))
    expect(window.location.search).toBe('?repo=o%2Fa&mostrar=en-curso,terminado')
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())
  })

  it('un id inexistente muestra «Trabajo no encontrado»', async () => {
    abrir(`/trabajos/${encodeURIComponent('o/a/issue/99')}`)
    expect(await screen.findByText('Trabajo no encontrado')).toBeTruthy()
  })

  it('en celular el detalle abre como Sheet desde abajo', async () => {
    celular()
    abrir(`/trabajos/${encodeURIComponent(detalleStack.id)}`)
    const sheet = await screen.findByRole('dialog', { name: 'Sombras suaves' })
    expect(sheet.dataset.slot).toBe('sheet-content')
  })

  it('tocar una fila de la lista abre el detalle', async () => {
    celular()
    abrir('/')
    const lista = await screen.findByRole('list', { name: 'Trabajos' })
    fireEvent.click(within(lista).getByRole('button', { name: /Sombras suaves/ }))
    expect(await screen.findByRole('dialog', { name: 'Sombras suaves' })).toBeTruthy()
  })
})

describe('grafo del detalle', () => {
  const filas = (grafo: Element) => [...grafo.querySelectorAll('[data-grafo-nodo]')].map((g) => [...g.querySelectorAll('text')].map((t) => t.textContent))

  it('en celular es vertical: ocupa el ancho del contenedor, sin scroll horizontal', async () => {
    celular()
    abrir(`/trabajos/${encodeURIComponent(detalleStack.id)}`)
    const sheet = await screen.findByRole('dialog', { name: 'Sombras suaves' })
    const grafo = within(sheet).getByRole('img', { name: 'Grafo' })
    expect(grafo.getAttribute('width')).toBe('100%')
    expect(grafo.getAttribute('viewBox')).toMatch(/^0 0 340 \d+$/)
    expect(grafo.parentElement?.className).not.toContain('overflow-x-auto')
    expect(grafo.querySelectorAll('[data-grafo-nodo]')).toHaveLength(9)
    expect(grafo.querySelectorAll('[data-activo]')).toHaveLength(1)
    expect(grafo.querySelectorAll('[data-rama="run-1"]')).toHaveLength(1)
  })

  it('en celular cada nodo muestra etiqueta, sub-estado y fecha como texto, de arriba hacia abajo', async () => {
    celular()
    abrir(`/trabajos/${encodeURIComponent(detalleStack.id)}`)
    const grafo = within(await screen.findByRole('dialog', { name: 'Sombras suaves' })).getByRole('img', { name: 'Grafo' })
    const textos = filas(grafo)
    expect(textos.map((t) => t.slice(0, 2))).toEqual([
      ['Grill', 'finalizado'],
      ['Spec', 'implementada'],
      ['Run', 'con PR'],
      ['PR 1/3', 'mergeado'],
      ['PR 2/3', 'mergeado'],
      ['PR 3/3', 'validado'],
      ['Mergeado 1/3', 'mergeado'],
      ['Mergeado 2/3', 'mergeado'],
      ['Run', 'fallado (interrupted)'],
    ])
    expect(textos.every((t) => t.length === 3 && /\d/.test(t[2] ?? ''))).toBe(true)
    const activo = grafo.querySelector('[data-activo] text')
    expect(activo?.textContent).toBe('Mergeado 2/3')
  })

  it('en desktop sigue horizontal, con su contenedor con scroll', async () => {
    abrir(`/trabajos/${encodeURIComponent(detalleStack.id)}`)
    const grafo = within(await screen.findByRole('dialog', { name: 'Sombras suaves' })).getByRole('img', { name: 'Grafo' })
    expect(grafo.getAttribute('width')).toBe('968')
    expect(grafo.getAttribute('height')).toBe('200')
    expect(grafo.parentElement?.className).toContain('overflow-x-auto')
    expect(grafo.querySelectorAll('[data-grafo-nodo]')).toHaveLength(9)
  })
})
