import { cn } from 'cn'
import { RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { edadCorta, fechaLarga } from '@/lib/time'

/**
 * Refresh de la cabecera: la etiqueta es la antigüedad de fetchedAt («2 min»);
 * gira mientras espera y pasa a error si la última actualización falló.
 */
export function Refresh({ fetchedAt, now, refreshing, failed, onRefresh }: {
  fetchedAt: string | null
  now: Date
  refreshing: boolean
  failed: boolean
  onRefresh: () => void
}) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant="ghost"
          aria-label="Actualizar"
          aria-busy={refreshing}
          onClick={onRefresh}
          className={cn('gap-1.5 px-2', failed && 'text-error-text hover:text-error-text')}
        >
          <RefreshCw aria-hidden className={cn('size-4', refreshing && 'animate-spin motion-reduce:animate-none motion-reduce:opacity-50')} />
          <span className="font-mono text-xs tabular-nums">{refreshing ? '…' : edadCorta(fetchedAt, now)}</span>
        </Button>
      </TooltipTrigger>
      <TooltipContent>Actualizado {fechaLarga(fetchedAt)}</TooltipContent>
    </Tooltip>
  )
}
