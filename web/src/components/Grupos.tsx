// Piezas de agrupado del board y la lista: el encabezado de cada repo, el de
// cada etapa dentro de la lista y el grupo colapsado de los trabajos sin traza
// (DESIGN_SYSTEM.md, ajustes).
import { useId, useState, type ReactNode } from 'react'
import { cn } from 'cn'
import { ChevronDown, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { Etapa } from '@/lib/api'
import { etapaInfo } from './etapas'
import { EtapaDot } from './marcas'

/** Encabezado de un grupo de repo: `owner/repo` como h3 y el conteo a la derecha. */
export function EncabezadoRepo({ repo, n, className }: { repo: string; n: number; className?: string }) {
  return (
    <div className={cn('flex h-7 items-center justify-between gap-2 rounded-sm bg-bg-subtle px-3', className)}>
      <h3 className="truncate font-mono text-xs font-semibold tracking-wide text-fg-secondary uppercase">{repo}</h3>
      <span className="shrink-0 font-mono text-xs text-fg-muted tabular-nums">{n}</span>
    </div>
  )
}

/** Encabezado de una sección de la lista: punto de la etapa, su nombre como h4 y el conteo a la derecha. */
export function EncabezadoEtapa({ etapa, n }: { etapa: Etapa; n: number }) {
  return (
    <h4 className="flex h-7 items-center gap-2 px-4 text-xs font-semibold tracking-wide uppercase">
      <EtapaDot etapa={etapa} />
      {etapaInfo(etapa).nombre}{' '}
      <span className="ml-auto font-mono text-fg-muted tabular-nums">{n}</span>
    </h4>
  )
}

/** «Sin traza (N)»: cerrado por default; al abrirlo muestra su contenido debajo. */
export function GrupoSinTraza({ n, children, className }: { n: number; children: ReactNode; className?: string }) {
  const [abierto, setAbierto] = useState(false)
  const id = useId()
  const Chevron = abierto ? ChevronDown : ChevronRight
  return (
    <div className={cn('flex flex-col gap-2', className)}>
      <Button variant="ghost" aria-expanded={abierto} aria-controls={abierto ? id : undefined} onClick={() => setAbierto(!abierto)} className="self-start px-2">
        <Chevron aria-hidden />
        Sin traza ({n})
      </Button>
      {abierto && <div id={id}>{children}</div>}
    </div>
  )
}
