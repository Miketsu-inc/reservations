import {
  AlertCircleIcon,
  Calendar03Icon,
  Clock01Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import { FacebookIcon, GoogleIcon } from "@reservations/assets";
import { Button, Icon, Input, ServerError } from "@reservations/components";
import {
  getSafeRedirectUrl,
  JABULANI_URL,
  MAX_INPUT_LENGTH,
  TANGO_URL,
  meQueryOptions,
} from "@reservations/lib";
import { createFileRoute, redirect } from "@tanstack/react-router";
import { useState } from "react";

const EMAIL_PATTERN = /^\S+@\S+\.\S+$/;

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

const signupHref = `${TANGO_URL}/signup?${new URLSearchParams({
  redirect: `${JABULANI_URL}/signup`,
})}`;

export const Route = createFileRoute("/login")({
  validateSearch: (search) => ({
    redirect: typeof search.redirect === "string" ? search.redirect : undefined,
  }),
  beforeLoad: async ({ context: { queryClient }, search }) => {
    try {
      // fetchQuery instead of ensureQueryData so a stale "me" left over from
      // a logout doesn't bounce the user back into the app
      await queryClient.fetchQuery(meQueryOptions());
    } catch {
      return;
    }

    throw redirect({ href: getRedirectHref(search.redirect), replace: true });
  },
  component: LoginPage,
});

function getRedirectHref(redirectParam) {
  const redirectUrl = getSafeRedirectUrl(redirectParam);
  if (!redirectUrl) return "/dashboard";

  // Absolute hrefs make the router do a full page load, only needed when
  // leaving for the other app
  const url = new URL(redirectUrl);
  return url.origin === window.location.origin
    ? url.pathname + url.search + url.hash
    : url.href;
}

function validate({ email, password }) {
  const errors = {};

  if (!email.trim()) {
    errors.email = "Enter your email address";
  } else if (!EMAIL_PATTERN.test(email.trim())) {
    errors.email = "Enter a valid email address, like name@example.com";
  }

  if (!password) {
    errors.password = "Enter your password";
  }

  return errors;
}

function LoginPage() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const [values, setValues] = useState({ email: "", password: "" });
  const [errors, setErrors] = useState({});
  const [serverError, setServerError] = useState();
  const [isLoading, setIsLoading] = useState(false);
  const [isCapsLockOn, setIsCapsLockOn] = useState(false);

  function handleInputData({ name, value }) {
    const nextValues = { ...values, [name]: value };
    setValues(nextValues);

    // Only revalidate fields that are already showing an error, so the user
    // isn't yelled at while typing for the first time
    if (errors[name]) {
      setErrors((prev) => ({ ...prev, [name]: validate(nextValues)[name] }));
    }
  }

  function handleCapsLock(e) {
    setIsCapsLockOn(e.getModifierState?.("CapsLock") ?? false);
  }

  async function handleSubmit(e) {
    e.preventDefault();
    if (isLoading) return;

    const nextErrors = validate(values);
    setErrors(nextErrors);

    const firstInvalid = ["email", "password"].find((key) => nextErrors[key]);
    if (firstInvalid) {
      document.getElementById(`login-${firstInvalid}`)?.focus();
      return;
    }

    setIsLoading(true);
    setServerError();

    try {
      const response = await fetch("/api/v1/auth/login", {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          email: values.email.trim(),
          password: values.password,
        }),
      });

      if (!response.ok) {
        const result = await response.json();
        setServerError(
          result.error?.code === "incorrect_email_or_password"
            ? "The email or password you entered is incorrect."
            : (result.error?.message ?? "Something went wrong, try again.")
        );
        setIsLoading(false);
        return;
      }
    } catch {
      setServerError(
        "Couldn't reach the server. Check your connection and try again."
      );
      setIsLoading(false);
      return;
    }

    // Loading isn't reset on success, a redirect to the other app is a full
    // page load and the button shouldn't become clickable in the meantime
    navigate({ href: getRedirectHref(search.redirect), replace: true });
  }

  return (
    <div className="bg-bg flex min-h-dvh">
      <main className="flex flex-1 flex-col px-6 py-6 sm:px-10 sm:py-8">
        <div
          className="mx-auto flex w-full max-w-sm flex-1 flex-col justify-center
            py-12"
        >
          <h1 className="text-3xl font-semibold tracking-tight">
            Log in to your business
          </h1>
          <p className="text-fg-muted mt-2">
            Manage your calendar, team and bookings.
          </p>

          <div className="mt-8 flex flex-col gap-3">
            <a
              className="border-border-strong bg-bg-input text-fg hover:bg-hover
                focus-visible:outline-primary shadow-card flex items-center
                justify-center gap-2.5 rounded-lg border py-2.5 text-sm
                font-medium transition-colors focus-visible:outline-2
                focus-visible:outline-offset-2"
              href="/api/v1/auth/oauth/google"
            >
              <GoogleIcon styles="fill-fg size-4.5" />
              Continue with Google
            </a>
            <a
              className="border-border-strong bg-bg-input text-fg hover:bg-hover
                focus-visible:outline-primary shadow-card flex items-center
                justify-center gap-2.5 rounded-lg border py-2.5 text-sm
                font-medium transition-colors focus-visible:outline-2
                focus-visible:outline-offset-2"
              href="/api/v1/auth/oauth/facebook"
            >
              <FacebookIcon styles="fill-fg size-4.5" />
              Continue with Facebook
            </a>
          </div>

          <div className="my-6 flex items-center gap-3">
            <hr className="border-border flex-1" />
            <span className="text-fg-muted text-xs">or with email</span>
            <hr className="border-border flex-1" />
          </div>

          <form
            onSubmit={handleSubmit}
            noValidate
            className="flex flex-col gap-5"
            aria-label="Log in with email"
          >
            <ServerError error={serverError} />

            <div>
              <Input
                id="login-email"
                name="email"
                type="email"
                labelText="Email"
                autoComplete="username"
                inputMode="email"
                autoCapitalize="none"
                spellCheck={false}
                maxLength={MAX_INPUT_LENGTH}
                required={false}
                aria-required="true"
                aria-invalid={!!errors.email}
                aria-describedby={
                  errors.email ? "login-email-error" : undefined
                }
                value={values.email}
                inputData={handleInputData}
                styles="py-2.5"
              />
              <FieldError id="login-email-error" message={errors.email} />
            </div>

            <div>
              <Input
                id="login-password"
                name="password"
                type="password"
                labelText="Password"
                autoComplete="current-password"
                maxLength={MAX_INPUT_LENGTH}
                required={false}
                aria-required="true"
                aria-invalid={!!errors.password}
                aria-describedby={
                  [
                    errors.password && "login-password-error",
                    isCapsLockOn && "login-capslock",
                  ]
                    .filter(Boolean)
                    .join(" ") || undefined
                }
                value={values.password}
                inputData={handleInputData}
                onKeyDown={handleCapsLock}
                onKeyUp={handleCapsLock}
                onBlur={() => setIsCapsLockOn(false)}
                styles="py-2.5"
              />
              <FieldError id="login-password-error" message={errors.password} />
              {isCapsLockOn && (
                <p
                  id="login-capslock"
                  className="text-fg-warning mt-1.5 text-sm"
                >
                  Caps Lock is on
                </p>
              )}
              <div className="mt-2 flex justify-end">
                <a
                  href={`${TANGO_URL}/forgot-password`}
                  className="text-fg-primary focus-visible:outline-primary
                    rounded text-sm font-medium hover:underline
                    focus-visible:outline-2 focus-visible:outline-offset-2"
                >
                  Forgot password?
                </a>
              </div>
            </div>

            <Button
              variant="primary"
              type="submit"
              styles="py-2.5 font-medium"
              buttonText={isLoading ? "Logging in" : "Log in"}
              isLoading={isLoading}
            />
          </form>

          <p className="text-fg-muted mt-8 text-center text-sm">
            Don't have an account?{" "}
            {/* Accounts are created on tango, jabulani's /signup then sets up
                the business itself */}
            <a
              href={signupHref}
              className="text-fg-primary focus-visible:outline-primary rounded
                font-medium hover:underline focus-visible:outline-2
                focus-visible:outline-offset-2"
            >
              Sign up
            </a>
          </p>
        </div>

        <p className="text-fg-muted mx-auto text-center text-sm">
          Looking to book an appointment?{" "}
          <a
            href={`${TANGO_URL}/login`}
            className="text-fg-primary focus-visible:outline-primary rounded
              font-medium hover:underline focus-visible:outline-2
              focus-visible:outline-offset-2"
          >
            Customer login
          </a>
        </p>
      </main>

      <aside
        className="bg-primary text-fg-on-primary relative hidden overflow-hidden
          lg:flex lg:w-[44%] lg:flex-col lg:justify-end lg:p-14 xl:p-20"
      >
        <div
          aria-hidden="true"
          className="absolute -top-32 -right-32 size-[28rem] rounded-full
            bg-white/10"
        />
        <div
          aria-hidden="true"
          className="absolute top-40 -right-10 size-56 rounded-full bg-white/10"
        />
        <div className="relative max-w-md">
          <p className="text-4xl leading-tight font-semibold tracking-tight">
            Everything your business needs to take bookings.
          </p>
          <ul className="mt-10 flex flex-col gap-5">
            {highlights.map((highlight) => (
              <li key={highlight.text} className="flex items-center gap-4">
                <span
                  className="flex size-10 shrink-0 items-center justify-center
                    rounded-lg bg-white/15"
                >
                  <Icon
                    icon={highlight.icon}
                    styles="size-5"
                    aria-hidden="true"
                  />
                </span>
                <span className="opacity-90">{highlight.text}</span>
              </li>
            ))}
          </ul>
        </div>
      </aside>
    </div>
  );
}

function FieldError({ id, message }) {
  if (!message) return null;

  return (
    <p
      id={id}
      className="text-fg-danger mt-1.5 flex items-center gap-1.5 text-sm"
    >
      <Icon
        icon={AlertCircleIcon}
        styles="size-4 shrink-0"
        aria-hidden="true"
      />
      {message}
    </p>
  );
}
