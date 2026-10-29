// Opciones de los controles de la cabecera, la Barra de filtros y el Sidebar.
import { Monitor, Moon, Sun } from 'lucide-react'
import type { Categoria } from '@/lib/api'
import type { Tema } from '@/lib/theme'
import type { SegmentedOption } from './Segmented'

/** Las cuatro categorías del control «Mostrar», en el orden en que se dibujan. */
export const CATEGORIA_INFO: ReadonlyArray<{ id: Categoria; nombre: string }> = [
  { id: 'en-curso', nombre: 'En curso' },
  { id: 'idea', nombre: 'Ideas' },
  { id: 'terminado', nombre: 'Terminados' },
  { id: 'sin-traza', nombre: 'Sin traza' },
]

export const TEMAS: ReadonlyArray<SegmentedOption<Tema>> = [
  { value: 'claro', label: 'Claro', Icon: Sun },
  { value: 'oscuro', label: 'Oscuro', Icon: Moon },
  { value: 'sistema', label: 'Sistema', Icon: Monitor },
]
