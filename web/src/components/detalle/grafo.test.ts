import { describe, expect, it } from 'vitest'
import type { Trabajo } from '@/lib/api'
import { detalleStack, stack } from '@/test/fixtures'
import { aristaPath, grafoLayout, recortar, type GrafoLayout } from './grafo'

const proyeccion = (l: GrafoLayout) => ({
  ancho: l.ancho,
  alto: l.alto,
  nodos: l.nodos.map((g) => ({ key: g.key, rama: g.rama, etapa: g.nodo.etapa, x: g.x, y: g.y, activo: g.activo })),
  aristas: l.aristas,
})

/** El nodo del que sale la arista a `hacia`: la bifurcación de una rama cerrada. */
function bifurcacion(l: GrafoLayout, hacia: string): string | undefined {
  const arista = l.aristas.find((a) => a.key === `a-${hacia}`)
  return l.nodos.find((g) => g.x === arista?.x1 && g.y === arista?.y1)?.key
}

const conNodos = (n: number): Trabajo => {
  const viva = stack.ramas[0]
  return { ...stack, etapa: 'spec', ramas: [{ id: 'principal', estado: 'viva', nodos: (viva?.nodos ?? []).slice(0, n) }] }
}

describe('grafoLayout vertical', () => {
  const v = grafoLayout(detalleStack, 'vertical')
  const principal = v.nodos.filter((g) => g.rama === 'principal')
  const cerrada = v.nodos.filter((g) => g.rama === 'run-1')

  it('la rama viva es una columna: x constante e y creciente de a un paso fijo', () => {
    expect(principal).toHaveLength(8)
    expect(new Set(principal.map((g) => g.x)).size).toBe(1)
    const pasos = principal.slice(1).map((g, i) => g.y - (principal[i]?.y ?? 0))
    expect(pasos.every((p) => p > 0 && p === pasos[0])).toBe(true)
  })

  it('cada rama cerrada es una columna a la derecha de la viva, debajo de sus nodos', () => {
    const viva = principal[0]
    expect(cerrada).toHaveLength(1)
    expect(cerrada[0]?.x).toBeGreaterThan(viva?.x ?? Infinity)
    expect(cerrada[0]?.y).toBeGreaterThan(principal.at(-1)?.y ?? Infinity)
  })

  it('el primer nodo de la cerrada se une con una arista curva al mismo nodo de bifurcación que en horizontal', () => {
    const h = grafoLayout(detalleStack)
    expect(bifurcacion(h, 'run-1-0')).toBe('principal-1')
    expect(bifurcacion(v, 'run-1-0')).toBe('principal-1')
    expect(v.aristas.find((a) => a.key === 'a-run-1-0')?.curva).toBe(true)
  })

  it('las aristas de la viva son verticales y rectas', () => {
    const rectas = v.aristas.filter((a) => a.key.startsWith('a-principal'))
    expect(rectas).toHaveLength(7)
    expect(rectas.every((a) => !a.curva && a.x1 === a.x2 && a.y2 > a.y1)).toBe(true)
  })

  it('el ancho no depende de la cantidad de nodos y el alto crece con la fila más baja', () => {
    const chico = grafoLayout(conNodos(2), 'vertical')
    const grande = grafoLayout(conNodos(8), 'vertical')
    expect(chico.ancho).toBe(grande.ancho)
    expect(chico.ancho).toBe(v.ancho)
    expect(grande.alto).toBeGreaterThan(chico.alto)
    expect(v.alto).toBeGreaterThan(Math.max(...v.nodos.map((g) => g.y)))
  })

  it('el nodo activo sigue siendo el último de la etapa actual de la viva', () => {
    expect(v.nodos.filter((g) => g.activo).map((g) => g.key)).toEqual(['principal-7'])
  })

  it('el texto de cada fila arranca a la derecha de todas las columnas de nodos', () => {
    const borde = Math.max(...v.nodos.map((g) => g.x)) + 16
    expect(v.columnas?.etiqueta).toBeGreaterThan(borde)
    expect(v.columnas?.subEstado).toBeGreaterThan(v.columnas?.etiqueta ?? Infinity)
    expect(v.columnas?.fecha).toBeGreaterThan(v.columnas?.subEstado ?? Infinity)
    expect(v.columnas?.fecha).toBeLessThan(v.ancho)
  })

  it('sin ramas cerradas el texto queda más cerca de la columna', () => {
    const sin = grafoLayout(conNodos(4), 'vertical')
    expect(sin.columnas?.etiqueta).toBeLessThan(v.columnas?.etiqueta ?? 0)
  })
})

