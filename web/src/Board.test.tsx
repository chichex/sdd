import { fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import App from './App'
import { mockApi, mockEventSource } from './test/api-mock'
import { workspace } from './test/fixtures'

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
