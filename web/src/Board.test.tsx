import { cleanup, fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import App from './App'
import { mockApi, mockEventSource } from './test/api-mock'
import { enPR, idea, sinTraza, specSuelta, stack, workspace } from './test/fixtures'

async function board() {
  mockApi(workspace())
  mockEventSource()
  render(<App />)
  return screen.findByRole('region', { name: 'Tablero' })
}

describe('board en desktop', () => {
  it('tiene siete carriles con nombre y conteo de nodos', async () => {
    const b = await board()
    const want: Array<[string, number]> = [['Idea', 1], ['Grill', 1], ['Spec', 2], ['Run', 2], ['PR', 4], ['Mergeado', 2], ['Publicado', 0]]
    for (const [name, count] of want) {
      expect(within(b).getByRole('heading', { name: `${name} ${count}` })).toBeTruthy()
    }
    expect(within(b).getByText('Sin trabajos en Publicado')).toBeTruthy()
  })

  it('una tira por trabajo ordenada por ultimaActividad, stacks n/N y nodos stale', async () => {
    const b = await board()
    const tiras = within(b).getAllByRole('group')
    expect(tiras.map((t) => t.getAttribute('aria-label'))).toEqual(['Sombras suaves', 'Exportar tablero', 'Idea sin artefactos'])
    expect(within(tiras[0]).getAllByText('1/3')).toHaveLength(2)
    expect(within(tiras[0]).getByText('fallado (interrupted)')).toBeTruthy()
    const stale = within(tiras[2]).getByRole('button', { name: /Idea sin artefactos · idea/ })
    expect(stale.dataset.stale).toBe('true')
  })

  it('las flechas mueven el foco entre nodos y Enter abre el detalle', async () => {
    const b = await board()
    const nodos = within(b).getAllByRole('button')
    const first = nodos[0]
    first.focus()
    fireEvent.keyDown(first, { key: 'ArrowRight' })
    expect(document.activeElement).toBe(nodos[1])
    fireEvent.keyDown(nodos[1], { key: 'ArrowLeft' })
    expect(document.activeElement).toBe(first)
    fireEvent.keyDown(first, { key: 'ArrowDown' })
    const tiras = within(b).getAllByRole('group')
    expect(tiras[1].contains(document.activeElement)).toBe(true)
    fireEvent.keyDown(document.activeElement as Element, { key: 'Enter' })
    fireEvent.click(document.activeElement as Element)
    expect(window.location.pathname).toBe(`/trabajos/${encodeURIComponent('o/a/issue/58')}`)
  })
})

describe('marcas de traza y ancla en el nodo', () => {
  async function boardCon(...trabajos: Parameters<typeof workspace>[0][]) {
    mockApi(workspace(Object.assign({}, ...trabajos)))
    mockEventSource()
    render(<App />)
    return screen.findByRole('region', { name: 'Tablero' })
  }

  it('un trabajo sin traza lleva el badge «sin traza» en la fila del sub-estado, después de él', async () => {
    const b = await boardCon({ trabajos: [stack, sinTraza] })
    const nodo = within(within(b).getByRole('group', { name: 'Grill viejo' })).getByRole('button', { name: /Grill viejo · grill/ })
    const badge = within(nodo).getByText('sin traza')
    const fila = badge.parentElement as HTMLElement
    expect(within(fila).getByText('pausado').compareDocumentPosition(badge)).toBe(Node.DOCUMENT_POSITION_FOLLOWING)
  })

  it('un trabajo con traza no lleva el badge', async () => {
    const b = await boardCon({ trabajos: [stack, sinTraza] })
    expect(within(within(b).getByRole('group', { name: 'Sombras suaves' })).queryByText('sin traza')).toBeNull()
  })

  it('el chip de referencia del nodo lleva el ícono de ancla por tipo, con aria-label', async () => {
    const b = await boardCon({ trabajos: [stack, sinTraza, specSuelta] })
    expect(within(within(b).getByRole('group', { name: 'Sombras suaves' })).getAllByLabelText('ancla: issue').length).toBeGreaterThan(0)
    expect(within(within(b).getByRole('group', { name: 'Grill viejo' })).getByLabelText('ancla: grill')).toBeTruthy()
    expect(within(within(b).getByRole('group', { name: 'Spec suelta' })).getByLabelText('ancla: spec')).toBeTruthy()
  })

  it('el ícono de ancla muestra un tooltip con el mismo texto', async () => {
    const b = await boardCon({ trabajos: [specSuelta] })
    fireEvent.pointerMove(within(b).getByLabelText('ancla: spec'), { pointerType: 'mouse' })
    expect((await screen.findByRole('tooltip')).textContent).toBe('ancla: spec')
  })
})

describe('grupos por repo y grupo sin traza en el board', () => {
  async function boardEn(path: string, trabajos = [stack, enPR, idea, sinTraza]) {
    window.history.replaceState(null, '', path)
    mockApi(workspace({ trabajos }))
    mockEventSource()
    render(<App />)
    return screen.findByRole('region', { name: 'Tablero' })
  }

  it('un encabezado h3 por repo en orden alfabético, con su conteo, y debajo sus tiras', async () => {
    const b = await boardEn('/', [idea, enPR, stack])
    const encabezados = within(b).getAllByRole('heading', { level: 3 })
    expect(encabezados.map((h) => h.textContent)).toEqual(['o/a', 'o/b'])
    expect(within(encabezados[0]?.parentElement as HTMLElement).getByText('2')).toBeTruthy()
    const tiraB = within(b).getByRole('group', { name: 'Idea sin artefactos' })
    expect(encabezados[1]?.compareDocumentPosition(tiraB)).toBe(Node.DOCUMENT_POSITION_FOLLOWING)
    expect(within(b).getAllByRole('group').map((t) => t.getAttribute('aria-label'))).toEqual(['Sombras suaves', 'Exportar tablero', 'Idea sin artefactos'])
  })

  it('con trabajos de un solo repo no hay encabezados', async () => {
    const b = await boardEn('/?repo=o%2Fa')
    expect(within(b).queryAllByRole('heading', { level: 3 })).toHaveLength(0)
  })

  it('con «todos» los sin traza van al final en «Sin traza (N)», cerrado, y al abrirlo aparecen por repo', async () => {
    const b = await boardEn('/?filtro=todos')
    const grupo = within(b).getByRole('button', { name: 'Sin traza (1)' })
    expect(grupo.getAttribute('aria-expanded')).toBe('false')
    expect(within(b).queryByRole('group', { name: 'Grill viejo' })).toBeNull()
    const ultima = within(b).getByRole('group', { name: 'Idea sin artefactos' })
    expect(ultima.compareDocumentPosition(grupo)).toBe(Node.DOCUMENT_POSITION_FOLLOWING)
    fireEvent.click(grupo)
    expect(grupo.getAttribute('aria-expanded')).toBe('true')
    const tira = within(b).getByRole('group', { name: 'Grill viejo' })
    expect(grupo.compareDocumentPosition(tira)).toBe(Node.DOCUMENT_POSITION_FOLLOWING)
    expect(within(b).getAllByRole('heading', { level: 3 }).map((h) => h.textContent)).toEqual(['o/a', 'o/b', 'o/b'])
  })

  it('los carriles cuentan los nodos de todo lo visible, incluido el grupo sin traza', async () => {
    const b = await boardEn('/?filtro=todos')
    expect(within(b).getByRole('heading', { name: 'Grill 2' })).toBeTruthy()
  })

  it.each(['/', '/?filtro=con-traza'])('en %s no existe el grupo sin traza', async (path) => {
    const b = await boardEn(path)
    expect(within(b).queryByRole('button', { name: /^Sin traza/ })).toBeNull()
  })

  it('N cuenta tras etapa y búsqueda, y con N = 0 el grupo no se dibuja', async () => {
    const b = await boardEn('/?filtro=todos&q=viejo')
    expect(within(b).getByRole('button', { name: 'Sin traza (1)' })).toBeTruthy()
    cleanup()
    const b2 = await boardEn('/?filtro=todos&etapa=idea')
    expect(within(b2).queryByRole('button', { name: /^Sin traza/ })).toBeNull()
  })
})
