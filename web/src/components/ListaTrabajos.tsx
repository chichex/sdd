import type { Trabajo } from '@/lib/api'
import type { Grupo, Vista } from '@/lib/filtrar'
import { edadCorta } from '@/lib/time'
import { EncabezadoRepo, GrupoSinTraza } from './Grupos'
import { AnclaIcono, EtapaDot, SinTrazaBadge, SubEstadoBadge } from './marcas'

interface GruposProps {
  grupos: Grupo[]
  encabezados: boolean
  now: Date
  onOpen: (t: Trabajo) => void
}

/** Un ítem por repo: su encabezado sticky y sus filas de 44 px, todo en una línea. */
function Grupos({ grupos, encabezados, now, onOpen }: GruposProps) {
  return grupos.map((g) => (
    <li key={g.repo}>
      {encabezados && <EncabezadoRepo repo={g.repo} n={g.trabajos.length} className="sticky top-0 z-10 mx-2 mt-2" />}
      <ul>
        {g.trabajos.map((t) => (
          <li key={t.id} className="border-b border-border">
            <button
              type="button"
              onClick={() => onOpen(t)}
              className="flex h-11 w-full cursor-pointer items-center gap-2 px-4 text-left hover:bg-bg-muted"
            >
              <EtapaDot etapa={t.etapa} />
              <AnclaIcono tipo={t.ancla.tipo} />
              <span className="min-w-0 flex-1 truncate text-sm font-semibold">{t.titulo}</span>
              {t.sinHistoria && <SinTrazaBadge />}
              <SubEstadoBadge subEstado={t.subEstado} />
              <span className="shrink-0 font-mono text-xs text-fg-muted tabular-nums">{edadCorta(t.ultimaActividad, now)}</span>
            </button>
          </li>
        ))}
      </ul>
    </li>
  ))
}

/** Lista de celular: grupos por repo y, con «todos», los sin traza al final en un grupo colapsado. */
export function ListaTrabajos({ vista, now, onOpen }: { vista: Vista; now: Date; onOpen: (t: Trabajo) => void }) {
  return (
    <ul aria-label="Trabajos" className="flex flex-col pb-[env(safe-area-inset-bottom)]">
      <Grupos grupos={vista.grupos} encabezados={vista.encabezados} now={now} onOpen={onOpen} />
      {vista.nSinTraza > 0 && (
        <li>
          <GrupoSinTraza n={vista.nSinTraza} className="px-2 pt-2">
            <ul>
              <Grupos grupos={vista.sinTraza} encabezados={vista.encabezados} now={now} onOpen={onOpen} />
            </ul>
          </GrupoSinTraza>
        </li>
      )}
    </ul>
  )
}
