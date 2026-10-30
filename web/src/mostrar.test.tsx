import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import App from './App'
import { mockApi, mockEventSource } from './test/api-mock'
import { cerrado, detalleStack, enPR, idea, sinTraza, stack, workspace } from './test/fixtures'

const TODOS = [stack, enPR, idea, cerrado, sinTraza]
const celular = () =>
  vi.stubGlobal('matchMedia', (q: string) => ({ matches: false, media: q, addEventListener: () => {}, removeEventListener: () => {} }))

/** Monta la app contra un server que filtra por ?mostrar= como el real. */
async function montar(path = '/', trabajos = TODOS) {
  window.history.replaceState(null, '', path)
  const api = mockApi(workspace({ trabajos }), { servidor: true })
  api.detalles.set(detalleStack.id, detalleStack)
  mockEventSource()
  render(<App />)
  return api
}

async function tablero(path?: string) {
  const api = await montar(path)
  return { api, board: await screen.findByRole('region', { name: 'Tablero' }) }
}

const tiras = (board: HTMLElement) => within(board).queryAllByRole('group').map((t) => t.getAttribute('aria-label'))
const pedidos = (api: Awaited<ReturnType<typeof montar>>) => api.calls.filter((c) => c.url.startsWith('/api/workspace')).map((c) => c.url)
const chip = (name: string) => screen.getByRole('button', { name })

describe('Mostrar en desktop', () => {
  it('es un grupo con cuatro botones en orden, con los conteos del server y En curso e Ideas presionadas', async () => {
    await tablero()
    const grupo = screen.getByRole('group', { name: 'Mostrar' })
    const botones = within(grupo).getAllByRole('button')
    expect(botones.map((b) => b.textContent)).toEqual(['En curso 2', 'Ideas 1', 'Terminados 1', 'Sin traza 1'])
    expect(botones.map((b) => b.getAttribute('aria-pressed'))).toEqual(['true', 'true', 'false', 'false'])
  })

  it('va en la Barra de filtros antes del grupo Etapas', async () => {
    await tablero()
    const mostrar = screen.getByRole('group', { name: 'Mostrar' })
    expect(mostrar.compareDocumentPosition(screen.getByRole('group', { name: 'Etapas' }))).toBe(Node.DOCUMENT_POSITION_FOLLOWING)
  })

  it('la cabecera ya no tiene el segmented «Mostrar» y conserva el de Tema', async () => {
    await tablero()
    expect(screen.queryByRole('radiogroup', { name: 'Mostrar' })).toBeNull()
    expect(screen.getByRole('radiogroup', { name: 'Tema' })).toBeTruthy()
  })

  it('sin elegir nada pide /api/workspace sin ?mostrar= y la URL queda limpia', async () => {
    const { api, board } = await tablero()
    expect(pedidos(api)).toEqual(['/api/workspace'])
    expect(window.location.search).toBe('')
    expect(tiras(board)).toEqual(['Sombras suaves', 'Exportar tablero', 'Idea sin artefactos'])
  })

  it('tocar Ideas escribe ?mostrar=en-curso sin sumar historia, pide esa lista y saca las tiras de Idea', async () => {
    const { api, board } = await tablero()
    const historia = window.history.length
    fireEvent.click(chip('Ideas 1'))
    expect(window.location.search).toBe('?mostrar=en-curso')
    expect(window.history.length).toBe(historia)
    expect(chip('Ideas 1').getAttribute('aria-pressed')).toBe('false')
    await waitFor(() => expect(pedidos(api)).toContain('/api/workspace?mostrar=en-curso'))
    await waitFor(() => expect(tiras(board)).toEqual(['Sombras suaves', 'Exportar tablero']))
    expect(chip('Ideas 1')).toBeTruthy() // el conteo sigue diciendo cuántas esconde
  })

  it('vuelve a pedir sin ?mostrar= al volver al default', async () => {
    const { api } = await tablero('/?mostrar=en-curso')
    fireEvent.click(chip('Ideas 1'))
    expect(window.location.search).toBe('')
    await waitFor(() => expect(pedidos(api).at(-1)).toBe('/api/workspace'))
  })

  it('suma categorías en el orden canónico', async () => {
    await tablero('/?mostrar=en-curso')
    fireEvent.click(chip('Sin traza 1'))
    fireEvent.click(chip('Terminados 1'))
    expect(window.location.search).toBe('?mostrar=en-curso,terminado,sin-traza')
    fireEvent.click(chip('Ideas 1'))
    expect(window.location.search).toBe('?mostrar=en-curso,idea,terminado,sin-traza')
  })

  it('tocar la única categoría activa no la destilda ni pide nada', async () => {
    const { api } = await tablero('/?mostrar=en-curso')
    const antes = pedidos(api).length
    fireEvent.click(chip('En curso 2'))
    expect(window.location.search).toBe('?mostrar=en-curso')
    expect(chip('En curso 2').getAttribute('aria-pressed')).toBe('true')
    expect(pedidos(api)).toHaveLength(antes)
  })

  it('lee ?mostrar= al cargar', async () => {
    await tablero('/?mostrar=terminado')
    expect(chip('Terminados 1').getAttribute('aria-pressed')).toBe('true')
    expect(chip('En curso 2').getAttribute('aria-pressed')).toBe('false')
    await waitFor(() => expect(tiras(screen.getByRole('region', { name: 'Tablero' }))).toEqual(['Modo oscuro']))
  })

  it('un ?filtro=todos viejo se ignora', async () => {
    const { api } = await tablero('/?filtro=todos')
    expect(pedidos(api)).toEqual(['/api/workspace'])
    expect(chip('Terminados 1').getAttribute('aria-pressed')).toBe('false')
  })
})

