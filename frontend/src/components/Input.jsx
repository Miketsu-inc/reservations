import { useState } from "react";
import InputBase from "./InputBase";

export default function Input(props) {
  const [inputValue, setInputValue] = useState("");
  const [errorText, setErrorText] = useState("");
  const [isValid, setIsValid] = useState(false);

  const isEmpty = inputValue.trim() === "";

  function onChangeHandler(e) {
    setInputValue(e.target.value);
    setErrorText("");
    setIsValid(true);
  }

  function onBlurHandler(e) {
    if (props.inputValidation(inputValue)) {
      setIsValid(true);
      setErrorText("");
    } else {
      setErrorText(props.errorText);
      setIsValid(false);
    }

    props.inputData({
      name: props.name,
      value: inputValue,
      isValid: isValid,
    });
  }

  return (
    <>
      <div
        className={`${isValid || isEmpty ? "justify-between border-customtxt focus-within:border-primary" : "border-red-600 focus-within:border-red-600"}
          relative mt-6 flex w-full items-center border-2 focus-within:outline-none`}
      >
        <InputBase
          styles={`${props.styles} peer mt-4`}
          type={props.type}
          value={inputValue}
          name={props.name}
          id={props.id}
          autoComplete={props.autoComplete}
          onChange={onChangeHandler}
          onBlur={onBlurHandler}
        />
        <label
          className={`${
            isEmpty
              ? `left-2.5 scale-110 text-gray-400 transition-all peer-focus:left-1
                peer-focus:-translate-y-4 peer-focus:scale-90 peer-focus:text-primary`
              : `${
                isValid
                    ? `text-customtxt transition-all peer-focus:left-1 peer-focus:-translate-y-4
                      peer-focus:scale-90 peer-focus:text-primary`
                    : "text-red-600"
                } left-1 -translate-y-4 scale-90`
 
            } pointer-events-none absolute peer-autofill:left-0.5
            peer-autofill:-translate-y-4 peer-autofill:scale-90`}
          htmlFor={props.labelHtmlFor}
        >
          {props.labelText}
        </label>
      </div>
      {errorText && !isEmpty && (
        <span className="text-sm text-red-600">{errorText}</span>
      )}
    </>
  );
}
