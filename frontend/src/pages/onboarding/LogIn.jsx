import GoogleIcon from "../../assets/GoogleIcon";
import Button from "../../components/Button";
import Input from "../../components/Input";

export default function LogIn() {
  return (
    <div className="flex min-h-screen min-w-min items-center justify-center bg-custombg sm:bg-white">
      <div
        className="flex min-h-screen w-full max-w-md flex-col bg-custombg px-10 shadow-sm sm:h-auto
          sm:min-h-0 sm:rounded-md sm:py-8 sm:shadow-lg md:rounded-md lg:h-auto
          lg:rounded-md lg:px-8 xl:h-auto xl:rounded-md xl:px-8"
      >
        <h2 className="mt-8 py-1 text-4xl font-bold text-customtxt sm:mt-4">
          Login
        </h2>
        <p className="mt-2 py-2 text-sm text-customtxt">Welcome back!</p>
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
          className="flex flex-col gap-4"
        >
          <div
            className="relative mt-6 flex w-full items-center justify-center border-2 border-customtxt
              focus-within:border-primary focus-within:outline-none"
          >
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
              className="pointer-events-none absolute left-2.5 scale-110 text-gray-400 transition-all
                peer-autofill:left-0.5 peer-autofill:-translate-y-4 peer-autofill:scale-90
                peer-autofill:text-customtxt peer-valid:left-1 peer-valid:-translate-y-4
                peer-valid:scale-90 peer-valid:text-customtxt peer-focus:left-1
                peer-focus:-translate-y-4 peer-focus:scale-90 peer-focus:text-primary"
              htmlFor="emailInput"
            >
              Email
            </label>
          </div>
          <div
            className="relative flex w-full items-center justify-between border-2 border-customtxt
              focus-within:border-primary focus-within:outline-none"
          >
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
              className="pointer-events-none absolute left-2.5 scale-110 text-gray-400 transition-all
                peer-autofill:left-0.5 peer-autofill:-translate-y-4 peer-autofill:scale-90
                peer-autofill:text-customtxt peer-valid:left-1 peer-valid:-translate-y-4
                peer-valid:scale-90 peer-valid:text-customtxt peer-focus:left-1
                peer-focus:-translate-y-4 peer-focus:scale-90 peer-focus:text-primary"
              htmlFor="passwordInput"
            >
              Password
            </label>
          </div>
          <a
            href="#"
            className="text-right text-sm text-customtxt hover:underline"
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
        <hr className="mt-10 border-gray-300" />
        <div className="mt-2 flex items-center justify-evenly pb-4 pt-8 text-sm sm:mt-2 sm:pt-8">
          <p className="flex-1 text-customtxt">
            If you don't have an account...
          </p>
          <a
            href="/signup"
            className="whitespace-nowrap rounded-lg border border-accent px-4 py-2 font-normal
              text-customtxt hover:border-2"
          >
            Sign up
          </a>
        </div>
      </div>
    </div>
  );
}
