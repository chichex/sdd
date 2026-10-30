import type { KeyboardEvent } from 'react'

const KEYS = new Set(['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'])

function nodos(el: Element): HTMLElement[] {
  return Array.from(el.querySelectorAll<HTMLElement>('[data-nodo]'))
}

/**
 * Flechas del board: izquierda y derecha recorren los nodos de la tira; arriba
 * y abajo saltan a la tira vecina conservando la posición. Enter lo resuelve
 * el botón del nodo (abre el detalle).
 */
export function moverFoco(e: KeyboardEvent<HTMLElement>): void {
  if (!KEYS.has(e.key)) return
  const actual = (e.target as HTMLElement).closest<HTMLElement>('[data-nodo]')
  const tira = actual?.closest<HTMLElement>('[data-tira]')
  if (!actual || !tira) return
  const propios = nodos(tira)
  const i = propios.indexOf(actual)
  let destino: HTMLElement | undefined
  if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
    destino = propios[i + (e.key === 'ArrowRight' ? 1 : -1)]
  } else {
    const tiras = Array.from(e.currentTarget.querySelectorAll<HTMLElement>('[data-tira]'))
    const vecina = tiras[tiras.indexOf(tira) + (e.key === 'ArrowDown' ? 1 : -1)]
    if (vecina) {
      const otros = nodos(vecina)
      destino = otros[Math.min(i, otros.length - 1)]
    }
  }
  if (destino) {
    e.preventDefault()
    destino.focus()
  }
}
