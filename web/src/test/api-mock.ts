// Dobles de la API del server para los tests del front: fetch con respuestas
// en cola por ruta y un EventSource controlable.
import { vi } from 'vitest'
import type { Workspace } from '@/lib/api'

type Reply = { status: number; body: unknown } | 'pending'

export interface ApiMock {
  calls: Array<{ method: string; url: string }>
  detalles: Map<string, unknown>
  queueWorkspace: (...replies: Array<Workspace | { status: number; body: unknown }>) => void
  holdRefresh: () => () => void
}

export function mockApi(first: Workspace | { status: number; body: unknown }): ApiMock {
  const calls: ApiMock['calls'] = []
  const detalles = new Map<string, unknown>()
  const workspaces: Reply[] = [toReply(first)]
  let last: Reply = toReply(first)
  let refreshGate: Promise<void> | null = null
  const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input)
    const method = init?.method ?? 'GET'
    calls.push({ method, url })
    if (url.startsWith('/api/workspace')) {
      last = workspaces.length > 1 ? (workspaces.shift() ?? last) : (workspaces[0] ?? last)
      if (last === 'pending') return new Promise<Response>(() => {})
      return json(last.status, last.body)
    }
    if (url === '/api/refresh' && method === 'POST') {
      if (refreshGate) await refreshGate
      return json(202, { fetchedAt: '2026-10-27T12:00:00Z' })
    }
    if (url === '/api/health') return json(200, { ok: true, version: '9.9.9', root: '/ws' })
    if (url.startsWith('/api/trabajos/')) {
      const detalle = detalles.get(decodeURIComponent(url.slice('/api/trabajos/'.length)))
      return detalle === undefined ? json(404, { error: 'trabajo no encontrado' }) : json(200, detalle)
    }
    return json(404, { error: 'no encontrado' })
  })
  vi.stubGlobal('fetch', fetchMock)
  return {
    calls,
    detalles,
    queueWorkspace: (...replies) => {
      workspaces.splice(0, workspaces.length, ...replies.map(toReply))
    },
    holdRefresh: () => {
      let release = () => {}
      refreshGate = new Promise<void>((resolve) => {
        release = () => {
          refreshGate = null
          resolve()
        }
      })
      return release
    },
  }
}

function toReply(r: Workspace | { status: number; body: unknown }): Reply {
  return 'status' in r && 'body' in r ? (r as { status: number; body: unknown }) : { status: 200, body: r }
}

function json(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

export class FakeEventSource {
  static instances: FakeEventSource[] = []
  static readonly CONNECTING = 0
  static readonly OPEN = 1
  static readonly CLOSED = 2
  readyState = FakeEventSource.OPEN
  onerror: ((e: Event) => void) | null = null
  private listeners = new Map<string, Array<(e: MessageEvent) => void>>()
  url: string
  constructor(url: string) {
    this.url = url
    FakeEventSource.instances.push(this)
  }
  addEventListener(type: string, fn: (e: MessageEvent) => void): void {
    this.listeners.set(type, [...(this.listeners.get(type) ?? []), fn])
  }
  removeEventListener(): void {}
  close(): void {
    this.readyState = FakeEventSource.CLOSED
  }
  emit(type: string, data: unknown): void {
    for (const fn of this.listeners.get(type) ?? []) fn(new MessageEvent(type, { data: JSON.stringify(data) }))
  }
  fail(): void {
    this.readyState = FakeEventSource.CLOSED
    this.onerror?.(new Event('error'))
  }
}

export function mockEventSource(): typeof FakeEventSource {
  FakeEventSource.instances = []
  vi.stubGlobal('EventSource', FakeEventSource)
  return FakeEventSource
}
