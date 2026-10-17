import { forwardRef, useImperativeHandle, useState } from "react";
import InputBase from "./InputBase";

export default forwardRef(function FloatingLabelInput(
  {
    id,
    name,
    labelText,
    styles,
    errorText,
    inputValidation,
    inputData,
    ...props
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

    if (inputValidation?.(e.target.value)) {
      valid = true;
      setErrorTriggered(false);
    } else {
      valid = false;
    }

    setIsValid(valid);

    inputData?.({
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
        className={`bg-bg-input relative flex w-full items-center rounded-lg
          border transition-[border-color,box-shadow] duration-150 ${styles} ${
            errorTriggered || (!isEmpty && !isValid)
              ? "border-danger ring-danger-subtle ring-3"
              : isInputFocused
                ? "border-primary ring-primary-focus ring-3"
                : "border-border-strong hover:border-fg-muted"
          } `}
      >
        <InputBase
          styles="peer mt-1 pt-4 p-2"
          id={id}
          value={inputValue}
          name={name}
          onChange={onChangeHandler}
          onBlur={onBlurHandler}
          onFocus={() => {
            setIsInputFocused(true);
          }}
          {...props}
        />
        <label
          className={`${
            isEmpty && !errorTriggered
              ? `peer-focus:text-fg-primary text-fg-muted left-3
                peer-focus:left-3 peer-focus:-translate-y-4 peer-focus:text-sm `
              : `${
                  isValid
                    ? `peer-focus:text-fg-primary peer-focus:left-3
                      peer-focus:-translate-y-4 peer-focus:text-sm`
                    : "text-fg-danger"
                } left-3 -translate-y-4 text-sm`
            } pointer-events-none absolute transition-all peer-autofill:left-3
            peer-autofill:-translate-y-4 peer-autofill:text-sm`}
          htmlFor={id}
        >
          {labelText}
        </label>
      </div>
      {(!isValid && !isEmpty) || errorTriggered ? (
        <span className="text-fg-danger text-sm text-nowrap">{errorText}</span>
      ) : (
        <></>
      )}
    </>
  );
});
