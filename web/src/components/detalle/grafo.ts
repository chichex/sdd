// Layout del grafo del detalle: la rama viva en una fila y cada rama cerrada
// (run fallado, spec reemplazada) en una fila propia que se bifurca desde el
// último nodo de la viva anterior a su etapa.
import type { Etapa, Nodo, Trabajo } from '@/lib/api'
import { ETAPA_INFO } from '../etapas'

export const GRAFO_PASO = 120
export const GRAFO_FILA = 96
const X0 = 64
const Y0 = 40

export interface GrafoNodo {
  key: string
  rama: string
  nodo: Nodo
  x: number
  y: number
  activo: boolean
}

export interface GrafoArista {
  key: string
  x1: number
  y1: number
  x2: number
  y2: number
  curva: boolean
}

const RANK = new Map<Etapa, number>(ETAPA_INFO.map((e, i) => [e.id, i]))

/** Nodos y aristas del grafo; el nodo activo es el último de la etapa actual en la rama viva. */
export function grafoLayout(t: Trabajo): { ancho: number; alto: number; nodos: GrafoNodo[]; aristas: GrafoArista[] } {
  const nodos: GrafoNodo[] = []
  const aristas: GrafoArista[] = []
  const [viva, ...cerradas] = t.ramas
  const principal = (viva?.nodos ?? []).map((nodo, i) => ({ key: `principal-${i}`, rama: viva?.id ?? 'principal', nodo, x: X0 + i * GRAFO_PASO, y: Y0, activo: false }))
  const activo = principal.map((g) => g.nodo.etapa).lastIndexOf(t.etapa)
  if (activo >= 0 && principal[activo]) principal[activo].activo = true
  nodos.push(...principal)
  principal.slice(1).forEach((g, i) => {
    const prev = principal[i]
    if (prev) aristas.push({ key: `a-${g.key}`, x1: prev.x, y1: prev.y, x2: g.x, y2: g.y, curva: false })
  })
  cerradas
    .filter((r) => r.nodos.length > 0)
    .forEach((r, k) => {
      const rank = RANK.get(r.nodos[0]?.etapa ?? 'idea') ?? 0
      const fork = [...principal].reverse().find((g) => (RANK.get(g.nodo.etapa) ?? 0) < rank)
      const y = Y0 + (k + 1) * GRAFO_FILA
      let prev = fork ? { x: fork.x, y: fork.y } : null
      r.nodos.forEach((nodo, i) => {
        const g = { key: `${r.id}-${i}`, rama: r.id, nodo, x: (fork?.x ?? X0 - GRAFO_PASO) + (i + 1) * GRAFO_PASO, y, activo: false }
        nodos.push(g)
        if (prev) aristas.push({ key: `a-${g.key}`, x1: prev.x, y1: prev.y, x2: g.x, y2: g.y, curva: prev.y !== g.y })
        prev = { x: g.x, y: g.y }
      })
    })
  const ancho = Math.max(...nodos.map((g) => g.x), 0) + X0
  const alto = Math.max(...nodos.map((g) => g.y), 0) + Y0 + 24
  return { ancho, alto, nodos, aristas }
}
