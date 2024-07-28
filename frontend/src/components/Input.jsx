import { forwardRef, useImperativeHandle, useState } from "react";
import InputBase from "./InputBase";

// forwardRef needed for assessing ref
export default forwardRef(function Input(props, ref) {
  const [errorTriggered, setErrorTriggered] = useState(false);
  const [inputValue, setInputValue] = useState("");
  const [errorText, setErrorText] = useState("");
  const [isValid, setIsValid] = useState(false);

  const isEmpty = inputValue.trim() === "";

  function onChangeHandler(e) {
    setInputValue(e.target.value);
    setErrorTriggered(false);
    setErrorText("");
    setIsValid(true);
  }

  function onBlurHandler(e) {
    let valid = isValid;

    if (props.inputValidation(inputValue)) {
      setErrorText("");
      valid = true;
    } else {
      setErrorText(props.errorText);
      valid = false;
    }

    setIsValid(valid);

    props.inputData({
      name: props.name,
      value: inputValue,
      isValid: valid,
    });
  }

  // expose triggerErrorText function to Parent component
  useImperativeHandle(ref, () => ({
    triggerValidationError() {
      setErrorText(props.errorText);
      setErrorTriggered(true);
      setIsValid(false);
    },
  }));

  return (
    <>
      <div
        className={`${(isValid || isEmpty) && !errorTriggered ? "justify-between border-customtxt focus-within:border-primary" : "border-red-600 focus-within:border-red-600"}
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
            isEmpty && !errorTriggered
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
      {(errorText && !isEmpty) || errorTriggered ? (
        <span className="text-sm text-red-600">{errorText}</span>
      ) : (
        <></>
      )}
    </>
  );
});
