import Input from "../../components/Input";

export default function PasswordPage() {
  return (
    <div className="mt-6 flex flex-col items-center justify-center gap-4">
      <div
        className="relative flex w-full items-center justify-between border-2 border-customtxt
          focus-within:border-primary focus-within:outline-none"
      >
        <Input
          styles={"peer mt-4"}
          type={"password"}
          name={"password"}
          ariaLabel={"password"}
          required={true}
          autoComplete={"new-password"}
          minLength={"6"}
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
      <div
        className="relative flex w-full items-center justify-between border-2 border-customtxt
          focus-within:border-primary focus-within:outline-none"
      >
        <Input
          styles={"peer mt-4"}
          type={"password"}
          name={"password"}
          ariaLabel={"password"}
          required={true}
          autoComplete={"current-password"}
          minLength={"6"}
          id={"confirmPassword"}
        />
        <label
          className="pointer-events-none absolute left-3 scale-110 text-gray-400 transition-all
            peer-autofill:left-0.5 peer-autofill:-translate-y-4 peer-autofill:scale-90
            peer-autofill:text-customtxt peer-valid:left-1 peer-valid:-translate-y-4
            peer-valid:scale-90 peer-valid:text-customtxt peer-focus:left-0.5
            peer-focus:-translate-y-4 peer-focus:scale-90 peer-focus:text-primary"
          htmlFor="confirmPassword"
        >
          Confirm Password
        </label>
      </div>
    </div>
  );
}
