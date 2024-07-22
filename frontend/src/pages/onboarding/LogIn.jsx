import GoogleIcon from "../../assets/GoogleIcon";
import Button from "../../components/Button";
import Input from "../../components/Input";

export default function LogIn() {
  return (
    <div className="bg-custombg sm:bg-white flex justify-center items-center min-h-screen min-w-min">
      <div className="flex flex-col bg-custombg shadow-sm w-full min-h-screen sm:shadow-lg sm:rounded-md md:rounded-md lg:rounded-md xl:rounded-md sm:min-h-0 sm:h-auto lg:h-auto xl:h-auto max-w-md px-10 lg:px-8 xl:px-8 sm:py-8">
        <h2 className="text-customtxt text-4xl font-bold py-1 mt-8 sm:mt-4">
          Login
        </h2>
        <p className="text-sm text-customtxt py-2 mt-2">Welcome back!</p>
        {/* <button className=" flex justify-center items-center rounded-lg text-customtxt font-medium gap-2 bg-secondary py-2 mt-10 mb-2 hover:bg-customhvr2 active:scale-95 active:shadow-none"> */}
        <Button
          type={"Button"}
          name={"Goolge button"}
          styles={"flex justify-center items-center gap-2 my-2"}
        >
          <GoogleIcon width={"20"} height={"20"} />
          Log in with google
        </Button>
        <div className="mt-4 grid grid-cols-3 items-center text-customtxt">
          <hr className="border-customtxt" />
          <p className="text-center text-sm">OR</p>
          <hr className="border-customtxt" />
        </div>
        <form
          method="POST"
          action=""
          autoComplete="on"
          className=" flex flex-col gap-4"
        >
          <div className="relative flex items-center justify-center w-full border-2 mt-6 border-customtxt focus-within:border-primary focus-within:outline-none ">
            <Input
              styles={"peer mt-4"}
              type={"text"}
              name={"email"}
              ariaLabel={"Email"}
              required={true}
              autocomplete={"email"}
              id={"emailInput"}
            />
            <label
              className="absolute left-2.5 text-gray-400 scale-110 pointer-events-none transition-all peer-focus:scale-90 peer-focus:-translate-y-4 peer-focus:left-1 peer-focus:text-primary peer-valid:scale-90 peer-valid:-translate-y-4 peer-valid:left-1 peer-valid:text-customtxt peer-autofill:scale-90 peer-autofill:-translate-y-4 peer-autofill:left-0.5 peer-autofill:text-customtxt"
              htmlFor="emailInput"
            >
              Email
            </label>
          </div>
          <div className="relative flex items-center justify-between w-full border-2 border-customtxt focus-within:border-primary focus-within:outline-none ">
            <Input
              styles={"peer mt-4"}
              type={"password"}
              name={"password"}
              ariaLabel={"Password"}
              required={true}
              autocomplete={"password"}
              id={"passwordInput"}
            />
            <label
              className="absolute left-2.5 text-gray-400 scale-110 pointer-events-none transition-all peer-focus:scale-90 peer-focus:-translate-y-4 peer-focus:left-1 peer-focus:text-primary peer-valid:scale-90 peer-valid:-translate-y-4 peer-valid:left-1 peer-valid:text-customtxt peer-autofill:scale-90 peer-autofill:-translate-y-4 peer-autofill:left-0.5 peer-autofill:text-customtxt"
              htmlFor="passwordInput"
            >
              Password
            </label>
          </div>
          <a
            href="#"
            className="text-sm text-customtxt text-right hover:underline"
          >
            Forgot your password?
          </a>
          <Button name={"login"} type={"submit"} styles={"mt-2"}>
            Login
          </Button>
        </form>

        {/*Checkbox-- remember me 
        <label>
          <input type="checkbox" id="remember" className="pr-2" /> Remember me
        </label>*/}
        <hr className="border-gray-300 mt-10" />
        <div className="flex justify-evenly items-center text-sm mt-2 pt-8 pb-4 sm:mt-2 sm:pt-8">
          <p className="flex-1 text-customtxt">
            If you don't have an account...
          </p>
          <a
            href="/signup"
            className="rounded-lg border text-customtxt font-normal border-accent hover:border-2 px-4 py-2 whitespace-nowrap"
          >
            Sign up
          </a>
        </div>
      </div>
    </div>
  );
}
