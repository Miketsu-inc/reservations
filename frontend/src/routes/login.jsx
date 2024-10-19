import { createFileRoute } from '@tanstack/react-router'
import LoginPage from "../pages/onboarding/LoginPage"

export const Route = createFileRoute('/login')({
  component: LoginPage,
})
