import { describe, expect, it } from 'vitest'
import css from './index.css?raw'
import doc from '../DESIGN_SYSTEM.md?raw'

// Tokens de la tabla «### Color» del design system: nombre, claro y oscuro.
function colorTable(): Array<{ token: string; light: string; dark: string }> {
  const section = doc.split('### Color')[1]?.split('\n### ')[0] ?? ''
  const rows = [...section.matchAll(/^\| `(--color-[a-z-]+)` \| `([^`]+)` \| `([^`]+)` \|/gm)]
  return rows.map(([, token, light, dark]) => ({ token, light, dark }))
}

function block(source: string, start: string): string {
  const from = source.indexOf(start)
  if (from < 0) return ''
  const open = source.indexOf('{', from)
  let depth = 0
  for (let i = open; i < source.length; i += 1) {
    if (source[i] === '{') depth += 1
    if (source[i] === '}') depth -= 1
    if (depth === 0) return source.slice(open, i + 1)
  }
  return ''
}

describe('tokens de color', () => {
  const table = colorTable()
  const theme = block(css, '@theme {')
  // El bloque de tokens oscuros por sistema (no el de la variante `dark`).
  const mediaStart = css.indexOf('@media (prefers-color-scheme: dark) {\n  :root:not([data-theme="light"])')
  const darkMedia = mediaStart < 0 ? '' : block(css.slice(mediaStart), '@media')
  const darkManual = block(css, ':root[data-theme="dark"]')

  it('lee las 29 filas de la tabla del design system', () => {
    expect(table).toHaveLength(29)
  })

  it.each(table)('$token tiene su valor claro bajo @theme y el oscuro en los dos modos oscuros', ({ token, light, dark }) => {
    expect(theme).toContain(`${token}: ${light};`)
    expect(darkMedia).toContain(':root:not([data-theme="light"])')
    expect(darkMedia).toContain(`${token}: ${dark};`)
    expect(darkManual).toContain(`${token}: ${dark};`)
  })

  it('define familias, radios y sombras del design system', () => {
    for (const name of ['--font-sans', '--font-mono', '--radius-sm: 4px', '--radius-md: 6px', '--radius-lg: 8px', '--shadow-sm', '--shadow-md', '--shadow-lg']) {
      expect(theme).toContain(name)
    }
  })
})
