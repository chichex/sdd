import type { Trabajo } from '@/lib/api'
import type { Grupo, Vista } from '@/lib/filtrar'
import { Skeleton } from '@/components/ui/skeleton'
import { ETAPA_INFO } from '../etapas'
import { EncabezadoRepo, GrupoSinTraza } from '../Grupos'
import { EtapaDot } from '../marcas'
import { moverFoco } from './keyboard'
import { Tira } from './Tira'

const LANE = 240
const GAP = 16

function Fondos() {
  return (
    <div aria-hidden className="pointer-events-none absolute inset-y-0 left-0">
      {ETAPA_INFO.map((e, i) => (
        <div key={e.id} className="absolute inset-y-0 rounded-lg bg-bg-subtle" style={{ left: i * (LANE + GAP), width: LANE }} />
      ))}
    </div>
  )
}

interface GruposProps {
  grupos: Grupo[]
  encabezados: boolean
  now: Date
  onOpen: (t: Trabajo) => void
}

/** Tiras por repo; el encabezado de cada repo cruza los siete carriles. */
function Grupos({ grupos, encabezados, now, onOpen }: GruposProps) {
  return grupos.map((g) => (
    <div key={g.repo}>
      {encabezados && <EncabezadoRepo repo={g.repo} n={g.trabajos.length} className="relative mt-2" />}
      {g.trabajos.map((t) => (
        <Tira key={t.id} trabajo={t} now={now} onOpen={() => onOpen(t)} />
      ))}
    </div>
  ))
}

/**
 * Board de desktop: siete carriles fijos y una tira por trabajo, agrupadas por
 * repo; con «todos», los sin traza al final en un grupo colapsado. Los
 * carriles cuentan los nodos de todo lo visible. El scroll horizontal vive solo acá.
 */
export function Board({ vista, now, onOpen }: { vista: Vista; now: Date; onOpen: (t: Trabajo) => void }) {
  const counts = new Map(ETAPA_INFO.map((e) => [e.id, 0]))
  for (const t of vista.visibles) for (const r of t.ramas) for (const n of r.nodos) counts.set(n.etapa, (counts.get(n.etapa) ?? 0) + 1)
  return (
    <section aria-label="Tablero" className="overflow-x-auto px-6 py-4" onKeyDown={moverFoco}>
      <div className="relative w-max pb-2">
        <Fondos />
        <div className="relative grid grid-cols-[repeat(7,240px)] gap-x-4">
          {ETAPA_INFO.map((e) => (
            <div key={e.id} className="flex flex-col gap-1 px-3 py-2">
              <h2 className="flex items-center gap-2 text-xs font-semibold tracking-wide uppercase">
                <EtapaDot etapa={e.id} />
                {e.nombre} <span className="font-mono text-fg-muted tabular-nums">{counts.get(e.id)}</span>
              </h2>
              {counts.get(e.id) === 0 && <p className="text-xs text-fg-muted">Sin trabajos en {e.nombre}</p>}
            </div>
          ))}
        </div>
        <div className="relative">
          <Grupos grupos={vista.grupos} encabezados={vista.encabezados} now={now} onOpen={onOpen} />
          {vista.nSinTraza > 0 && (
            <GrupoSinTraza n={vista.nSinTraza} className="pt-2">
              <Grupos grupos={vista.sinTraza} encabezados={vista.encabezados} now={now} onOpen={onOpen} />
            </GrupoSinTraza>
          )}
        </div>
      </div>
    </section>
  )
}

/** Skeleton con la forma del board: siete carriles con tres nodos cada uno. */
export function BoardSkeleton() {
  return (
    <div role="status" aria-label="Cargando el tablero" aria-busy="true" className="overflow-hidden px-6 py-4">
      <div className="grid w-max grid-cols-[repeat(7,240px)] gap-x-4">
        {ETAPA_INFO.map((e) => (
          <div key={e.id} className="flex flex-col gap-2 rounded-lg bg-bg-subtle p-2">
            <Skeleton className="h-4 w-24" />
            {[0, 1, 2].map((i) => (
              <Skeleton key={i} className="h-[88px] w-full rounded-md" />
            ))}
          </div>
        ))}
      </div>
    </div>
  )
}
