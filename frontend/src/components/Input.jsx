import { useState } from "react";
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
}) {
  const [inputValue, setInputValue] = useState("");

  function handleChange(e) {
    const value = e.target.value;
    setInputValue(value);
    inputData({
      name: name,
      value: e.target.value,
    });
  }

  const isEmpty = hasError && !inputValue;

  return (
    <label htmlFor={id} className="mt-4 flex flex-col">
      <span>{labelText}</span>
      <InputBase
        styles={`${styles} peer border-2 bg-transparent p-2 outline-none
          invalid:[&:not(:placeholder-shown):not(:focus)]:border-red-600 mt-2
          invalid:[&:not(:placeholder-shown):not(:focus)]:autofill:border-text_color
          ${isEmpty ? "border-red-600 focus:border-red" : "border-text_color focus:border-primary"}`}
        type={type}
        name={name}
        id={id}
        autoComplete={autoComplete}
        pattern={pattern}
        placeholder={placeholder}
        onChange={handleChange}
        required={true}
      />
      {isEmpty && (
        <span className="text-sm text-red-600">
          Please fill out this field!
        </span>
      )}
      <span
        className="hidden text-sm text-red-600
          peer-[&:not(:placeholder-shown):not(:focus):invalid]:block"
      >
        {errorText}
      </span>
    </label>
  );
}
