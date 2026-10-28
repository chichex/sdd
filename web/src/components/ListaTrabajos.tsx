import type { Trabajo } from '@/lib/api'
import { edadCorta } from '@/lib/time'
import { AnclaIcono, EtapaDot, SinTrazaBadge, SubEstadoBadge } from './marcas'

/** Lista de celular: una fila de 44 px por trabajo, todo en una línea. */
export function ListaTrabajos({ trabajos, now, onOpen }: { trabajos: Trabajo[]; now: Date; onOpen: (t: Trabajo) => void }) {
  return (
    <ul aria-label="Trabajos" className="flex flex-col pb-[env(safe-area-inset-bottom)]">
      {trabajos.map((t) => (
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
  )
}
