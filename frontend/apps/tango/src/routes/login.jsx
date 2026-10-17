import { FacebookIcon, GoogleIcon } from "@reservations/assets";
import {
  Button,
  FloatingLabelInput,
  ServerError,
} from "@reservations/components";
import {
  getSafeRedirectUrl,
  invalidateLocalStorageAuth,
  MAX_INPUT_LENGTH,
  MIN_PASSWORD_LENGTH,
} from "@reservations/lib";
import { createFileRoute, Link, useRouter } from "@tanstack/react-router";
import { useRef, useState } from "react";

const defaultLoginData = {
  email: {
    value: "",
    isValid: false,
  },
  password: {
    value: "",
    isValid: false,
  },
};

const defaultErrorMeassage = {
  email: "Please enter your email",
  password: "Please enter your password",
};

export const Route = createFileRoute("/login")({
  component: LoginPage,
});

function LoginPage() {
  const router = useRouter();
  const search = Route.useSearch();
  const emailRef = useRef();
  const passwordRef = useRef();
  const [loginData, setLoginData] = useState(defaultLoginData);
  const [isLoading, setIsLoading] = useState(false);
  const [serverError, setServerError] = useState();
  const [errorMessage, setErrorMessage] = useState(defaultErrorMeassage);

  function handleInputData(data) {
    setLoginData((prevLoginData) => ({
      ...prevLoginData,
      [data.name]: {
        value: data.value,
        isValid: data.isValid,
      },
    }));
  }

  function updateErrors(key, message) {
    setErrorMessage((prevErrorMessage) => ({
      ...prevErrorMessage,
      [key]: message,
    }));
  }

  function emailValidation(email) {
    if (email.length > MAX_INPUT_LENGTH) {
      updateErrors(
        "email",
        `Inputs must be ${MAX_INPUT_LENGTH} characters or less!`
      );
      return false;
    }

    if (!email.includes("@")) {
      updateErrors("email", "Please enter a valid email");
      return false;
    }

    return true;
  }

  function passwordValidation(password) {
    if (password.length < MIN_PASSWORD_LENGTH) {
      updateErrors(
        "password",
        `Password must be ${MIN_PASSWORD_LENGTH} characters or more!`
      );
      return false;
    }

    return true;
  }

  async function formSubmitHandler(e) {
    e.preventDefault();

    if (!loginData.email.isValid) {
      emailRef.current.triggerValidationError();
      return;
    }
    if (!loginData.password.isValid) {
      passwordRef.current.triggerValidationError();
      return;
    }

    setIsLoading(true);
    try {
      const response = await fetch("/api/v1/auth/login", {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          email: loginData.email.value,
          password: loginData.password.value,
        }),
      });

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        setServerError();

        const redirectUrl = getSafeRedirectUrl(search.redirect);
        if (redirectUrl) {
          window.location.assign(redirectUrl);
        } else {
          router.navigate({ from: Route.fullPath, to: "/" });
        }
      }
    } catch (err) {
      setServerError(err.message);
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div
      className="flex min-h-screen min-w-min items-center justify-center py-10"
    >
      <div
        className="sm:bg-bg-raised sm:border-border sm:shadow-card flex w-full
          max-w-md flex-col px-6 sm:rounded-2xl sm:border sm:p-8"
      >
        <h1 className="text-2xl font-semibold">Login</h1>
        <ServerError styles="mb-2 mt-4" error={serverError} />
        <p className="text-fg-muted mt-1 mb-6 text-sm">Welcome back!</p>

        <a
          className="border-border-strong bg-bg-input text-fg hover:bg-hover
            focus-visible:outline-primary my-1.5 flex items-center
            justify-center gap-2.5 rounded-lg border py-2 text-sm font-medium
            transition-colors focus-visible:outline-2
            focus-visible:outline-offset-2"
          href="/api/v1/auth/oauth/google"
        >
          <GoogleIcon styles="fill-fg size-4" />
          Login with Google
        </a>
        <a
          className="border-border-strong bg-bg-input text-fg hover:bg-hover
            focus-visible:outline-primary my-1.5 flex items-center
            justify-center gap-2.5 rounded-lg border py-2 text-sm font-medium
            transition-colors focus-visible:outline-2
            focus-visible:outline-offset-2"
          href="/api/v1/auth/oauth/facebook"
        >
          <FacebookIcon styles="fill-fg size-5" />
          Login with Facebook
        </a>
        <div className="my-5 flex items-center gap-3">
          <hr className="border-border flex-1" />
          <p className="text-fg-muted text-xs">OR</p>
          <hr className="border-border flex-1" />
        </div>
        <form
          onSubmit={formSubmitHandler}
          method="POST"
          autoComplete="on"
          className="flex flex-col gap-4"
        >
          <FloatingLabelInput
            ref={emailRef}
            type="text"
            name="email"
            id="emailInput"
            autoComplete="email"
            labelText="Email"
            errorText={errorMessage.email}
            inputValidation={emailValidation}
            inputData={handleInputData}
          />
          <FloatingLabelInput
            ref={passwordRef}
            type="password"
            name="password"
            id="passwordInput"
            autoComplete="current-password"
            labelText="Password"
            errorText={errorMessage.password}
            inputValidation={passwordValidation}
            inputData={handleInputData}
          />
          <Link
            to="/forgot-password"
            className="text-fg-primary -mt-1 text-right text-sm font-medium
              hover:underline focus:underline focus:outline-hidden"
          >
            Forgot your password?
          </Link>
          <Button
            variant="primary"
            name="login"
            type="submit"
            styles="py-2.5"
            buttonText="Login"
            isLoading={isLoading}
          />
        </form>
        <hr className="border-border-subtle mt-8" />
        <div
          className="text-fg-muted flex items-center justify-center gap-1 pt-6
            text-sm"
        >
          <p>Don't have an account?</p>
          <Link
            from={Route.fullPath}
            to="/signup"
            search={{
              redirect: search.redirect,
            }}
            className="text-fg-primary font-medium whitespace-nowrap
              hover:underline"
          >
            Sign up
          </Link>
        </div>
      </div>
    </div>
  );
}
