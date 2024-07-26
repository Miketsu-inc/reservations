import { useState } from "react";
import GoogleIcon from "../../assets/GoogleIcon";
import Button from "../../components/Button";
import Input from "../../components/Input";

export default function LogIn() {
  const [formValues, setFormValues] = useState({
    email: "",
    password: "",
  });
  const [errors, setErrors] = useState({});
  const [isValid, setIsValid] = useState({ email: true, password: true });

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormValues((prevData) => ({ ...prevData, [name]: value }));
    setErrors((prevErrors) => ({ ...prevErrors, [name]: "" }));
    setIsValid((prevIsValid) => ({ ...prevIsValid, [name]: true }));
  };

  const handleBlur = (e) => {
    const { name, value } = e.target;

    const Errors = { ...errors };
    const newIsValid = { ...isValid };
    if (
      formValues.email.includes("@") !== true &&
      name === "email" &&
      formValues.email.trim() !== ""
    ) {
      Errors.email = "Please enter a valid email!";
      newIsValid.email = false;
    } else if (
      name === "password" &&
      value.length < 6 &&
      formValues.password.trim() !== ""
    ) {
      Errors.password = "Please enter your password!";
      newIsValid.password = false;
    } else {
      delete Errors[name];
      newIsValid[name] = true;
    }

    setIsValid(newIsValid);
    setErrors(Errors);
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    const Errors = { ...errors };
    const newIsValid = { ...isValid };
    if (formValues.email.trim() == "") {
      Errors.email = "This field is is required";
      newIsValid.email = false;
    }
    if (formValues.password.trim() == "") {
      Errors.password = "This field is is required";
      newIsValid.password = false;
    }
    setErrors(Errors);
    setIsValid(newIsValid);

    if (isValid.email === true && isValid.password === true) {
      console.log(formValues);
    }
  };

  return (
    <div className="flex min-h-screen min-w-min items-center justify-center bg-custombg">
      <div
        className="flex min-h-screen w-full max-w-md flex-col bg-custombg px-10 shadow-sm sm:h-auto
          sm:min-h-0 sm:rounded-md sm:bg-slate-400 sm:bg-opacity-5 sm:py-8 sm:shadow-lg
          md:rounded-md lg:h-auto lg:rounded-md lg:px-8 xl:h-auto xl:rounded-md xl:px-8"
      >
        <h2
          className={
            isValid.password === true || isValid.email === true
              ? "mt-8 py-1 text-4xl font-bold text-customtxt sm:mt-4"
              : "mt-8 py-1 text-4xl font-bold text-customtxt sm:mt-1"
          }
        >
          Login
        </h2>
        <p className="mt-2 py-2 text-sm text-customtxt">Welcome back!</p>
        {/* <button className=" flex justify-center items-center rounded-lg text-customtxt font-medium gap-2 bg-secondary py-2 mt-10 mb-2 hover:bg-customhvr2 active:scale-95 active:shadow-none"> */}
        <Button
          type="Button"
          name="Goolge button"
          styles="group flex justify-center items-center gap-2 my-2 bg-transparent border border-secondary hover:border-customhvr2 hover:bg-transparent text-secondary hover:text-customhvr2"
        >
          <GoogleIcon
            width="20"
            height="20"
            styles="fill-secondary group-hover:fill-customhvr2"
          />
          Log in with google
        </Button>
        <div className="mt-4 grid grid-cols-3 items-center text-customtxt">
          <hr className="border-customtxt" />
          <p className="text-center text-sm">OR</p>
          <hr className="border-customtxt" />
        </div>
        <form
          onSubmit={handleSubmit}
          method="POST"
          action=""
          autoComplete="on"
          className="flex flex-col"
        >
          <div
            className={
              isValid.email === true
                ? `relative mt-6 flex w-full items-center justify-center border-2 border-customtxt
                  focus-within:border-primary focus-within:outline-none`
                : `relative mt-6 flex w-full items-center justify-center border-2 border-red-600
                  focus-within:border-red-600 focus-within:outline-none`
            }
          >
            <Input
              styles="peer mt-4"
              type="text"
              name="email"
              ariaLabel="Email"
              autocomplete="email"
              id="emailInput"
              value={formValues.email}
              onChange={handleChange}
              onBlur={handleBlur}
            />
            <label
              className={
                formValues.email.trim() == "" && isValid.email === true
                  ? `pointer-events-none absolute left-2.5 scale-110 text-gray-400 transition-all
                    peer-autofill:left-0.5 peer-autofill:-translate-y-4 peer-autofill:scale-90
                    peer-autofill:text-customtxt peer-focus:left-1 peer-focus:-translate-y-4
                    peer-focus:scale-90 peer-focus:text-primary`
                  : isValid.email === false
                    ? `pointer-events-none absolute left-1 -translate-y-4 scale-90 text-red-600
                      peer-autofill:left-0.5 peer-autofill:-translate-y-4 peer-autofill:scale-90
                      peer-autofill:text-customtxt`
                    : `pointer-events-none absolute left-1 -translate-y-4 scale-90 text-customtxt
                      transition-all peer-autofill:left-0.5 peer-autofill:-translate-y-4
                      peer-autofill:scale-90 peer-autofill:text-customtxt peer-focus:left-1
                      peer-focus:-translate-y-4 peer-focus:scale-90 peer-focus:text-primary`
              }
              htmlFor="emailInput"
            >
              Email
            </label>
          </div>
          {errors.email && (
            <span className="text-sm text-red-600">{errors.email}</span>
          )}
          <div
            className={
              isValid.password === true
                ? `relative mt-4 flex w-full items-center justify-between border-2 border-customtxt
                  focus-within:border-primary focus-within:outline-none`
                : `relative mt-4 flex w-full items-center justify-between border-2 border-red-600
                  focus-within:border-red-600 focus-within:outline-none`
            }
          >
            <Input
              styles="peer mt-4"
              type="password"
              name="password"
              ariaLabel="Password"
              autocomplete="password"
              id="passwordInput"
              value={formValues.password}
              onChange={handleChange}
              onBlur={handleBlur}
            />
            <label
              className={
                formValues.password.trim() == "" && isValid.password === true
                  ? `pointer-events-none absolute left-2.5 scale-110 text-gray-400 transition-all
                    peer-autofill:left-0.5 peer-autofill:-translate-y-4 peer-autofill:scale-90
                    peer-autofill:text-customtxt peer-focus:left-1 peer-focus:-translate-y-4
                    peer-focus:scale-90 peer-focus:text-primary`
                  : isValid.password === false
                    ? `pointer-events-none absolute left-1 -translate-y-4 scale-90 text-red-600
                      peer-autofill:left-0.5 peer-autofill:-translate-y-4 peer-autofill:scale-90
                      peer-autofill:text-customtxt`
                    : `pointer-events-none absolute left-1 -translate-y-4 scale-90 text-customtxt
                      transition-all peer-autofill:left-0.5 peer-autofill:-translate-y-4
                      peer-autofill:scale-90 peer-autofill:text-customtxt peer-focus:left-1
                      peer-focus:-translate-y-4 peer-focus:scale-90 peer-focus:text-primary`
              }
              htmlFor="passwordInput"
            >
              {/*  */}
              Password
            </label>
          </div>
          {errors.password && (
            <span className="text-sm text-red-600 transition-all">
              {errors.password}
            </span>
          )}
          <a
            href="#"
            className={
              isValid.password === true
                ? "mt-3 text-right text-sm text-customtxt hover:underline"
                : "mb-2 text-right text-sm text-customtxt hover:underline"
            }
          >
            Forgot your password?
          </a>
          <Button name="login" type="submit" styles="mt-4">
            Login
          </Button>
        </form>

        {/*Checkbox-- remember me
        <label>
          <input type="checkbox" id="remember" className="pr-2" /> Remember me
        </label>*/}
        <hr
          className={
            isValid.password === true
              ? "mt-10 border-gray-300"
              : "mt-8 border-gray-300"
          }
        />
        <div className="mt-2 flex items-center justify-evenly pb-4 pt-8 text-sm sm:mt-2 sm:pt-8">
          <p className="flex-1 text-customtxt">
            If you don't have an account...
          </p>
          <a
            href="/signup"
            className="whitespace-nowrap px-4 py-2 font-normal text-customtxt hover:underline"
          >
            Sign up
          </a>
        </div>
      </div>
    </div>
  );
}
