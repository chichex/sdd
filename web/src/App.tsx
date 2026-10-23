import { useEffect, useState } from 'react'

type Health = { ok: boolean; version: string; root: string }

export default function App() {
  const [health, setHealth] = useState<Health | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetch('/api/health')
      .then((r) => (r.ok ? r.json() : Promise.reject(new Error(`HTTP ${r.status}`))))
      .then(setHealth)
      .catch((e: Error) => setError(e.message))
  }, [])

  return (
    <main className="min-h-dvh bg-neutral-950 px-4 py-8 text-neutral-100">
      <h1 className="text-2xl font-semibold">sdd</h1>
      <p className="mt-1 text-neutral-400">Visor del flujo SDD. Esqueleto: el tablero llega con la spec.</p>
      {health && (
        <p className="mt-4 font-mono text-sm">
          server {health.version} · escaneando {health.root}
        </p>
      )}
      {error && <p className="mt-4 text-sm text-red-400">sin server: {error}</p>}
    </main>
  )
}
