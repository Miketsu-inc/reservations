import Input from "../../components/Input";

export default function EmailPage({
  formValues,
  handleBlur,
  handleChange,
  isValid,
  errors,
}) {
  return (
    <>
      <div
        className="relative mt-6 flex w-full items-center justify-center border-2 border-customtxt
          focus-within:border-primary focus-within:outline-none"
      >
        <Input
          styles="peer mt-4"
          type="text"
          name="email"
          ariaLabel="Email"
          required={true}
          autoComplete="email"
          id="emailInput"
          value={formValues.email}
          onChange={handleChange}
          onBlur={handleBlur}
        />
        <label
          className="pointer-events-none absolute left-2.5 scale-110 text-gray-400 transition-all
            peer-valid:left-1 peer-valid:-translate-y-4 peer-valid:scale-90
            peer-valid:text-customtxt peer-focus:left-1 peer-focus:-translate-y-4
            peer-focus:scale-90 peer-focus:text-primary"
          htmlFor="emailInput"
        >
          Email
        </label>
      </div>
      {errors.email && (
        <span className="text-sm text-red-600">{errors.email}</span>
      )}
    </>
  );
}
