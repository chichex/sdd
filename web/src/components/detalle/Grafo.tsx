import type { SVGProps } from 'react'
import { cn } from 'cn'
import type { Trabajo } from '@/lib/api'
import { fechaCorta } from '@/lib/time'
import { etapaInfo } from '../etapas'
import { aristaPath, grafoLayout, recortar, type GrafoColumnas, type GrafoNodo } from './grafo'

const FILL: Record<string, string> = {
  idea: 'fill-stage-idea',
  grill: 'fill-stage-grill',
  spec: 'fill-stage-spec',
  run: 'fill-stage-run',
  pr: 'fill-stage-pr',
  mergeado: 'fill-stage-merged',
  publicado: 'fill-stage-published',
}

interface NodoProps {
  g: GrafoNodo
  etiqueta: string
}

/** Horizontal: la etiqueta, el sub-estado y la fecha centrados debajo del círculo. */
function TextoDebajo({ g, etiqueta }: NodoProps) {
  return (
    <>
      <text x={g.x} y={g.y + 32} textAnchor="middle" className={cn('text-xs', g.activo ? 'fill-primary-text font-semibold' : 'fill-fg')}>
        {etiqueta}
      </text>
      <text x={g.x} y={g.y + 46} textAnchor="middle" className="fill-fg-secondary text-xs">
        {g.nodo.subEstado}
      </text>
      <text x={g.x} y={g.y + 60} textAnchor="middle" className="fill-fg-muted font-mono text-xs">
        {fechaCorta(g.nodo.ts)}
      </text>
    </>
  )
}

/** Vertical: etiqueta, sub-estado y fecha en una sola línea, en tres columnas a la derecha de los nodos. */
function TextoALaDerecha({ g, etiqueta, columnas }: NodoProps & { columnas: GrafoColumnas }) {
  const y = g.y + 4 // centra el texto de 12 px en el círculo
  return (
    <>
      <text x={columnas.etiqueta} y={y} className={cn('text-xs', g.activo ? 'fill-primary-text font-semibold' : 'fill-fg font-semibold')}>
        {recortar(etiqueta, columnas.maxEtiqueta)}
      </text>
      <text x={columnas.subEstado} y={y} className="fill-fg-secondary text-xs">
        {recortar(g.nodo.subEstado, columnas.maxSubEstado)}
      </text>
      <text x={columnas.fecha} y={y} className="fill-fg-muted font-mono text-xs">
        {fechaCorta(g.nodo.ts)}
      </text>
    </>
  )
}

/** Atributos del `svg`: el horizontal mide lo que su layout y scrollea; el vertical ocupa el ancho del contenedor y crece a lo alto. */
function atributosSvg(vertical: boolean, ancho: number, alto: number): SVGProps<SVGSVGElement> {
  return vertical
    ? { width: '100%', preserveAspectRatio: 'xMinYMin meet', className: 'block h-auto w-full' }
    : { width: ancho, height: alto, className: 'block' }
}

/**
 * Grafo SVG del trabajo: un nodo por etapa alcanzada, ramas cerradas como
 * bifurcaciones. En celular (`vertical`) es una columna de nodos sin scroll
 * horizontal; en desktop es una fila dentro de un contenedor con scroll.
 */
export function Grafo({ trabajo, vertical = false }: { trabajo: Trabajo; vertical?: boolean }) {
  const orientacion = vertical ? 'vertical' : 'horizontal'
  const { ancho, alto, nodos, aristas, columnas } = grafoLayout(trabajo, orientacion)
  const svg = (
    <svg role="img" aria-label="Grafo" viewBox={`0 0 ${ancho} ${alto}`} {...atributosSvg(vertical, ancho, alto)}>
      {aristas.map((a) => (
        <path key={a.key} d={aristaPath(a, orientacion)} className="fill-none stroke-border-strong" strokeWidth={1.5} />
      ))}
      {nodos.map((g) => {
        const etiqueta = g.nodo.capa ? `${etapaInfo(g.nodo.etapa).nombre} ${g.nodo.capa}` : etapaInfo(g.nodo.etapa).nombre
        const titulo = `${etiqueta} · ${g.nodo.subEstado}${g.nodo.refs.pr !== null ? ` · PR ${g.nodo.refs.pr}` : ''}${g.nodo.ruta ? ` · ${g.nodo.ruta}` : ''}`
        const cuerpo = (
          <g data-grafo-nodo data-rama={g.rama} data-activo={g.activo ? 'true' : undefined}>
            <title>{titulo}</title>
            <circle cx={g.x} cy={g.y} r={16} className={cn(FILL[g.nodo.etapa], g.activo ? 'stroke-primary-text' : 'stroke-border-strong')} strokeWidth={g.activo ? 3 : 1} />
            {columnas ? <TextoALaDerecha g={g} etiqueta={etiqueta} columnas={columnas} /> : <TextoDebajo g={g} etiqueta={etiqueta} />}
          </g>
        )
        return g.nodo.url ? (
          <a key={g.key} href={g.nodo.url} target="_blank" rel="noreferrer" aria-label={titulo}>
            {cuerpo}
          </a>
        ) : (
          <g key={g.key}>{cuerpo}</g>
        )
      })}
    </svg>
  )
  return vertical ? <div>{svg}</div> : <div className="overflow-x-auto">{svg}</div>
}
