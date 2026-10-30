import { useCallback, useEffect, useRef, useState } from 'react'
import { categoriasEnOrden, fetchWorkspace, type Categoria, type Workspace } from '@/lib/api'

/** Espera antes de reabrir /api/stream cuando el server cerró la conexión. */
export const RECONNECT_MS = 2000

/**
 * Lee /api/workspace con las categorías elegidas y lo vuelve a leer al
 * cambiarlas y en cada `snapshot` de /api/stream. Ante un error conserva el
 * último dato bueno y expone el mensaje; una respuesta vieja nunca pisa a una
 * más nueva.
 */
export function useWorkspace(mostrar: readonly Categoria[]) {
  // La selección llega como un array nuevo en cada render: su clave evita refetchear de más.
  const clave = mostrar.join(',')
  const [data, setData] = useState<Workspace | null>(null)
  const [error, setError] = useState<string | null>(null)
  const seq = useRef(0)

  const reload = useCallback(async (): Promise<void> => {
    const id = ++seq.current
    try {
      const ws = await fetchWorkspace(categoriasEnOrden(clave.split(',')))
      if (id !== seq.current) return
      setData(ws)
      setError(null)
    } catch (e) {
      if (id === seq.current) setError(e instanceof Error ? e.message : String(e))
    }
  }, [clave])

  const latest = useRef(reload)
  useEffect(() => {
    latest.current = reload
    void reload() // reload atrapa sus errores: la promesa nunca rechaza
  }, [reload])

  useEffect(() => {
    let es: EventSource | null = null
    let timer: number | undefined
    let disposed = false
    const connect = () => {
      es = new EventSource('/api/stream')
      es.addEventListener('snapshot', () => {
        void latest.current() // ídem: los errores quedan en `error`
      })
      es.onerror = () => {
        if (!disposed && es?.readyState === EventSource.CLOSED) timer = window.setTimeout(connect, RECONNECT_MS)
      }
    }
    connect()
    return () => {
      disposed = true
      es?.close()
      window.clearTimeout(timer)
    }
  }, [])

  return { data, error, reload }
}
