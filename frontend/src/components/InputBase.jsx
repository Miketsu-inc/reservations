import { useRef, useState } from "react";
import EyeIcon from "../assets/icons/EyeIcon";
import EyeSlashIcon from "../assets/icons/EyeSlashIcon";
import { useAutofill } from "../lib/hooks";

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
}) {
  const isTypePassword = type === "password";
  const [visible, setVisible] = useState(false);
  const inputRef = useRef();

  useAutofill(inputRef, onBlur);

  return (
    <>
      <input
        className={`${styles} ${isTypePassword ? "w-5/6" : "w-full"} autofill bg-transparent p-2
          outline-none dark:[color-scheme:dark]`}
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
