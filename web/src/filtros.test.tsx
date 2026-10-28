import { act, cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import App from './App'
import { mockApi, mockEventSource } from './test/api-mock'
import { detalleStack, enPR, idea, sinTraza, stack, workspace } from './test/fixtures'

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

describe('filtro Mostrar con traza', () => {
  it('ofrece Activos, Con traza y Todos en ese orden', async () => {
    mockApi(workspace())
    mockEventSource()
    render(<App />)
    const mostrar = await screen.findByRole('radiogroup', { name: 'Mostrar' })
    expect(within(mostrar).getAllByRole('radio').map((r) => r.textContent)).toEqual(['Activos', 'Con traza', 'Todos'])
  })

  it('elegir Con traza escribe ?filtro=con-traza y pide ese filtro al server', async () => {
    const api = mockApi(workspace())
    mockEventSource()
    render(<App />)
    await screen.findByRole('region', { name: 'Tablero' })
    fireEvent.click(screen.getByRole('radio', { name: 'Con traza' }))
    expect(window.location.search).toBe('?filtro=con-traza')
    await waitFor(() => expect(api.calls.some((c) => c.url === '/api/workspace?filtro=con-traza')).toBe(true))
  })

  it('lee ?filtro=con-traza al cargar y en celular el Sidebar lo ofrece como fila', async () => {
    window.history.replaceState(null, '', '/?filtro=con-traza')
    vi.stubGlobal('matchMedia', (q: string) => ({ matches: false, media: q, addEventListener: () => {}, removeEventListener: () => {} }))
    const api = mockApi(workspace())
    mockEventSource()
    render(<App />)
    fireEvent.click(await screen.findByRole('button', { name: 'Menú' }))
    const dialog = await screen.findByRole('dialog', { name: 'sdd' })
    expect(api.calls.some((c) => c.url === '/api/workspace?filtro=con-traza')).toBe(true)
    const filas = ['Activos', 'Con traza', 'Todos'].map((name) => within(dialog).getByRole('button', { name }))
    expect(filas[0]?.compareDocumentPosition(filas[1] as Node)).toBe(Node.DOCUMENT_POSITION_FOLLOWING)
    expect(filas[1]?.compareDocumentPosition(filas[2] as Node)).toBe(Node.DOCUMENT_POSITION_FOLLOWING)
  })
})

const celular = () =>
  vi.stubGlobal('matchMedia', (q: string) => ({ matches: false, media: q, addEventListener: () => {}, removeEventListener: () => {} }))

async function tablero(path = '/', trabajos = [stack, enPR, idea, sinTraza]) {
  window.history.replaceState(null, '', path)
  const api = mockApi(workspace({ trabajos }))
  mockEventSource()
  render(<App />)
  return { api, board: await screen.findByRole('region', { name: 'Tablero' }) }
}

const tiras = (board: HTMLElement) => within(board).queryAllByRole('group').map((t) => t.getAttribute('aria-label'))

describe('chips de etapa', () => {
  it('una chip por etapa, en orden, con el conteo de la respuesta y sin presionar', async () => {
    await tablero()
    const chips = within(screen.getByRole('group', { name: 'Etapas' })).getAllByRole('button')
    expect(chips.map((c) => c.textContent)).toEqual(['Idea 1', 'Grill 1', 'Spec 0', 'Run 0', 'PR 1', 'Mergeado 1', 'Publicado 0'])
    expect(chips.every((c) => c.getAttribute('aria-pressed') === 'false')).toBe(true)
  })

  it('tocar Grill escribe ?etapa=grill, la presiona y deja solo los trabajos en Grill', async () => {
    const { board } = await tablero()
    const before = window.history.length
    fireEvent.click(screen.getByRole('button', { name: 'Grill 1' }))
    expect(window.location.search).toBe('?etapa=grill')
    expect(window.history.length).toBe(before)
    expect(screen.getByRole('button', { name: 'Grill 1' }).getAttribute('aria-pressed')).toBe('true')
    expect(tiras(board)).toEqual(['Grill viejo'])
  })

  it('multi-selección en el orden de las etapas, con conteos que no cambian; sin selección vuelve todo', async () => {
    const { board } = await tablero()
    fireEvent.click(screen.getByRole('button', { name: 'PR 1' }))
    fireEvent.click(screen.getByRole('button', { name: 'Idea 1' }))
    expect(window.location.search).toBe('?etapa=idea,pr')
    expect(tiras(board)).toEqual(['Exportar tablero', 'Idea sin artefactos'])
    expect(screen.getByRole('button', { name: 'Mergeado 1' })).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: 'PR 1' }))
    fireEvent.click(screen.getByRole('button', { name: 'Idea 1' }))
    expect(window.location.search).toBe('')
    expect(tiras(board)).toHaveLength(4)
  })

  it('un id desconocido en ?etapa se ignora', async () => {
    const { board } = await tablero('/?etapa=zzz,idea')
    expect(screen.getByRole('button', { name: 'Idea 1' }).getAttribute('aria-pressed')).toBe('true')
    expect(tiras(board)).toEqual(['Idea sin artefactos'])
  })

  it('en celular viven en la sección Etapas del Sidebar, que no se cierra al tocar una', async () => {
    celular()
    mockApi(workspace({ trabajos: [stack, enPR, idea, sinTraza] }))
    mockEventSource()
    render(<App />)
    await screen.findByRole('list', { name: 'Trabajos' })
    expect(screen.queryByRole('group', { name: 'Etapas' })).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: 'Menú' }))
    const dialog = await screen.findByRole('dialog', { name: 'sdd' })
    expect(within(dialog).getByText('Etapas')).toBeTruthy()
    fireEvent.click(within(dialog).getByRole('button', { name: 'Grill 1' }))
    expect(window.location.search).toBe('?etapa=grill')
    expect(within(dialog).getByRole('button', { name: 'Grill 1' }).getAttribute('aria-pressed')).toBe('true')
    expect(screen.getByRole('dialog', { name: 'sdd' })).toBeTruthy()
  })
})

