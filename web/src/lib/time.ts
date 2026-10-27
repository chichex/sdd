// Fechas del visor: los `ts` llegan en UTC; la antigüedad es relativa y corta
// («3 h», «20 min», «1 d») y la fecha absoluta va en la hora local del navegador.

/** Antigüedad corta de ts respecto de now; «—» sin fecha. */
export function edadCorta(ts: string | null, now: Date): string {
  if (!ts) return '—'
  const then = Date.parse(ts)
  if (Number.isNaN(then)) return '—'
  const seconds = Math.floor((now.getTime() - then) / 1000)
  if (seconds < 60) return 'ahora'
  if (seconds < 3600) return `${Math.floor(seconds / 60)} min`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)} h`
  return `${Math.floor(seconds / 86400)} d`
}

/** Fecha y hora absolutas en la zona del navegador, para tooltips. */
export function fechaLarga(ts: string | null): string {
  if (!ts) return 'sin dato'
  const d = new Date(ts)
  return Number.isNaN(d.getTime()) ? ts : d.toLocaleString()
}

/** Fecha corta (día y mes) en la zona del navegador, para el grafo y el Gantt. */
export function fechaCorta(ts: string): string {
  const d = new Date(ts)
  return Number.isNaN(d.getTime()) ? ts : d.toLocaleDateString(undefined, { day: 'numeric', month: 'short' })
}
