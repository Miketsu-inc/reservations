import {
  CalendarCheckIn01Icon,
  Clock01Icon,
  Search01Icon,
} from "@hugeicons/core-free-icons";
import {
  LoginPage,
  redirectIfLoggedIn,
  validateLoginSearch,
} from "@reservations/components";
import { JABULANI_URL } from "@reservations/lib";
import { createFileRoute } from "@tanstack/react-router";

const FALLBACK_HREF = "/";

const highlights = [
  {
    icon: Search01Icon,
    text: "Discover local businesses and everything they offer",
  },
  {
    icon: Clock01Icon,
    text: "Book any time of the day, no phone calls needed",
  },
  {
    icon: CalendarCheckIn01Icon,
    text: "Keep all of your appointments in one place",
  },
];

export const Route = createFileRoute("/login")({
  validateSearch: validateLoginSearch,
  beforeLoad: redirectIfLoggedIn(FALLBACK_HREF),
  component: RouteComponent,
});

function RouteComponent() {
  const { redirect, error } = Route.useSearch();

  return (
    <LoginPage
      redirect={redirect}
      oauthError={error}
      fallbackHref={FALLBACK_HREF}
      homeLink={{ to: "/" }}
      switchLink={{ href: `${JABULANI_URL}/login`, label: "For businesses" }}
      title="Welcome back"
      description="Log in to book and manage your appointments."
      forgotPasswordLink={{ to: "/forgot-password" }}
      signupLink={{ to: "/signup", search: { redirect } }}
      headline="Your next appointment is only a few taps away."
      highlights={highlights}
    />
  );
}
