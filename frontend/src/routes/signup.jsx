import { createFileRoute } from "@tanstack/react-router";
import SignUpPage from "../pages/onboarding/SignUpPage";

export const Route = createFileRoute("/signup")({
  component: SignUpPage,
});
