import type { ComponentType } from 'react'
import { cn } from 'cn'

export interface SegmentedOption<T extends string> {
  value: T
  label: string
  Icon?: ComponentType<{ className?: string; 'aria-hidden'?: boolean }>
}

/** Control segmentado accesible como radiogroup (activos · todos, tema). */
export function Segmented<T extends string>({
  label,
  value,
  options,
  onChange,
  className,
}: {
  label: string
  value: T
  options: ReadonlyArray<SegmentedOption<T>>
  onChange: (v: T) => void
  className?: string
}) {
  return (
    <div role="radiogroup" aria-label={label} className={cn('inline-flex h-7 rounded-md border border-border p-px pointer-coarse:h-11', className)}>
      {options.map((o) => (
        <button
          key={o.value}
          type="button"
          role="radio"
          aria-checked={o.value === value}
          onClick={() => onChange(o.value)}
          className={cn(
            'inline-flex flex-1 cursor-pointer items-center justify-center gap-1.5 rounded-sm px-2 text-sm font-medium whitespace-nowrap text-fg-secondary hover:text-fg',
            o.value === value && 'bg-primary-subtle text-primary-text hover:text-primary-text',
          )}
        >
          {o.Icon && <o.Icon aria-hidden className="size-4" />}
          {o.label}
        </button>
      ))}
    </div>
  )
}
