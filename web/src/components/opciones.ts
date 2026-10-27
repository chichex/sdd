// Opciones de los controles segmentados de la cabecera y del Sidebar.
import { Monitor, Moon, Sun } from 'lucide-react'
import type { Filtro } from '@/lib/router'
import type { Tema } from '@/lib/theme'
import type { SegmentedOption } from './Segmented'

export const FILTROS: ReadonlyArray<SegmentedOption<Filtro>> = [
  { value: 'activos', label: 'Activos' },
  { value: 'todos', label: 'Todos' },
]

export const TEMAS: ReadonlyArray<SegmentedOption<Tema>> = [
  { value: 'claro', label: 'Claro', Icon: Sun },
  { value: 'oscuro', label: 'Oscuro', Icon: Moon },
  { value: 'sistema', label: 'Sistema', Icon: Monitor },
]
