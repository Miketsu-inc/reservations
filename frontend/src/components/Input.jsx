import InputBase from "./InputBase";

export default function Input({
  id,
  name,
  type,
  autoComplete,
  styles,
  labelText,
  errorText,
  placeholder,
  pattern,
  inputData,
  hasError,
  value,
  required,
  min,
  max,
  autoFocus,
}) {
  function handleChange(e) {
    inputData({
      name: name,
      value: e.target.value,
    });
  }

  const isEmpty = hasError && !value;

  return (
    <label htmlFor={id} className="flex w-full flex-col">
      {labelText && <span className="pb-1">{labelText}</span>}
      <InputBase
        styles={`${styles} ps-1 peer border-2 bg-transparent outline-hidden
          invalid:[&:not(:placeholder-shown):not(:focus)]:border-red-600
          invalid:[&:not(:placeholder-shown):not(:focus)]:autofill:border-text_color
          ${isEmpty ? "border-red-600 focus:border-red-600" : "border-text_color focus:border-primary"}`}
        type={type}
        name={name}
        id={id}
        autoComplete={autoComplete}
        pattern={pattern}
        placeholder={placeholder}
        onChange={handleChange}
        required={required === undefined ? true : required}
        onBlur={() => {}}
        value={value}
        min={min}
        max={max}
        autoFocus={autoFocus}
      />
      {isEmpty && (
        <span className="text-sm text-red-600">
          Please fill out this field!
        </span>
      )}
      {errorText && (
        <span
          className="hidden text-sm text-red-600
            peer-[&:not(:placeholder-shown):not(:focus):invalid]:block"
        >
          {errorText}
        </span>
      )}
    </label>
  );
}
