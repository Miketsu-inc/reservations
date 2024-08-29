import { useRef, useState } from "react";
import GoogleIcon from "../../assets/GoogleIcon";
import Button from "../../components/Button";
import Input from "../../components/Input";
import {
  MAX_INPUT_LENGTH,
  MAX_PASSWORD_LENGTH,
MIN_PASSWORD_LENGTH,
} from "../../lib/constants";

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

export default function LoginPage() {
  const emailRef = useRef();
  const passwordRef = useRef();
  const [loginData, setLoginData] = useState(defaultLoginData);
  const [isSubmitting, setIsSubmitting] = useState(false);
const [errorMessage, setErrorMessage] = useState(
    "Please fill out this field!"
  );

  function handleInputData(data) {
    setLoginData((prevLoginData) => ({
      ...prevLoginData,
      [data.name]: {
        value: data.value,
        isValid: data.isValid,
      },
    }));
  }

  function emailValidation(email) {
    if (email.length > MAX_INPUT_LENGTH) {
      setErrorMessage(`Inputs must be ${MAX_INPUT_LENGTH} characters or less!`);
      return false;
    }
    if (!email.includes("@")) {
      setErrorMessage("Please enter a valid email!");
      return false;
    }
    return true;
  }

  function passwordValidation(password) {
    if (password.length < MIN_PASSWORD_LENGTH) {
      setErrorMessage(
        `Password must be ${MIN_PASSWORD_LENGTH} characters or more!`
      );
      return false;
    }
    if (password.length > MAX_PASSWORD_LENGTH) {
      setErrorMessage(
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
          const response = await fetch("/api/login", {
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
          const result = await response.json();
          console.log(result);
        } catch (err) {
          console.log(err);
        } finally {
          setIsSubmitting(false);
        }
      };
      sendRequest();
    }
  }, [loginData, isSubmitting]);

  function formSubmitHandler(e) {
    e.preventDefault();

    if (!loginData.email.isValid) {
      emailRef.current.triggerValidationError();
    }

    if (!loginData.password.isValid) {
      passwordRef.current.triggerValidationError();
    }
  }

  return (
    <div className="flex min-h-screen min-w-min items-center justify-center">
      <div
        className="sm:bg-layer-bg flex min-h-screen w-full max-w-md flex-col px-10 sm:h-auto
sm:min-h-0 sm:rounded-md sm:py-8 sm:shadow-lg lg:px-8"
      >
        <h2 className="mt-8 py-1 text-4xl font-bold sm:mt-4">Login</h2>
        <p className="mt-2 py-2 text-sm">Welcome back!</p>
        <Button
          type="Button"
          name="Goolge button"
          styles="group flex justify-center items-center my-2 dark:bg-transparent bg-secondary/50
            dark:border dark:border-secondary dark:hover:border-hvr-secondary
            dark:text-secondary text-text-color dark:hover:text-hvr-secondary
            dark:focus:outline-none dark:focus:text-hvr-secondary
dark:focus:border-hvr-secondary hover:bg-hvr-secondary/50 focus:bg-hvr-secondary"
          buttonText="Log in with Google"
        >
          <GoogleIcon styles="dark:fill-secondary dark:group-hover:fill-hvr-secondary fill-text-color" />
        </Button>
        <div className="mt-4 grid grid-cols-3 items-center">
          <hr className="border-text-color" />
          <p className="text-center text-sm">OR</p>
          <hr className="border-text-color" />
        </div>
        <form
          onSubmit={formSubmitHandler}
          method="POST"
          action=""
          autoComplete="on"
          className="flex flex-col"
        >
          <Input
            ref={emailRef}
            styles=""
            type="text"
            name="email"
            id="emailInput"
            ariaLabel="Email"
            autoComplete="email"
            labelText="Email"
            labelHtmlFor="emailInput"
            errorText={errorMessage}
            inputValidation={emailValidation}
            inputData={handleInputData}
          />
          <Input
            ref={passwordRef}
            styles=""
            type="password"
            name="password"
            id="passwordInput"
            ariaLabel="Password"
            autoComplete="password"
            labelText="Password"
            labelHtmlFor="passwordInput"
            errorText={errorMessage}
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
            styles="mt-4 focus-visible:outline-1 hover:bg-hvr-primary text-white"
            buttonText="Login"
            isLoading={isSubmitting}
          />
        </form>
        <hr className="border-text-color mt-10" />
        <div className="mt-2 flex items-center justify-evenly pb-4 pt-8 text-sm sm:mt-2 sm:pt-8">
          <p className="flex-1">If you don't have an account...</p>
          <a
            href="/signup"
            className="font- whitespace-nowrap px-4 py-2 hover:underline"
          >
            Sign up
          </a>
        </div>
      </div>
    </div>
  );
}
