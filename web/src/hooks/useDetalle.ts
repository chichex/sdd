import { useEffect, useState } from 'react'
import { ApiError, fetchTrabajo, type Detalle } from '@/lib/api'

export type DetalleEstado =
  | { status: 'cargando' }
  | { status: 'ok'; detalle: Detalle }
  | { status: 'no-encontrado' }
  | { status: 'error'; error: string }

/** Lee /api/trabajos/{id}; se relee cuando cambia `version` (un snapshot nuevo). */
export function useDetalle(id: string, version: string | null): DetalleEstado {
  const [estado, setEstado] = useState<DetalleEstado>({ status: 'cargando' })
  useEffect(() => {
    let vigente = true
    fetchTrabajo(id)
      .then((detalle) => vigente && setEstado({ status: 'ok', detalle }))
      .catch((e: unknown) => {
        if (!vigente) return
        if (e instanceof ApiError && e.status === 404) setEstado({ status: 'no-encontrado' })
        else setEstado({ status: 'error', error: e instanceof Error ? e.message : String(e) })
      })
    return () => {
      vigente = false
    }
  }, [id, version])
  return estado
}
