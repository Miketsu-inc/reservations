import {
  Calendar03Icon,
  Clock01Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import {
  LoginPage,
  redirectIfLoggedIn,
  validateLoginSearch,
} from "@reservations/components";
import { JABULANI_URL, TANGO_URL } from "@reservations/lib";
import { createFileRoute } from "@tanstack/react-router";

const FALLBACK_HREF = "/dashboard";

const highlights = [
  {
    icon: Calendar03Icon,
    text: "One shared calendar for your whole team",
  },
  {
    icon: Clock01Icon,
    text: "Take online bookings around the clock",
  },
  {
    icon: UserGroupIcon,
    text: "Keep your customers, services and products organized",
  },
];

// Accounts are created on tango, jabulani's /signup then sets up the business
// itself
const signupHref = `${TANGO_URL}/signup?${new URLSearchParams({
  redirect: `${JABULANI_URL}/signup`,
})}`;

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
      switchLink={{ href: `${TANGO_URL}/login`, label: "For customers" }}
      title="Log in to your business"
      description="Manage your calendar, team and bookings."
      forgotPasswordLink={{ href: `${TANGO_URL}/forgot-password` }}
      signupLink={{ href: signupHref }}
      headline="Everything your business needs to take bookings."
      highlights={highlights}
    />
  );
}
