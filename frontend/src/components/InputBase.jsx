import EyeIcon from "@icons/EyeIcon";
import EyeSlashIcon from "@icons/EyeSlashIcon";
import { useAutofill } from "@lib/hooks";
import { useRef, useState } from "react";

export default function InputBase({
  id,
  name,
  type,
  value,
  autoComplete,
  styles,
  onChange,
  onBlur,
  onFocus,
  placeholder,
  pattern,
  required,
  min,
  max,
  autoFocus = false,
}) {
  const isTypePassword = type === "password";
  const [visible, setVisible] = useState(false);
  const inputRef = useRef();

  useAutofill(inputRef, onBlur);

  return (
    <>
      <input
        className={`${styles} ${isTypePassword ? "w-5/6" : "w-full"} autofill bg-transparent
          outline-hidden`}
        type={isTypePassword ? (visible ? "text" : type) : type}
        value={value}
        name={name}
        id={id}
        autoComplete={autoComplete}
        pattern={pattern}
        required={required}
        onChange={onChange}
        onBlur={onBlur}
        onFocus={onFocus}
        placeholder={placeholder}
        min={min}
        max={max}
        autoFocus={autoFocus}
        ref={inputRef}
      />
      {isTypePassword ? (
        <div>
          {visible ? (
            <EyeSlashIcon
              onClick={() => {
                setVisible(!visible);
              }}
              styles="absolute -translate-y-1/2 right-4"
            />
          ) : (
            <EyeIcon
              onClick={() => {
                setVisible(!visible);
              }}
              styles="absolute -translate-y-1/2 right-4"
            />
          )}
        </div>
      ) : (
        <></>
      )}
    </>
  );
}
