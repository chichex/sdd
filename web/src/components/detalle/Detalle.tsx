import type { ReactNode } from 'react'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { Skeleton } from '@/components/ui/skeleton'
import { useDetalle } from '@/hooks/useDetalle'
import type { Detalle as DetalleT } from '@/lib/api'
import { EmptyState } from '../EmptyState'
import { refCorta } from '../etapas'
import { ChipRef } from '../marcas'
import { Gantt } from './Gantt'
import { Grafo } from './Grafo'

function Links({ d }: { d: DetalleT }) {
  return (
    <section className="flex flex-col gap-2">
      <h3 className="text-xs font-semibold tracking-wide text-fg-secondary uppercase">Links</h3>
      <ul className="flex flex-wrap gap-2">
        {d.links.prs.map((pr) => (
          <li key={pr.numero}>
            <ChipRef href={pr.url}>{`PR ${pr.numero}`}</ChipRef>
          </li>
        ))}
      </ul>
      <ul className="flex flex-col gap-1">
        {[...d.links.grills, ...d.links.specs].map((ruta) => (
          <li key={ruta}>
            <code className="block truncate text-xs text-fg-secondary select-all">{ruta}</code>
          </li>
        ))}
      </ul>
    </section>
  )
}

function Contenido({ d, hoy }: { d: DetalleT; hoy: Date }) {
  return (
    <div className="flex min-w-0 flex-col gap-6">
      <section className="flex flex-col gap-2">
        <h3 className="text-xs font-semibold tracking-wide text-fg-secondary uppercase">Grafo</h3>
        <Grafo trabajo={d} />
      </section>
      <section className="flex flex-col gap-2">
        <h3 className="text-xs font-semibold tracking-wide text-fg-secondary uppercase">Gantt</h3>
        <Gantt detalle={d} hoy={hoy} />
      </section>
      <Links d={d} />
    </div>
  )
}

/**
 * Detalle de un trabajo en /trabajos/<id>: Dialog en desktop, Sheet desde
 * abajo en celular. Cerrar (X, Esc u overlay) vuelve al board con los filtros.
 */
export function Detalle({ id, desktop, version, hoy, onClose }: { id: string; desktop: boolean; version: string | null; hoy: Date; onClose: () => void }) {
  const estado = useDetalle(id, version)
  const d = estado.status === 'ok' ? estado.detalle : null
  const titulo = d ? d.titulo : estado.status === 'cargando' ? 'Cargando…' : 'Detalle'
  const chips: ReactNode = d && (
    <span className="flex flex-wrap items-center gap-2">
      <ChipRef>{d.repo}</ChipRef>
      {d.ancla.tipo === 'issue' ? <ChipRef href={d.links.issue}>{d.ancla.ref}</ChipRef> : <ChipRef>{refCorta(d.ancla.ref)}</ChipRef>}
    </span>
  )
  let cuerpo: ReactNode
  if (d) cuerpo = <Contenido d={d} hoy={hoy} />
  else if (estado.status === 'no-encontrado') cuerpo = <EmptyState>Trabajo no encontrado</EmptyState>
  else if (estado.status === 'error') cuerpo = <EmptyState>No se pudo leer el trabajo ({estado.error}).</EmptyState>
  else cuerpo = <Skeleton className="h-48 w-full" />
  const cerrar = (open: boolean) => {
    if (!open) onClose()
  }
  if (desktop) {
    return (
      <Dialog open onOpenChange={cerrar}>
        <DialogContent className="max-h-[90dvh] overflow-y-auto sm:max-w-4xl">
          <DialogHeader>
            <DialogTitle className="truncate pr-8">{titulo}</DialogTitle>
            <DialogDescription asChild>
              <div>{chips}</div>
            </DialogDescription>
          </DialogHeader>
          {cuerpo}
        </DialogContent>
      </Dialog>
    )
  }
  return (
    <Sheet open onOpenChange={cerrar}>
      <SheetContent side="bottom" className="max-h-[90dvh] overflow-y-auto rounded-t-lg px-4 pb-[max(1rem,env(safe-area-inset-bottom))]">
        <SheetHeader className="px-0">
          <SheetTitle className="truncate pr-10">{titulo}</SheetTitle>
          <SheetDescription asChild>
            <div>{chips}</div>
          </SheetDescription>
        </SheetHeader>
        {cuerpo}
      </SheetContent>
    </Sheet>
  )
}
