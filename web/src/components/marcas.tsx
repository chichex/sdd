// Piezas chicas compartidas: punto de etapa, badge de sub-estado y chip de
// referencia (DESIGN_SYSTEM.md).
import type { ReactNode } from 'react'
import { cn } from 'cn'
import { Badge } from '@/components/ui/badge'
import type { Etapa } from '@/lib/api'
import { badgeTone, etapaInfo } from './etapas'

export function EtapaDot({ etapa, className }: { etapa: Etapa; className?: string }) {
  return <span aria-hidden className={cn('inline-block size-2 shrink-0 rounded-full', etapaInfo(etapa).dot, className)} />
}

export function SubEstadoBadge({ subEstado }: { subEstado: string }) {
  return <Badge variant={badgeTone(subEstado)}>{subEstado}</Badge>
}

export function ChipRef({ children, href, className }: { children: ReactNode; href?: string | null; className?: string }) {
  const style = cn('inline-flex max-w-full truncate rounded-sm bg-bg-muted px-1.5 font-mono text-xs text-fg-secondary', className)
  if (href) {
    return (
      <a href={href} target="_blank" rel="noreferrer" className={cn(style, 'hover:text-fg')}>
        {children}
      </a>
    )
  }
  return <span className={style}>{children}</span>
}
