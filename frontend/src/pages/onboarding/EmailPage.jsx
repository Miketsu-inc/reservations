import Input from "../../components/Input";

export default function EmailPage() {
  return (
    <div className="relative flex items-center justify-center w-full mt-6 border-2 border-customtxt focus-within:border-primary focus-within:outline-none ">
      <Input
        styles={"peer mt-4"}
        type={"text"}
        name={"email"}
        ariaLabel={"Email"}
        required={true}
        autoComplete={"email"}
        id={"emailInput"}
      />
      <label
        className="absolute left-2.5 text-gray-400 scale-110 pointer-events-none transition-all peer-focus:scale-90 peer-focus:-translate-y-4 peer-focus:left-1 peer-focus:text-primary peer-valid:scale-90 peer-valid:-translate-y-4 peer-valid:left-1 peer-valid:text-customtxt"
        htmlFor="emailInput"
      >
        Email
      </label>
    </div>
  );
}
