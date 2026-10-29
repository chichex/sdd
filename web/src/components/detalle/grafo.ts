// Layout del grafo del detalle. Horizontal (desktop): la rama viva en una fila
// y cada rama cerrada (run fallado, spec reemplazada) en una fila propia que se
// bifurca desde el último nodo de la viva anterior a su etapa. Vertical
// (celular): un nodo por fila, la viva en una columna y cada rama cerrada en
// una columna a su derecha, con el texto de cada fila a la derecha de todas.
import type { Etapa, Nodo, Rama, Trabajo } from '@/lib/api'
import { ETAPA_INFO } from '../etapas'

export const GRAFO_PASO = 120
export const GRAFO_FILA = 96
const X0 = 64
const Y0 = 40

// Vertical: el viewBox mide siempre V_ANCHO y el SVG se estira al contenedor.
const V_ANCHO = 340
const V_X = 24 // columna de la rama viva
const V_Y0 = 24
const V_PASO = 56 // alto de una fila
const V_RAMA = 40 // separación entre columnas de nodos
const V_TEXTO = 24 // aire entre la última columna de nodos y el texto
const V_ETIQUETA = 76 // ancho de la columna de la etiqueta
const V_FECHA = 292 // la fecha corta ocupa el borde derecho
const V_CODO = 24 // radio del codo con que una rama cerrada se separa de la viva
const V_CHAR = 5.8 // ancho medio de un carácter `xs` (12 px), para recortar con «…»

export type Orientacion = 'horizontal' | 'vertical'

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

/** Columnas de texto del grafo vertical: dónde arranca cada una y cuántos caracteres entran. */
export interface GrafoColumnas {
  etiqueta: number
  subEstado: number
  fecha: number
  maxEtiqueta: number
  maxSubEstado: number
}

export interface GrafoLayout {
  ancho: number
  alto: number
  nodos: GrafoNodo[]
  aristas: GrafoArista[]
  /** Solo en vertical. */
  columnas?: GrafoColumnas
}

const RANK = new Map<Etapa, number>(ETAPA_INFO.map((e, i) => [e.id, i]))

/** Recorta con «…» lo que no entra en `max` caracteres. */
export function recortar(texto: string, max: number): string {
  return texto.length <= max ? texto : `${texto.slice(0, Math.max(max - 1, 0))}…`
}

/** El último nodo de la viva anterior a la etapa con que arranca la rama cerrada: de ahí se bifurca. */
function bifurcacion(principal: readonly GrafoNodo[], r: Rama): GrafoNodo | undefined {
  const rank = RANK.get(r.nodos[0]?.etapa ?? 'idea') ?? 0
  return [...principal].reverse().find((g) => (RANK.get(g.nodo.etapa) ?? 0) < rank)
}

/** Marca como activo el último nodo de la etapa actual en la rama viva. */
function marcarActivo(principal: GrafoNodo[], etapa: Etapa): void {
  const activo = principal.map((g) => g.nodo.etapa).lastIndexOf(etapa)
  if (activo >= 0 && principal[activo]) principal[activo].activo = true
}

/** Una arista recta entre cada par de nodos consecutivos de la viva. */
function aristasDeLaViva(principal: readonly GrafoNodo[]): GrafoArista[] {
  return principal.slice(1).map((g, i) => {
    const prev = principal[i] ?? g
    return { key: `a-${g.key}`, x1: prev.x, y1: prev.y, x2: g.x, y2: g.y, curva: false }
  })
}

/** Trazo SVG de una arista: recta, S horizontal (desktop) o codo y bajada hasta la rama (vertical). */
export function aristaPath(a: GrafoArista, orientacion: Orientacion): string {
  if (!a.curva) return `M ${a.x1} ${a.y1} L ${a.x2} ${a.y2}`
  if (orientacion === 'horizontal') return `M ${a.x1} ${a.y1} C ${a.x1 + 60} ${a.y1}, ${a.x2 - 60} ${a.y2}, ${a.x2} ${a.y2}`
  const codo = Math.min(V_CODO, a.y2 - a.y1)
  return `M ${a.x1} ${a.y1} C ${a.x2} ${a.y1}, ${a.x2} ${a.y1}, ${a.x2} ${a.y1 + codo} L ${a.x2} ${a.y2}`
}

/** Nodos y aristas del grafo; el nodo activo es el último de la etapa actual en la rama viva. */
export function grafoLayout(t: Trabajo, orientacion: Orientacion = 'horizontal'): GrafoLayout {
  return orientacion === 'vertical' ? layoutVertical(t) : layoutHorizontal(t)
}

function layoutHorizontal(t: Trabajo): GrafoLayout {
  const nodos: GrafoNodo[] = []
  const [viva, ...cerradas] = t.ramas
  const principal = (viva?.nodos ?? []).map((nodo, i) => ({ key: `principal-${i}`, rama: viva?.id ?? 'principal', nodo, x: X0 + i * GRAFO_PASO, y: Y0, activo: false }))
  marcarActivo(principal, t.etapa)
  nodos.push(...principal)
  const aristas = aristasDeLaViva(principal)
  cerradas
    .filter((r) => r.nodos.length > 0)
    .forEach((r, k) => {
      const fork = bifurcacion(principal, r)
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

/** Columnas de texto a la derecha de las `nCerradas` columnas de nodos extra. */
function columnasDeTexto(nCerradas: number): GrafoColumnas {
  const etiqueta = V_X + V_RAMA * nCerradas + V_TEXTO
  const subEstado = etiqueta + V_ETIQUETA
  return {
    etiqueta,
    subEstado,
    fecha: V_FECHA,
    maxEtiqueta: Math.floor((V_ETIQUETA - 4) / V_CHAR),
    maxSubEstado: Math.floor((V_FECHA - subEstado - 6) / V_CHAR),
  }
}

function layoutVertical(t: Trabajo): GrafoLayout {
  const [viva, ...resto] = t.ramas
  const cerradas = resto.filter((r) => r.nodos.length > 0)
  const nodos: GrafoNodo[] = []
  const y = (fila: number) => V_Y0 + fila * V_PASO
  const principal = (viva?.nodos ?? []).map((nodo, i) => ({ key: `principal-${i}`, rama: viva?.id ?? 'principal', nodo, x: V_X, y: y(i), activo: false }))
  marcarActivo(principal, t.etapa)
  nodos.push(...principal)
  const aristas = aristasDeLaViva(principal)
  let fila = principal.length // cada nodo ocupa su propia fila: las cerradas siguen debajo de la viva
  cerradas.forEach((r, k) => {
    const fork = bifurcacion(principal, r)
    let prev = fork ? { x: fork.x, y: fork.y } : null
    r.nodos.forEach((nodo, i) => {
      const g = { key: `${r.id}-${i}`, rama: r.id, nodo, x: V_X + V_RAMA * (k + 1), y: y(fila), activo: false }
      fila += 1
      nodos.push(g)
      if (prev) aristas.push({ key: `a-${g.key}`, x1: prev.x, y1: prev.y, x2: g.x, y2: g.y, curva: prev.x !== g.x })
      prev = { x: g.x, y: g.y }
    })
  })
  const alto = Math.max(...nodos.map((g) => g.y), 0) + V_Y0
  return { ancho: V_ANCHO, alto, nodos, aristas, columnas: columnasDeTexto(cerradas.length) }
}
