import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'
import { mockApi, mockEventSource } from './test/api-mock'
import { workspace } from './test/fixtures'

describe('Sidebar en celular', () => {
  beforeEach(() => {
    vi.stubGlobal('matchMedia', (q: string) => ({ matches: false, media: q, addEventListener: () => {}, removeEventListener: () => {} }))
    mockApi(workspace())
    mockEventSource()
  })

  it('la hamburguesa abre el Sidebar con Repo, Mostrar, Tema y el pie', async () => {
    render(<App />)
    fireEvent.click(await screen.findByRole('button', { name: 'Menú' }))
    const dialog = await screen.findByRole('dialog', { name: 'sdd' })
    for (const section of ['Repo', 'Mostrar', 'Tema']) expect(within(dialog).getByText(section)).toBeTruthy()
    expect(within(dialog).getByText('9.9.9')).toBeTruthy()
    expect(within(dialog).getByText('/ws')).toBeTruthy()
    expect(within(dialog).getByRole('button', { name: /o\/a\s*2/ })).toBeTruthy()
  })

  it('Esc lo cierra y devuelve el foco a la hamburguesa', async () => {
    render(<App />)
    const menu = await screen.findByRole('button', { name: 'Menú' })
    menu.focus()
    fireEvent.click(menu)
    const dialog = await screen.findByRole('dialog', { name: 'sdd' })
    fireEvent.keyDown(dialog, { key: 'Escape' })
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())
    await waitFor(() => expect(document.activeElement).toBe(menu))
  })

  it('elegir un repo cierra el Sidebar y escribe ?repo', async () => {
    render(<App />)
    fireEvent.click(await screen.findByRole('button', { name: 'Menú' }))
    const dialog = await screen.findByRole('dialog', { name: 'sdd' })
    fireEvent.click(within(dialog).getByRole('button', { name: /o\/b\s*1/ }))
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())
    expect(window.location.search).toBe('?repo=o%2Fb')
    expect(screen.getByText('sdd · o/b')).toBeTruthy()
  })
})
