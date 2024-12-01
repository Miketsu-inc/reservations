import { createFileRoute } from '@tanstack/react-router'
import CalendarPage from '../../../pages/dashboard/CalendarPage'

export const Route = createFileRoute('/_authenticated/_sidepanel/calendar')({
  component: CalendarPage,
})
