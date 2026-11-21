import { ArrowRight01Icon } from "@hugeicons/core-free-icons";
import { FacebookIcon, GoogleIcon } from "@reservations/assets";
import {
  getSafeRedirectUrl,
  MAX_INPUT_LENGTH,
  meQueryOptions,
} from "@reservations/lib";
import { Link, redirect, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import Button from "./Button.jsx";
import Icon from "./Icon.jsx";
import Input from "./Input.jsx";
import ServerError from "./ServerError.jsx";

const EMAIL_PATTERN = /^\S+@\S+\.\S+$/;

const OAUTH_ERRORS = {
  oauth_account_already_exists:
    "An account with this email already exists. Log in with your email and password instead.",
  oauth_email_unavailable:
    "We couldn't get an email address from that account. Try another way to log in.",
};

function getOauthErrorMessage(code) {
  if (!code) return undefined;

  return (
    OAUTH_ERRORS[code] ?? "Something went wrong while logging in, try again."
  );
}

export function validateLoginSearch(search) {
  return {
    redirect: typeof search.redirect === "string" ? search.redirect : undefined,
    error: typeof search.error === "string" ? search.error : undefined,
  };
}

export function redirectIfLoggedIn(fallbackHref) {
  return async ({ context: { queryClient }, search }) => {
    try {
      // fetchQuery instead of ensureQueryData so a stale "me" left over from
      // a logout doesn't bounce the user back into the app
      await queryClient.fetchQuery(meQueryOptions());
    } catch {
      return;
    }

    throw redirect({
      href: getRedirectHref(search.redirect, fallbackHref),
      replace: true,
    });
  };
}

function getRedirectHref(redirectParam, fallbackHref) {
  const redirectUrl = getSafeRedirectUrl(redirectParam);
  if (!redirectUrl) return fallbackHref;

  const url = new URL(redirectUrl);
  return url.origin === window.location.origin
    ? url.pathname + url.search + url.hash
    : url.href;
}

function getOauthHref(provider, redirectParam, fallbackHref) {
  const redirectUrl =
    getSafeRedirectUrl(redirectParam) ??
    new URL(fallbackHref, window.location.origin).href;

  return `/api/v1/auth/oauth/${provider}?${new URLSearchParams({
    redirect: redirectUrl,
  })}`;
}

function TextLink({ link, styles = "", children }) {
  const className = `${styles} text-fg-primary focus-visible:outline-primary
    rounded font-medium hover:underline focus-visible:outline-2
    focus-visible:outline-offset-2`;

  if (link.href) {
    return (
      <a href={link.href} className={className}>
        {children}
      </a>
    );
  }

  return (
    <Link to={link.to} search={link.search} className={className}>
      {children}
    </Link>
  );
}

function OauthButton({ href, icon: ProviderIcon, label }) {
  return (
    <a
      href={href}
      className="border-border-strong bg-bg-input text-fg hover:bg-hover
        focus-visible:outline-primary shadow-card flex items-center
        justify-center gap-2.5 rounded-lg border py-2.5 text-sm font-medium
        transition-colors focus-visible:outline-2
        focus-visible:outline-offset-2"
    >
      <ProviderIcon styles="fill-fg size-4.5" />
      {label}
    </a>
  );
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

export default function LoginPage({
  redirect: redirectParam,
  oauthError,
  fallbackHref,
  homeLink,
  switchLink,
  title,
  description,
  forgotPasswordLink,
  signupLink,
  headline,
  highlights,
}) {
  const navigate = useNavigate();
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

    // The oauth error is no longer relevant once the user tries again
    if (oauthError) {
      navigate({
        to: ".",
        search: (prev) => ({ ...prev, error: undefined }),
        replace: true,
      });
    }

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
    navigate({
      href: getRedirectHref(redirectParam, fallbackHref),
      replace: true,
    });
  }

  return (
    <div className="bg-bg flex min-h-dvh">
      <main className="flex flex-1 flex-col px-6 py-6 sm:px-10 sm:py-8">
        <header className="flex items-center justify-between gap-4">
          {homeLink ? (
            <Link
              to={homeLink.to}
              className="focus-visible:outline-primary rounded-lg font-semibold
                focus-visible:outline-2 focus-visible:outline-offset-4"
            >
              Reservations
            </Link>
          ) : (
            <span className="font-semibold">Reservations</span>
          )}
          <a
            href={switchLink.href}
            className="border-border-strong bg-bg-input text-fg hover:bg-hover
              focus-visible:outline-primary shadow-card flex shrink-0
              items-center gap-1.5 rounded-lg border px-3 py-1.5 text-sm
              font-medium transition-colors focus-visible:outline-2
              focus-visible:outline-offset-2"
          >
            {switchLink.label}
            <Icon
              icon={ArrowRight01Icon}
              styles="text-fg-muted size-4"
              aria-hidden="true"
            />
          </a>
        </header>
        <div
          className="mx-auto flex w-full max-w-sm flex-1 flex-col justify-center
            py-12"
        >
          <h1 className="text-3xl font-semibold tracking-tight">{title}</h1>
          <p className="text-fg-muted mt-2">{description}</p>

          <div className="mt-8 flex flex-col gap-3">
            <OauthButton
              href={getOauthHref("google", redirectParam, fallbackHref)}
              icon={GoogleIcon}
              label="Continue with Google"
            />
            <OauthButton
              href={getOauthHref("facebook", redirectParam, fallbackHref)}
              icon={FacebookIcon}
              label="Continue with Facebook"
            />
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
            <ServerError
              error={serverError ?? getOauthErrorMessage(oauthError)}
            />

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
                <TextLink link={forgotPasswordLink} styles="text-sm">
                  Forgot password?
                </TextLink>
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
            <TextLink link={signupLink}>Sign up</TextLink>
          </p>
        </div>
      </main>

      <aside
        className="bg-primary text-fg-on-primary relative hidden overflow-hidden
          lg:flex lg:w-[44%] lg:flex-col lg:justify-end lg:p-14 xl:p-20"
      >
        <div
          aria-hidden="true"
          className="absolute -top-32 -right-32 size-112 rounded-full
            bg-white/10"
        />
        <div
          aria-hidden="true"
          className="absolute top-40 -right-10 size-56 rounded-full bg-white/10"
        />
        <div className="relative max-w-md">
          <p className="text-4xl leading-tight font-semibold tracking-tight">
            {headline}
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
      {message}
    </p>
  );
}
