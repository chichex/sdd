// Piezas chicas compartidas: punto de etapa, badge de sub-estado, badge «sin
// traza», ícono de ancla y chip de referencia (DESIGN_SYSTEM.md).
import type { ReactNode } from 'react'
import { cn } from 'cn'
import { CircleDot, FileText, MessagesSquare, type LucideIcon } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import type { Etapa } from '@/lib/api'
import { badgeTone, etapaInfo } from './etapas'

export function EtapaDot({ etapa, className }: { etapa: Etapa; className?: string }) {
  return <span aria-hidden className={cn('inline-block size-2 shrink-0 rounded-full', etapaInfo(etapa).dot, className)} />
}

export function SubEstadoBadge({ subEstado }: { subEstado: string }) {
  return <Badge variant={badgeTone(subEstado)}>{subEstado}</Badge>
}

/** Marca de un trabajo sin eventos en events.jsonl (sinHistoria). */
export function SinTrazaBadge() {
  return <Badge variant="neutral">sin traza</Badge>
}

const ICONOS_ANCLA: Partial<Record<string, LucideIcon>> = { issue: CircleDot, grill: MessagesSquare, spec: FileText }

/** Ícono del tipo de ancla (issue, grill o spec), con aria-label y tooltip «ancla: <tipo>». */
export function AnclaIcono({ tipo }: { tipo: string }) {
  const Icono = ICONOS_ANCLA[tipo]
  if (!Icono) return null
  const texto = `ancla: ${tipo}`
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span role="img" aria-label={texto} className="inline-flex shrink-0 text-fg-muted">
          <Icono aria-hidden className="size-3.5" />
        </span>
      </TooltipTrigger>
      <TooltipContent>{texto}</TooltipContent>
    </Tooltip>
  )
}

const CHIP = 'inline-flex max-w-full truncate rounded-sm bg-bg-muted px-1.5 font-mono text-xs text-fg-secondary'

function Enlace({ href, className, children }: { href: string; className?: string; children: ReactNode }) {
  return (
    <a href={href} target="_blank" rel="noreferrer" className={cn(className, 'hover:text-fg')}>
      {children}
    </a>
  )
}

/**
 * Chip de referencia en mono. Con `ancla`, lleva delante el ícono del tipo de
 * ancla; el ícono queda fuera del link para no cambiar su nombre accesible.
 */
export function ChipRef({ children, href, ancla, className }: { children: ReactNode; href?: string | null; ancla?: string; className?: string }) {
  if (ancla) {
    return (
      <span className={cn(CHIP, 'items-center gap-1', className)}>
        <AnclaIcono tipo={ancla} />
        {href ? <Enlace href={href} className="min-w-0 truncate">{children}</Enlace> : <span className="min-w-0 truncate">{children}</span>}
      </span>
    )
  }
  if (href) return <Enlace href={href} className={cn(CHIP, className)}>{children}</Enlace>
  return <span className={cn(CHIP, className)}>{children}</span>
}
