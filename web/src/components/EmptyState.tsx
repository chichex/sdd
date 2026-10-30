import type { ReactNode } from 'react'
import { FolderSearch } from 'lucide-react'

/** Estado vacío: ícono, texto y, si aplica, un comando para copiar o una acción. */
export function EmptyState({ children, comando, accion }: { children: ReactNode; comando?: string; accion?: ReactNode }) {
  return (
    <div className="flex flex-col items-center gap-3 px-4 py-16 text-center">
      <FolderSearch aria-hidden className="size-8 text-fg-muted" />
      <p className="max-w-md text-sm text-fg-secondary">{children}</p>
      {comando && <code className="rounded-sm bg-bg-muted px-2 py-1 text-xs select-all">{comando}</code>}
      {accion}
    </div>
  )
}
