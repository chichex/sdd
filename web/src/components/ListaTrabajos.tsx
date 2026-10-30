import type { Trabajo } from '@/lib/api'
import { seccionarPorEtapa, type Grupo, type Vista } from '@/lib/filtrar'
import { edadCorta } from '@/lib/time'
import { EncabezadoEtapa, EncabezadoRepo, GrupoSinTraza } from './Grupos'
import { EtiquetaAncla, SinTrazaBadge, SubEstadoBadge } from './marcas'

interface GruposProps {
  grupos: Grupo[]
  encabezados: boolean
  now: Date
  onOpen: (t: Trabajo) => void
}

interface FilaProps {
  trabajo: Trabajo
  now: Date
  onOpen: (t: Trabajo) => void
}

/**
 * Fila de 44 px en una sola línea: la etapa la da el encabezado de su sección y solo el título se trunca.
 * El espacio tras la etiqueta no se dibuja (es un hijo flex) pero separa la palabra del título en el nombre accesible.
 */
function Fila({ trabajo: t, now, onOpen }: FilaProps) {
  return (
    <button
      type="button"
      onClick={() => onOpen(t)}
      className="flex h-11 w-full cursor-pointer items-center gap-2 px-4 text-left hover:bg-bg-muted"
    >
      <EtiquetaAncla tipo={t.ancla.tipo} />{' '}
      <span className="min-w-0 flex-1 truncate text-sm font-semibold">{t.titulo}</span>
      {t.sinHistoria && <SinTrazaBadge />}
      <SubEstadoBadge subEstado={t.subEstado} />
      <span className="shrink-0 font-mono text-xs text-fg-muted tabular-nums">{edadCorta(t.ultimaActividad, now)}</span>
    </button>
  )
}

interface GruposProps {
  grupos: Grupo[]
  encabezados: boolean
  now: Date
  onOpen: (t: Trabajo) => void
}

/** Un ítem por repo: su encabezado sticky y, bajo él, una sección por etapa con sus filas. */
function Grupos({ grupos, encabezados, now, onOpen }: GruposProps) {
  return grupos.map((g) => (
    <li key={g.repo}>
      {encabezados && <EncabezadoRepo repo={g.repo} n={g.trabajos.length} className="sticky top-0 z-10 mx-2 mt-2" />}
      <ul>
        {seccionarPorEtapa(g.trabajos).map((s) => (
          <li key={s.etapa}>
            <EncabezadoEtapa etapa={s.etapa} n={s.trabajos.length} />
            <ul>
              {s.trabajos.map((t) => (
                <li key={t.id} className="border-b border-border">
                  <Fila trabajo={t} now={now} onOpen={onOpen} />
                </li>
              ))}
            </ul>
          </li>
        ))}
      </ul>
    </li>
  ))
}

/** Lista de celular: grupos por repo, cada uno por etapa y, con la categoría sin traza elegida, esos trabajos al final en un grupo colapsado. */
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
