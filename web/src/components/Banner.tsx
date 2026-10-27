import type { ReactNode } from 'react'
import { cn } from 'cn'
import { OctagonX, TriangleAlert } from 'lucide-react'
import { Button } from '@/components/ui/button'

/** Barra bajo la cabecera con aviso o error y una acción siempre a la derecha. */
export function Banner({ tone, children, onRetry }: { tone: 'warning' | 'error'; children: ReactNode; onRetry: () => void }) {
  const Icon = tone === 'error' ? OctagonX : TriangleAlert
  return (
    <div
      role={tone === 'error' ? 'alert' : 'status'}
      aria-label={tone === 'error' ? 'Error' : 'Aviso'}
      className={cn(
        'flex items-center gap-3 border-b px-4 py-2 md:px-6',
        tone === 'error' ? 'bg-error-subtle text-error-text' : 'bg-warning-subtle text-warning-text',
      )}
    >
      <Icon aria-hidden className="size-4 shrink-0" />
      <p className="line-clamp-2 min-w-0 flex-1 text-sm">{children}</p>
      <Button variant="secondary" size="sm" onClick={onRetry}>
        Reintentar
      </Button>
    </div>
  )
}
