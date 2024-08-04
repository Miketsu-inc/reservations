import { forwardRef, useImperativeHandle, useState } from "react";
import InputBase from "./InputBase";

// forwardRef needed for assessing ref
export default forwardRef(function Input(
  {
    id,
    name,
    type,
    autoComplete,
    labelText,
    labelHtmlFor,
    styles,
    errorText,
    inputValidation,
    inputData,
  },
  ref
) {
  const [errorTriggered, setErrorTriggered] = useState(false);
  const [inputValue, setInputValue] = useState("");
  const [isValid, setIsValid] = useState(false);

  const isEmpty = inputValue.trim() === "";

  function onChangeHandler(e) {
    setInputValue(e.target.value);
    setErrorTriggered(false);
    setIsValid(true);
  }

  function onBlurHandler(e) {
    let valid = isValid;

    if (inputValidation(inputValue)) {
      valid = true;
    } else {
      valid = false;
    }

    setIsValid(valid);

    inputData({
      name: name,
      value: inputValue,
      isValid: valid,
    });
  }

  // expose triggerErrorText function to Parent component
  useImperativeHandle(ref, () => ({
    triggerValidationError() {
      setErrorTriggered(true);
      setIsValid(false);
    },
  }));

  return (
    <>
      <div
        className={`${(isValid || isEmpty) && !errorTriggered ? "justify-between focus-within:border-primary" : "border-red-600 focus-within:border-red-600"}
          relative mt-6 flex w-full items-center border-2 focus-within:outline-none`}
      >
        <InputBase
          styles={`${styles} peer mt-4`}
          type={type}
          value={inputValue}
          name={name}
          id={id}
          autoComplete={autoComplete}
          onChange={onChangeHandler}
          onBlur={onBlurHandler}
        />
        <label
          className={`${
            isEmpty && !errorTriggered
              ? `left-2.5 scale-110 text-gray-400 transition-all peer-focus:left-1
                peer-focus:-translate-y-4 peer-focus:scale-90 peer-focus:text-primary`
              : `${
                isValid
                    ? `transition-all peer-focus:left-1 peer-focus:-translate-y-4 peer-focus:scale-90
                      peer-focus:text-primary`
                    : "text-red-600"
                } left-1 -translate-y-4 scale-90`
 
            } pointer-events-none absolute peer-autofill:left-0.5
            peer-autofill:-translate-y-4 peer-autofill:scale-90`}
          htmlFor={labelHtmlFor}
        >
          {labelText}
        </label>
      </div>
      {(!isValid && !isEmpty) || errorTriggered ? (
        <span className="text-sm text-red-600">{errorText}</span>
      ) : (
        <></>
      )}
    </>
  );
});
