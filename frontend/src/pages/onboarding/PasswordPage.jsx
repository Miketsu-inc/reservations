import Input from "../../components/Input";

export default function PasswordPage() {
  return (
    <div className="flex flex-col items-center justify-center gap-4 mt-6">
      <div className="relative flex items-center justify-between w-full border-2 border-customtxt focus-within:border-primary focus-within:outline-none ">
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
          className="absolute left-2.5 text-gray-400 scale-110 pointer-events-none transition-all peer-focus:scale-90 peer-focus:-translate-y-4 peer-focus:left-1 peer-focus:text-primary peer-valid:scale-90 peer-valid:-translate-y-4 peer-valid:left-1 peer-valid:text-customtxt peer-autofill:scale-90 peer-autofill:-translate-y-4 peer-autofill:left-0.5 peer-autofill:text-customtxt"
          htmlFor="passwordInput"
        >
          Password
        </label>
      </div>
      <div className="relative flex items-center justify-between w-full border-2 border-customtxt focus-within:border-primary focus-within:outline-none ">
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
          className="absolute left-3 text-gray-400 scale-110 pointer-events-none transition-all peer-focus:scale-90 peer-focus:-translate-y-4 peer-focus:left-0.5 peer-focus:text-primary peer-valid:scale-90 peer-valid:-translate-y-4 peer-valid:left-1 peer-valid:text-customtxt peer-autofill:scale-90 peer-autofill:-translate-y-4 peer-autofill:left-0.5 peer-autofill:text-customtxt"
          htmlFor="confirmPassword"
        >
          Confirm Password
        </label>
      </div>
    </div>
  );
}
