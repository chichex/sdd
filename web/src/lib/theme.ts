// Tema claro, oscuro o del sistema. Se aplica con data-theme en <html> (el
// CSS de tokens lo lee) y se recuerda en localStorage por navegador.
export type Tema = 'claro' | 'oscuro' | 'sistema'

const KEY = 'sdd-tema'

export function aplicarTema(t: Tema): void {
  const root = document.documentElement
  if (t === 'sistema') root.removeAttribute('data-theme')
  else root.dataset.theme = t === 'oscuro' ? 'dark' : 'light'
}

export function guardarTema(t: Tema): void {
  try {
    window.localStorage.setItem(KEY, t)
  } catch {
    // sin storage (modo privado o bloqueado): el tema dura lo que la pestaña
  }
}

export function leerTema(): Tema {
  try {
    const v = window.localStorage.getItem(KEY)
    return v === 'claro' || v === 'oscuro' ? v : 'sistema'
  } catch {
    return 'sistema'
  }
}
