import { describe, expect, it } from 'vitest'
import { detalleStack } from '@/test/fixtures'
import { GANTT_LABEL, ganttLayout } from './gantt'

const hoy = new Date('2026-10-27T12:00:00Z')

describe('ganttLayout', () => {
  const layout = ganttLayout(detalleStack, 1000, hoy)

  it('una fila por barra, ramas cerradas como filas paralelas etiquetadas', () => {
    expect(layout.barras.map((b) => [b.rama, b.etapa])).toEqual([
      ['principal', 'run'],
      ['principal', 'pr'],
      ['principal', 'mergeado'],
      ['run-1', 'run'],
    ])
    expect(new Set(layout.barras.map((b) => b.y)).size).toBe(4)
    expect(layout.barras[3].etiqueta).toContain('fallado (interrupted)')
  })

  it('posiciona por fecha: el inicio más temprano en el borde y las barras en orden temporal', () => {
    const [run, pr, merged, fallado] = layout.barras
    expect(fallado.x).toBe(GANTT_LABEL)
    expect(run.x).toBeGreaterThan(fallado.x)
    expect(pr.x).toBeCloseTo(run.x + run.w, 5)
    expect(merged.x).toBeCloseTo(pr.x + pr.w, 5)
  })

  it('una barra abierta llega hasta hoy y la línea de hoy cierra el eje', () => {
    const merged = layout.barras[2]
    expect(merged.abierta).toBe(true)
    expect(merged.x + merged.w).toBeCloseTo(layout.hoyX, 5)
    expect(layout.hoyX).toBeLessThanOrEqual(layout.ancho)
    expect(layout.ticks.length).toBeGreaterThan(1)
  })

  it('un trabajo sin historia no tiene barras', () => {
    expect(ganttLayout({ ...detalleStack, sinHistoria: true, gantt: [] }, 1000, hoy).barras).toEqual([])
  })
})
