import { forwardRef, useImperativeHandle, useState } from "react";
import InputBase from "./InputBase";

export default forwardRef(function FloatingLabelInput(
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
  const [isInputFocused, setIsInputFocused] = useState(false);

  const isEmpty = inputValue.trim() === "";

  function onChangeHandler(e) {
    setInputValue(e.target.value);
    setErrorTriggered(false);
    setIsValid(true);
  }

  function onBlurHandler(e) {
    if (!e.detail?.isAutofillEvent) {
      setIsInputFocused(false);
    }
    let valid = isValid;

    if (inputValidation(e.target.value)) {
      valid = true;
      setErrorTriggered(false);
    } else {
      valid = false;
    }

    setIsValid(valid);

    inputData({
      name: name,
      value: e.target.value,
      isValid: valid,
    });
  }

  useImperativeHandle(ref, () => ({
    triggerValidationError() {
      setErrorTriggered(true);
      setIsValid(false);
    },
  }));

  return (
    <>
      <div
        className={`relative flex w-full items-center border-2 ${styles} ${
          errorTriggered || (!isEmpty && !isValid)
            ? "border-red-600"
            : isInputFocused
              ? "border-primary"
              : "border-text_color"
          } `}
      >
        <InputBase
          styles="peer mt-1 pt-4"
          type={type}
          value={inputValue}
          name={name}
          id={id}
          autoComplete={autoComplete}
          onChange={onChangeHandler}
          onBlur={onBlurHandler}
          onFocus={() => {
            setIsInputFocused(true);
          }}
        />
        <label
          className={`${
            isEmpty && !errorTriggered
              ? `left-2 text-lg text-gray-400 peer-focus:left-2 peer-focus:-translate-y-4
                peer-focus:text-sm peer-focus:text-primary`
              : `${
                isValid
                    ? `peer-focus:left-2 peer-focus:-translate-y-4 peer-focus:text-sm
                      peer-focus:text-primary`
                    : "text-red-600"
                } left-2 -translate-y-4 text-sm`
 
            } pointer-events-none absolute transition-all peer-autofill:left-2
            peer-autofill:-translate-y-4 peer-autofill:text-sm`}
          htmlFor={labelHtmlFor}
        >
          {labelText}
        </label>
      </div>
      {(!isValid && !isEmpty) || errorTriggered ? (
        <span className="text-nowrap text-sm text-red-600">{errorText}</span>
      ) : (
        <></>
      )}
    </>
  );
});
