import { useRef, useState } from "react";
import GoogleIcon from "../../assets/GoogleIcon";
import Button from "../../components/Button";
import Input from "../../components/Input";
import { MIN_PASSWORD_LENGTH } from "../../lib/constants";

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

export default function LogIn() {
  const emailRef = useRef();
  const passwordRef = useRef();
  const [loginData, setLoginData] = useState(defaultLoginData);

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
    return email.includes("@");
  }

  function passwordValidation(password) {
    return password.length > MIN_PASSWORD_LENGTH;
  }

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
        className="flex min-h-screen w-full max-w-md flex-col px-10 shadow-sm sm:h-auto sm:min-h-0
          sm:rounded-md sm:bg-slate-400 sm:bg-opacity-5 sm:py-8 sm:shadow-lg md:rounded-md
          lg:h-auto lg:rounded-md lg:px-8 xl:h-auto xl:rounded-md xl:px-8"
      >
        <h2 className={"mt-8 py-1 text-4xl font-bold sm:mt-4"}>Login</h2>
        <p className="mt-2 py-2 text-sm">Welcome back!</p>
        <Button
          type="Button"
          name="Goolge button"
          styles="group flex justify-center items-center gap-2 my-2 bg-transparent border
            border-secondary hover:border-customhvr2 hover:bg-transparent text-secondary
            hover:text-customhvr2"
        >
          <GoogleIcon styles="fill-secondary group-hover:fill-customhvr2" />
          Log in with google
        </Button>
        <div className="mt-4 grid grid-cols-3 items-center">
          <hr />
          <p className="text-center text-sm">OR</p>
          <hr />
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
            errorText="Please enter a valid email!"
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
            errorText="Please enter your password!"
            inputValidation={passwordValidation}
            inputData={handleInputData}
          />
          <a href="#" className={"mt-3 text-right text-sm hover:underline"}>
            Forgot your password?
          </a>
          <Button name="login" type="submit" styles="mt-4">
            Login
          </Button>
        </form>
        <hr className={"mt-10 border-gray-300"} />
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