describe('grupo «Sin traza (N)»', () => {
  it('aparece, cerrado, cuando la categoría Sin traza está elegida, y al abrirlo trae sus trabajos', async () => {
    const { board } = await tablero('/?mostrar=en-curso,sin-traza')
    const grupo = await within(board).findByRole('button', { name: 'Sin traza (1)' })
    expect(grupo.getAttribute('aria-expanded')).toBe('false')
    expect(within(board).queryByRole('group', { name: 'Grill viejo' })).toBeNull()
    fireEvent.click(grupo)
    expect(within(board).getByRole('group', { name: 'Grill viejo' })).toBeTruthy()
  })

  it('tocar la chip Sin traza lo hace aparecer', async () => {
    const { board } = await tablero()
    expect(within(board).queryByRole('button', { name: /^Sin traza/ })).toBeNull()
    fireEvent.click(chip('Sin traza 1'))
    expect(await within(board).findByRole('button', { name: 'Sin traza (1)' })).toBeTruthy()
  })

  it('las Ideas van en línea con los demás, no en el grupo', async () => {
    const { board } = await tablero('/?mostrar=idea,sin-traza')
    expect(tiras(board)).toEqual(['Idea sin artefactos'])
    expect(await within(board).findByRole('button', { name: 'Sin traza (1)' })).toBeTruthy()
  })
})

describe('Mostrar en celular', () => {
  it('las cuatro filas viven en la sección Mostrar del Sidebar y tocar una no lo cierra', async () => {
    celular()
    const api = await montar()
    await screen.findByRole('list', { name: 'Trabajos' })
    expect(screen.queryByRole('group', { name: 'Mostrar' })).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: 'Menú' }))
    const dialog = await screen.findByRole('dialog', { name: 'sdd' })
    expect(within(dialog).getByRole('heading', { name: 'Mostrar' })).toBeTruthy()
    const filas = ['En curso 2', 'Ideas 1', 'Terminados 1', 'Sin traza 1'].map((name) => within(dialog).getByRole('button', { name }))
    expect(filas.map((f) => f.getAttribute('aria-pressed'))).toEqual(['true', 'true', 'false', 'false'])
    fireEvent.click(filas[1] as HTMLElement)
    expect(window.location.search).toBe('?mostrar=en-curso')
    expect(screen.getByRole('dialog', { name: 'sdd' })).toBeTruthy()
    expect(within(dialog).getByRole('button', { name: 'Ideas 1' }).getAttribute('aria-pressed')).toBe('false')
    await waitFor(() => expect(pedidos(api)).toContain('/api/workspace?mostrar=en-curso'))
  })

  it('la última categoría activa no se destilda en el Sidebar', async () => {
    celular()
    await montar('/?mostrar=terminado')
    fireEvent.click(await screen.findByRole('button', { name: 'Menú' }))
    const dialog = await screen.findByRole('dialog', { name: 'sdd' })
    fireEvent.click(within(dialog).getByRole('button', { name: 'Terminados 1' }))
    expect(window.location.search).toBe('?mostrar=terminado')
    expect(within(dialog).getByRole('button', { name: 'Terminados 1' }).getAttribute('aria-pressed')).toBe('true')
  })
})

describe('Mostrar con el detalle', () => {
  it('abrir y cerrar el detalle conserva mostrar', async () => {
    const query = '?mostrar=en-curso,terminado'
    const { board } = await tablero(`/${query}`)
    fireEvent.click(within(board).getAllByRole('button', { name: /^Sombras suaves/ })[0] as HTMLElement)
    expect(window.location.pathname + window.location.search).toBe(`/trabajos/${encodeURIComponent(detalleStack.id)}${query}`)
    fireEvent.keyDown(await screen.findByRole('dialog', { name: 'Sombras suaves' }), { key: 'Escape' })
    await waitFor(() => expect(window.location.pathname).toBe('/'))
    expect(window.location.search).toBe(query)
  })
})
