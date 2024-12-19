import Button from "@components/Button";
import FloatingLabelInput from "@components/FloatingLabelInput";
import ServerError from "@components/ServerError";
import GoogleIcon from "@icons/GoogleIcon";
import {
  MAX_INPUT_LENGTH,
  MAX_PASSWORD_LENGTH,
  MIN_PASSWORD_LENGTH,
} from "@lib/constants";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { createFileRoute, Link, useRouter } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";

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
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [serverError, setServerError] = useState(undefined);
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

  useEffect(() => {
    if (isSubmitting) {
      const sendRequest = async () => {
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
            invalidateLocalSotrageAuth(response.status);
            const result = await response.json();
            setServerError(result.error.message);
          } else {
            router.history.push(search.redirect);
            setServerError(undefined);
          }
        } catch (err) {
          setServerError(err.message);
        } finally {
          setIsSubmitting(false);
        }
      };
      sendRequest();
    }
  }, [loginData, isSubmitting, router, search]);

  function formSubmitHandler(e) {
    e.preventDefault();
    let hasError = false;

    if (!loginData.email.isValid) {
      emailRef.current.triggerValidationError();
      hasError = true;
    }
    if (!loginData.password.isValid) {
      passwordRef.current.triggerValidationError();
      hasError = true;
    }
    if (!hasError) {
      setIsSubmitting(true);
    }
  }
  return (
    <div className="flex min-h-screen min-w-min items-center justify-center">
      <div
        className="flex min-h-screen w-full max-w-md flex-col px-10 sm:h-auto sm:min-h-0
          sm:rounded-md sm:bg-layer_bg sm:py-8 sm:shadow-lg lg:px-8"
      >
        <h2 className="mt-8 py-1 text-4xl font-bold sm:mt-4">Login</h2>
        <ServerError styles="mb-2 mt-4" error={serverError} />
        <p className="text-sms mt-2 py-2">Welcome back!</p>

        <Button
          type="Button"
          name="Goolge button"
          styles="group flex justify-center items-center my-2 dark:bg-transparent bg-secondary
            dark:border dark:border-secondary dark:hover:border-hvr_secondary
            dark:text-secondary text-text_color dark:hover:text-hvr_secondary
            dark:focus:outline-none dark:focus:text-hvr_secondary
            dark:focus:border-hvr_secondary hover:bg-hvr_secondary/90 focus:bg-secondary"
          buttonText="Log in with Google"
        >
          <GoogleIcon styles="dark:fill-secondary dark:group-hover:fill-hvr_secondary fill-text_color mr-3" />
        </Button>
        <div className="mt-4 grid grid-cols-3 items-center">
          <hr className="border-text_color" />
          <p className="text-center text-sm">OR</p>
          <hr className="border-text_color" />
        </div>
        <form
          onSubmit={formSubmitHandler}
          method="POST"
          action=""
          autoComplete="on"
          className="flex flex-col"
        >
          <FloatingLabelInput
            ref={emailRef}
            styles="mt-4"
            type="text"
            name="email"
            id="emailInput"
            ariaLabel="Email"
            autoComplete="email"
            labelText="Email"
            errorText={errorMessage.email}
            inputValidation={emailValidation}
            inputData={handleInputData}
          />
          <FloatingLabelInput
            ref={passwordRef}
            styles="mt-4"
            type="password"
            name="password"
            id="passwordInput"
            ariaLabel="Password"
            autoComplete="password"
            labelText="Password"
            errorText={errorMessage.password}
            inputValidation={passwordValidation}
            inputData={handleInputData}
          />
          <a
            href="#"
            className="mt-3 text-right text-sm hover:underline focus:underline focus:outline-none"
          >
            Forgot your password?
          </a>
          <Button
            name="login"
            type="submit"
            styles="mt-4 focus-visible:outline-1 hover:bg-hvr_primary text-white"
            buttonText="Login"
            isLoading={isSubmitting}
          />
        </form>
        <hr className="mt-10 border-text_color" />
        <div className="mt-2 flex items-center justify-evenly pb-4 pt-8 text-sm sm:mt-2 sm:pt-8">
          <p className="flex-1">If you don't have an account...</p>
          <Link
            to="/signup"
            className="whitespace-nowrap px-4 py-2 hover:underline"
          >
            Sign up
          </Link>
        </div>
      </div>
    </div>
  );
}
