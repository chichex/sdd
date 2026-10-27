import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import App from './App'
import { mockApi, mockEventSource } from './test/api-mock'
import { workspace } from './test/fixtures'

describe('estados del tablero', () => {
  it('muestra el skeleton hasta el primer snapshot con scanning false', async () => {
    const api = mockApi(workspace({ scanning: true }))
    const es = mockEventSource()
    render(<App />)
    expect(await screen.findByRole('status', { name: 'Cargando el tablero' })).toBeTruthy()
    expect(screen.queryByRole('region', { name: 'Tablero' })).toBeNull()
    api.queueWorkspace(workspace())
    act(() => es.instances[0]?.emit('snapshot', { scannedAt: 'x', fetchedAt: null }))
    expect(await screen.findByRole('region', { name: 'Tablero' })).toBeTruthy()
    expect(screen.queryByRole('status', { name: 'Cargando el tablero' })).toBeNull()
  })

  it('sin GitHub muestra el banner persistente y Reintentar hace POST /api/refresh', async () => {
    const api = mockApi(workspace({ github: { ok: false, reason: 'red', fetchedAt: null } }))
    mockEventSource()
    render(<App />)
    const banner = await screen.findByRole('status', { name: 'Aviso' })
    expect(within(banner).getByText('Sin GitHub: mostrando solo lo local')).toBeTruthy()
    fireEvent.click(within(banner).getByRole('button', { name: 'Reintentar' }))
    await waitFor(() => expect(api.calls.some((c) => c.method === 'POST' && c.url === '/api/refresh')).toBe(true))
  })

  it('si /api/workspace falla muestra el banner de error y conserva el último dato', async () => {
    const api = mockApi(workspace())
    const es = mockEventSource()
    render(<App />)
    expect(await screen.findAllByText('Sombras suaves')).not.toHaveLength(0)
    api.queueWorkspace({ status: 500, body: { error: 'boom' } })
    act(() => es.instances[0]?.emit('snapshot', {}))
    expect(await screen.findByRole('alert')).toBeTruthy()
    expect(screen.getAllByText('Sombras suaves').length).toBeGreaterThan(0)
  })

  it('sin repos muestra el EmptyState de página con el comando', async () => {
    mockApi(workspace({ repos: [], trabajos: [] }))
    mockEventSource()
    render(<App />)
    expect(await screen.findByText('sdd --dir <otra carpeta>')).toBeTruthy()
    expect(screen.queryByRole('region', { name: 'Tablero' })).toBeNull()
  })

  it('se suscribe a /api/stream, refetchea en cada snapshot y reconecta tras un corte', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    try {
      const api = mockApi(workspace())
      const es = mockEventSource()
      render(<App />)
      await screen.findByRole('region', { name: 'Tablero' })
      const before = api.calls.filter((c) => c.url.startsWith('/api/workspace')).length
      expect(es.instances[0]?.url).toBe('/api/stream')
      act(() => es.instances[0]?.emit('snapshot', {}))
      await waitFor(() => expect(api.calls.filter((c) => c.url.startsWith('/api/workspace')).length).toBe(before + 1))
      act(() => es.instances[0]?.fail())
      await act(async () => {
        await vi.advanceTimersByTimeAsync(3000)
      })
      expect(es.instances).toHaveLength(2)
    } finally {
      vi.useRealTimers()
    }
  })
})

describe('refresh y accesibilidad', () => {
  it('Refresh hace POST /api/refresh, muestra … mientras espera y un toast al terminar', async () => {
    const api = mockApi(workspace())
    mockEventSource()
    render(<App />)
    await screen.findByRole('region', { name: 'Tablero' })
    const release = api.holdRefresh()
    const button = screen.getByRole('button', { name: 'Actualizar' })
    fireEvent.click(button)
    expect(await within(button).findByText('…')).toBeTruthy()
    expect(button.getAttribute('aria-busy')).toBe('true')
    release()
    expect(await screen.findByText(/^Actualizado/)).toBeTruthy()
    expect(api.calls.filter((c) => c.method === 'POST')).toHaveLength(1)
  })

  it('tiene aria-label en hamburguesa, Refresh y tema, y ningún select nativo', async () => {
    mockApi(workspace())
    mockEventSource()
    vi.stubGlobal('matchMedia', (q: string) => ({ matches: false, media: q, addEventListener: () => {}, removeEventListener: () => {} }))
    const { container } = render(<App />)
    await screen.findByRole('list', { name: 'Trabajos' })
    expect(screen.getByRole('button', { name: 'Menú' })).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Actualizar' })).toBeTruthy()
    expect(container.querySelector('select')).toBeNull()
  })

  it('el tema elegido se aplica y persiste en localStorage', async () => {
    mockApi(workspace())
    mockEventSource()
    const first = render(<App />)
    const tema = await screen.findByRole('radiogroup', { name: 'Tema' })
    fireEvent.click(within(tema).getByRole('radio', { name: 'Oscuro' }))
    expect(document.documentElement.dataset.theme).toBe('dark')
    expect(window.localStorage.getItem('sdd-tema')).toBe('oscuro')
    first.unmount()
    document.documentElement.removeAttribute('data-theme')
    render(<App />)
    expect(await screen.findByRole('radio', { name: 'Oscuro', checked: true })).toBeTruthy()
    expect(document.documentElement.dataset.theme).toBe('dark')
  })
})
