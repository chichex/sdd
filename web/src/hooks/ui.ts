import { useCallback, useEffect, useState, useSyncExternalStore } from 'react'
import { toast } from 'sonner'
import { fetchHealth, postRefresh } from '@/lib/api'
import { NAVIGATE_EVENT } from '@/lib/router'
import { aplicarTema, guardarTema, leerTema, type Tema } from '@/lib/theme'
import { fechaLarga } from '@/lib/time'

function subscribeLocation(onChange: () => void): () => void {
  window.addEventListener('popstate', onChange)
  window.addEventListener(NAVIGATE_EVENT, onChange)
  return () => {
    window.removeEventListener('popstate', onChange)
    window.removeEventListener(NAVIGATE_EVENT, onChange)
  }
}

/** Ruta y query actuales; cambia con popstate y con navigate/writeFilters. */
export function useLocationPath(): { pathname: string; search: string } {
  const href = useSyncExternalStore(subscribeLocation, () => window.location.pathname + window.location.search)
  const q = href.indexOf('?')
  return q < 0 ? { pathname: href, search: '' } : { pathname: href.slice(0, q), search: href.slice(q) }
}

/** Coincidencia de una media query (false si el entorno no la soporta). */
export function useMediaQuery(query: string): boolean {
  return useSyncExternalStore(
    (onChange) => {
      const mql = window.matchMedia?.(query)
      mql?.addEventListener?.('change', onChange)
      return () => mql?.removeEventListener?.('change', onChange)
    },
    () => window.matchMedia?.(query).matches ?? false,
  )
}

/** La hora actual, refrescada cada intervalo para las antigüedades cortas. */
export function useNow(intervalMs: number): Date {
  const [now, setNow] = useState(() => new Date())
  useEffect(() => {
    const id = window.setInterval(() => setNow(new Date()), intervalMs)
    return () => window.clearInterval(id)
  }, [intervalMs])
  return now
}

export function useTema(): [Tema, (t: Tema) => void] {
  const [tema, setTema] = useState<Tema>(leerTema)
  useEffect(() => aplicarTema(tema), [tema])
  const elegir = useCallback((t: Tema) => {
    guardarTema(t)
    setTema(t)
  }, [])
  return [tema, elegir]
}

export function useHealth(): { version: string; root: string } | null {
  const [health, setHealth] = useState<{ version: string; root: string } | null>(null)
  useEffect(() => {
    let alive = true
    fetchHealth()
      .then((h) => alive && setHealth(h))
      .catch(() => alive && setHealth(null)) // sin health el pie del Sidebar queda vacío
    return () => {
      alive = false
    }
  }, [])
  return health
}

/** Refresh manual: POST /api/refresh, estado de espera, toast y relectura. */
export function useRefresh(onDone: () => Promise<void>) {
  const [state, setState] = useState<'idle' | 'busy' | 'error'>('idle')
  const refresh = useCallback(async (): Promise<void> => {
    setState('busy')
    try {
      const fetchedAt = await postRefresh()
      await onDone()
      setState('idle')
      toast.success(`Actualizado · ${fechaLarga(fetchedAt)}`)
    } catch (e) {
      setState('error')
      toast.error(`No se pudo actualizar: ${e instanceof Error ? e.message : String(e)}`, { duration: Infinity })
    }
  }, [onDone])
  return { refreshing: state === 'busy', failed: state === 'error', refresh }
}
