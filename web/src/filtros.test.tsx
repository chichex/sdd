import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import App from './App'
import { mockApi, mockEventSource } from './test/api-mock'
import { workspace } from './test/fixtures'

describe('filtros en la URL', () => {
  it('lee ?repo y ?filtro al cargar', async () => {
    window.history.replaceState(null, '', '/?repo=o%2Fb&filtro=todos')
    const api = mockApi(workspace())
    mockEventSource()
    render(<App />)
    const board = await screen.findByRole('region', { name: 'Tablero' })
    expect(api.calls.some((c) => c.url === '/api/workspace?filtro=todos')).toBe(true)
    expect(within(board).getAllByText('Idea sin artefactos').length).toBeGreaterThan(0)
    expect(within(board).queryByText('Sombras suaves')).toBeNull()
    expect(screen.getByRole('radio', { name: 'Todos', checked: true })).toBeTruthy()
  })

  it('cambiar el filtro escribe la URL con replaceState y refetchea', async () => {
    const api = mockApi(workspace())
    mockEventSource()
    render(<App />)
    await screen.findByRole('region', { name: 'Tablero' })
    const before = window.history.length
    fireEvent.click(screen.getByRole('radio', { name: 'Todos' }))
    expect(window.location.search).toBe('?filtro=todos')
    expect(window.history.length).toBe(before)
    await waitFor(() => expect(api.calls.some((c) => c.url === '/api/workspace?filtro=todos')).toBe(true))
  })

  it('el Select de repo lista los repos del snapshot con su conteo y escribe ?repo', async () => {
    mockApi(workspace())
    mockEventSource()
    render(<App />)
    await screen.findByRole('region', { name: 'Tablero' })
    const trigger = screen.getByRole('combobox', { name: 'Repo' })
    fireEvent.pointerDown(trigger, { button: 0, ctrlKey: false, pointerType: 'mouse' })
    fireEvent.click(trigger)
    const option = await screen.findByRole('option', { name: /o\/a\s*2/ })
    expect(screen.getByRole('option', { name: /o\/b\s*1/ })).toBeTruthy()
    fireEvent.click(option)
    await waitFor(() => expect(window.location.search).toBe('?repo=o%2Fa'))
  })
})