describe('búsqueda', () => {
  it('escribir filtra en el cliente y escribe ?q= sin sumar historia; la X la limpia', async () => {
    const { api, board } = await tablero()
    const buscar = screen.getByRole('searchbox', { name: 'Buscar' })
    expect(buscar.getAttribute('placeholder')).toBe('Buscar trabajo, #NN o repo')
    const before = window.history.length
    const pedidos = api.calls.length
    fireEvent.change(buscar, { target: { value: '57' } })
    expect(window.location.search).toBe('?q=57')
    expect(window.history.length).toBe(before)
    expect(tiras(board)).toEqual(['Sombras suaves'])
    expect(api.calls.length).toBe(pedidos)
    fireEvent.click(screen.getByRole('button', { name: 'Limpiar búsqueda' }))
    expect(window.location.search).toBe('')
    expect((buscar as HTMLInputElement).value).toBe('')
    expect(tiras(board)).toHaveLength(4)
  })

  it('en celular la barra tiene solo la búsqueda', async () => {
    celular()
    mockApi(workspace())
    mockEventSource()
    render(<App />)
    await screen.findByRole('list', { name: 'Trabajos' })
    expect(screen.getByRole('searchbox', { name: 'Buscar' })).toBeTruthy()
    expect(screen.queryByRole('button', { name: /^Idea \d/ })).toBeNull()
  })
})

describe('vacío por filtros', () => {
  it('si etapa y búsqueda no dejan nada, muestra el EmptyState y Limpiar filtros conserva repo y filtro', async () => {
    window.history.replaceState(null, '', '/?repo=o%2Fa&filtro=todos&etapa=publicado&q=zzz')
    mockApi(workspace({ trabajos: [stack, enPR, idea, sinTraza] }))
    mockEventSource()
    render(<App />)
    expect(await screen.findByText('Ningún trabajo coincide con los filtros')).toBeTruthy()
    expect(screen.queryByRole('region', { name: 'Tablero' })).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: 'Limpiar filtros' }))
    expect(window.location.search).toBe('?repo=o%2Fa&filtro=todos')
    const board = await screen.findByRole('region', { name: 'Tablero' })
    expect(tiras(board)).toEqual(['Sombras suaves', 'Exportar tablero'])
  })

  it('sin trabajos en la respuesta se mantienen los carriles vacíos', async () => {
    const { board } = await tablero('/?etapa=idea', [])
    expect(within(board).getByText('Sin trabajos en Idea')).toBeTruthy()
    expect(screen.queryByText('Ningún trabajo coincide con los filtros')).toBeNull()
  })
})

describe('los cuatro filtros en la URL', () => {
  const query = '?repo=o%2Fa&filtro=todos&etapa=mergeado&q=sombras'

  it('se leen al cargar, se conservan al abrir y cerrar el detalle y sobreviven a una recarga', async () => {
    const { api, board } = await tablero(`/${query}`)
    api.detalles.set(detalleStack.id, detalleStack)
    expect(tiras(board)).toEqual(['Sombras suaves'])
    expect((screen.getByRole('searchbox', { name: 'Buscar' }) as HTMLInputElement).value).toBe('sombras')
    expect(screen.getByRole('button', { name: 'Mergeado 1' }).getAttribute('aria-pressed')).toBe('true')
    fireEvent.click(within(board).getAllByRole('button', { name: /^Sombras suaves/ })[0] as HTMLElement)
    expect(window.location.pathname + window.location.search).toBe(`/trabajos/${encodeURIComponent(detalleStack.id)}${query}`)
    fireEvent.keyDown(await screen.findByRole('dialog', { name: 'Sombras suaves' }), { key: 'Escape' })
    await waitFor(() => expect(window.location.pathname).toBe('/'))
    expect(window.location.search).toBe(query)
    cleanup() // recarga: se desmonta todo y se vuelve a montar con la misma URL
    render(<App />)
    const recargado = await screen.findByRole('region', { name: 'Tablero' })
    expect(tiras(recargado)).toEqual(['Sombras suaves'])
    expect(screen.getByRole('button', { name: 'Mergeado 1' }).getAttribute('aria-pressed')).toBe('true')
    expect(window.location.search).toBe(query)
  })

  it('el skeleton y el stream no los pisan', async () => {
    window.history.replaceState(null, '', `/${query}`)
    const api = mockApi(workspace({ scanning: true, trabajos: [stack, enPR, idea, sinTraza] }))
    const es = mockEventSource()
    render(<App />)
    await screen.findByRole('status', { name: 'Cargando el tablero' })
    expect(window.location.search).toBe(query)
    api.queueWorkspace(workspace({ trabajos: [stack, enPR, idea, sinTraza] }))
    act(() => es.instances[0]?.emit('snapshot', {}))
    const board = await screen.findByRole('region', { name: 'Tablero' })
    expect(window.location.search).toBe(query)
    expect(tiras(board)).toEqual(['Sombras suaves'])
  })
})
