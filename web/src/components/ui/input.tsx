import * as React from "react"
import { cn } from "cn"

// shadcn/ui input adaptado a los tokens del design system: 32 px, radio md,
// borde --color-border, fondo --color-bg y placeholder --color-fg-muted. El
// foco usa el anillo global de :focus-visible (index.css), sin ring propio.
function Input({ className, type, ...props }: React.ComponentProps<"input">) {
  return (
    <input
      type={type}
      data-slot="input"
      className={cn(
        "h-8 w-full min-w-0 rounded-md border border-border bg-bg px-3 py-1 text-base text-fg transition-colors selection:bg-primary selection:text-primary-fg placeholder:text-fg-muted disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50",
        "aria-invalid:border-error-text",
        className
      )}
      {...props}
    />
  )
}

export { Input }
