import Button from "@components/Button";
import FloatingLabelInput from "@components/FloatingLabelInput";
import ServerError from "@components/ServerError";
import GoogleIcon from "@icons/GoogleIcon";
import {
  MAX_INPUT_LENGTH,
  MAX_PASSWORD_LENGTH,
  MIN_PASSWORD_LENGTH,
} from "@lib/constants";
import { invalidateLocalStorageAuth } from "@lib/lib";
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

    if (password.length > MAX_PASSWORD_LENGTH) {
      updateErrors(
        "password",
        `Password must be ${MAX_PASSWORD_LENGTH} characters or less!`
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
      const response = await fetch("/api/v1/auth/user/login", {
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

        if (search.redirect) {
          router.history.push(search.redirect);
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
    <div className="flex min-h-screen min-w-min items-center justify-center">
      <div
        className="flex w-full max-w-md flex-col px-8 sm:h-auto sm:min-h-0 sm:rounded-xl
          sm:bg-layer_bg sm:py-8 sm:shadow-lg lg:px-8"
      >
        <h2 className="mt-8 py-1 text-4xl font-bold sm:mt-4">Login</h2>
        <ServerError styles="mb-2 mt-4" error={serverError} />
        <p className="text-sms mt-2 py-2">Welcome back!</p>

        <Button
          type="Button"
          name="Goolge button"
          styles="group flex justify-center items-center my-2 bg-transparent border
            border-secondary hover:border-hvr_secondary text-secondary! font-normal!
            hover:*:text-hvr_secondary focus:outline-hidden focus:*:text-hvr_secondary
            focus:border-hvr_secondary hover:bg-transparent py-2"
          buttonText="Log in with Google"
        >
          <GoogleIcon styles="group-hover:fill-hvr_secondary fill-secondary mr-3" />
        </Button>
        <div className="my-4 grid grid-cols-3 items-center">
          <hr className="border-text_color" />
          <p className="text-center text-sm">OR</p>
          <hr className="border-text_color" />
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
          <a
            href="#"
            className="text-right text-sm hover:underline focus:underline focus:outline-hidden"
          >
            Forgot your password?
          </a>
          <Button
            variant="primary"
            name="login"
            type="submit"
            styles="py-2"
            buttonText="Login"
            isLoading={isLoading}
          />
        </form>
        <hr className="mt-10 border-text_color" />
        <div className="mt-2 flex items-center justify-evenly pb-4 pt-8 text-sm sm:mt-2 sm:pt-8">
          <p className="flex-1">If you don't have an account...</p>
          <Link
            from={Route.fullPath}
            to="/signup"
            search={{
              redirect: search.redirect,
            }}
            className="whitespace-nowrap px-4 py-2 hover:underline"
          >
            Sign up
          </Link>
        </div>
      </div>
    </div>
  );
}
