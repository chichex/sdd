import type { Etapa, Nodo, Trabajo } from '@/lib/api'
import { ETAPA_INFO } from '../etapas'
import { NodoCard, StackCard } from './NodoCard'

const COL = new Map<Etapa, number>(ETAPA_INFO.map((e, i) => [e.id, i]))

/** Ramas con nodos, la viva primero; cada una es una fila de la tira. */
function filas(t: Trabajo): Array<{ id: string; porEtapa: Map<number, Nodo[]> }> {
  return t.ramas
    .filter((r) => r.nodos.length > 0)
    .map((r) => {
      const porEtapa = new Map<number, Nodo[]>()
      for (const n of r.nodos) {
        const col = COL.get(n.etapa) ?? 0
        porEtapa.set(col, [...(porEtapa.get(col) ?? []), n])
      }
      return { id: r.id, porEtapa }
    })
}

/**
 * Tira de un trabajo: atraviesa los siete carriles; sus ramas van una debajo
 * de la otra y los nodos de cada rama se unen con un conector de 1 px.
 */
export function Tira({ trabajo, now, onOpen }: { trabajo: Trabajo; now: Date; onOpen: () => void }) {
  return (
    <div role="group" aria-label={trabajo.titulo} data-tira className="grid grid-cols-[repeat(7,240px)] gap-x-4 gap-y-2 border-t border-border py-2">
      {filas(trabajo).map((fila, row) => {
        const cols = [...fila.porEtapa.keys()].sort((a, b) => a - b)
        const first = cols[0] ?? 0
        const last = cols[cols.length - 1] ?? 0
        return [
          last > first && (
            <div
              key={`${fila.id}-conector`}
              aria-hidden
              className="mx-[120px] h-px self-start bg-border-strong"
              style={{ gridColumn: `${first + 1} / ${last + 2}`, gridRow: row + 1, marginTop: 20 }}
            />
          ),
          ...cols.map((col) => {
            const nodos = fila.porEtapa.get(col) ?? []
            return (
              <div key={`${fila.id}-${col}`} className="px-2" style={{ gridColumn: col + 1, gridRow: row + 1 }}>
                {nodos.length > 1 ? (
                  <StackCard trabajo={trabajo} nodos={nodos} now={now} onOpen={onOpen} />
                ) : (
                  nodos[0] && <NodoCard trabajo={trabajo} nodo={nodos[0]} now={now} onOpen={onOpen} />
                )}
              </div>
            )
          }),
        ]
      })}
    </div>
  )
}