describe('geometría del grafo vertical (spec, inferencia 12)', () => {
  const v = grafoLayout(detalleStack, 'vertical')
  const nodo = (key: string) => v.nodos.find((g) => g.key === key)

  it('la viva está en x 24 con un paso de fila de 56 px, la cerrada 40 px a la derecha y el viewBox mide 340', () => {
    expect(v.nodos.filter((g) => g.rama === 'principal').map((g) => [g.x, g.y - (nodo('principal-0')?.y ?? 0)])).toEqual(
      [0, 56, 112, 168, 224, 280, 336, 392].map((dy) => [24, dy]),
    )
    expect(nodo('run-1-0')?.x).toBe(64)
    expect(v.ancho).toBe(340)
  })
})

describe('aristaPath', () => {
  it('la recta va de un nodo al otro en las dos orientaciones', () => {
    const recta = { key: 'a', x1: 10, y1: 20, x2: 10, y2: 76, curva: false }
    expect(aristaPath(recta, 'vertical')).toBe('M 10 20 L 10 76')
    expect(aristaPath(recta, 'horizontal')).toBe('M 10 20 L 10 76')
  })

  it('en horizontal la curva es la S de siempre', () => {
    const curva = { key: 'a-run-1-0', x1: 184, y1: 40, x2: 304, y2: 136, curva: true }
    expect(aristaPath(curva, 'horizontal')).toBe('M 184 40 C 244 40, 244 136, 304 136')
  })

  it('en vertical la curva sale hacia la derecha del nodo de bifurcación, dobla y baja hasta la rama', () => {
    const curva = { key: 'a-run-1-0', x1: 24, y1: 80, x2: 64, y2: 472, curva: true }
    expect(aristaPath(curva, 'vertical')).toBe('M 24 80 C 64 80, 64 80, 64 104 L 64 472')
  })
})

describe('recortar', () => {
  it.each([
    ['fallado (interrupted)', 21, 'fallado (interrupted)'],
    ['fallado (interrupted)', 10, 'fallado ('+'…'],
    ['abc', 3, 'abc'],
    ['abc', 0, '…'],
  ])('«%s» en %i caracteres → «%s»', (texto, max, want) => {
    expect(recortar(texto, max)).toBe(want)
  })
})

describe('grafoLayout horizontal', () => {
  // Salida de grafoLayout(detalleStack) antes de existir la variante vertical.
  const hoy = {
    ancho: 968,
    alto: 200,
    nodos: [
      { key: 'principal-0', rama: 'principal', etapa: 'grill', x: 64, y: 40, activo: false },
      { key: 'principal-1', rama: 'principal', etapa: 'spec', x: 184, y: 40, activo: false },
      { key: 'principal-2', rama: 'principal', etapa: 'run', x: 304, y: 40, activo: false },
      { key: 'principal-3', rama: 'principal', etapa: 'pr', x: 424, y: 40, activo: false },
      { key: 'principal-4', rama: 'principal', etapa: 'pr', x: 544, y: 40, activo: false },
      { key: 'principal-5', rama: 'principal', etapa: 'pr', x: 664, y: 40, activo: false },
      { key: 'principal-6', rama: 'principal', etapa: 'mergeado', x: 784, y: 40, activo: false },
      { key: 'principal-7', rama: 'principal', etapa: 'mergeado', x: 904, y: 40, activo: true },
      { key: 'run-1-0', rama: 'run-1', etapa: 'run', x: 304, y: 136, activo: false },
    ],
    aristas: [
      { key: 'a-principal-1', x1: 64, y1: 40, x2: 184, y2: 40, curva: false },
      { key: 'a-principal-2', x1: 184, y1: 40, x2: 304, y2: 40, curva: false },
      { key: 'a-principal-3', x1: 304, y1: 40, x2: 424, y2: 40, curva: false },
      { key: 'a-principal-4', x1: 424, y1: 40, x2: 544, y2: 40, curva: false },
      { key: 'a-principal-5', x1: 544, y1: 40, x2: 664, y2: 40, curva: false },
      { key: 'a-principal-6', x1: 664, y1: 40, x2: 784, y2: 40, curva: false },
      { key: 'a-principal-7', x1: 784, y1: 40, x2: 904, y2: 40, curva: false },
      { key: 'a-run-1-0', x1: 184, y1: 40, x2: 304, y2: 136, curva: true },
    ],
  }

  it('sin segundo argumento devuelve exactamente lo de hoy', () => {
    expect(proyeccion(grafoLayout(detalleStack))).toEqual(hoy)
  })

  it("con 'horizontal' devuelve lo mismo y no trae columnas de texto", () => {
    const h = grafoLayout(detalleStack, 'horizontal')
    expect(proyeccion(h)).toEqual(hoy)
    expect('columnas' in h).toBe(false)
  })
})
