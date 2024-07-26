import Input from "../../components/Input";

export default function PersonalInfo({
  formValues,
  handleBlur,
  handleChange,
  isValid,
  errors,
}) {
  return (
    <div className="mt-6 flex flex-col items-center justify-center gap-4">
      {/* Full name input box*/}
      <div
        className="relative flex w-full items-center justify-center border-2 border-customtxt
          focus-within:border-primary focus-within:outline-none"
      >
        <Input
          styles="peer mt-4"
          type="text"
          ariaLabel="first name"
          name="first Name"
          autoComplete="family-name"
          id="firstName"
          value={formValues.firstName}
          onChange={handleChange}
          onBlur={handleBlur}
        />
        <label
          className="pointer-events-none absolute left-2.5 scale-110 text-gray-400 transition-all
            peer-autofill:left-0.5 peer-autofill:-translate-y-4 peer-autofill:scale-90
            peer-autofill:text-customtxt peer-valid:left-1 peer-valid:-translate-y-4
            peer-valid:scale-90 peer-valid:text-customtxt peer-focus:left-1
            peer-focus:-translate-y-4 peer-focus:scale-90 peer-focus:text-primary"
          htmlFor="firstName"
        >
          First Name
        </label>
      </div>
      {errors.firstName && (
        <span className="text-sm text-red-600">{errors.email}</span>
      )}
      <div
        className="relative flex w-full items-center justify-center border-2 border-customtxt
          focus-within:border-primary focus-within:outline-none"
      >
        <Input
          styles="peer mt-4"
          type="text"
          ariaLabel="last name"
          name="Last Name"
          autoComplete="given-name"
          id="lastName"
          value={formValues.lastName}
          onChange={handleChange}
          onBlur={handleBlur}
        />
        <label
          className="pointer-events-none absolute left-2.5 scale-110 text-gray-400 transition-all
            peer-autofill:left-0.5 peer-autofill:-translate-y-4 peer-autofill:scale-90
            peer-autofill:text-customtxt peer-valid:left-1 peer-valid:-translate-y-4
            peer-valid:scale-90 peer-valid:text-customtxt peer-focus:left-1
            peer-focus:-translate-y-4 peer-focus:scale-90 peer-focus:text-primary"
          htmlFor="lastName"
        >
          Last Name
        </label>
      </div>
      {errors.lastName && (
        <span className="text-sm text-red-600">{errors.email}</span>
      )}
    </div>
  );
}
