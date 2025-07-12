import EyeIcon from "@icons/EyeIcon";
import EyeSlashIcon from "@icons/EyeSlashIcon";
import { useAutofill } from "@lib/hooks";
import { useRef, useState } from "react";

export default function InputBase({
  id,
  name,
  type,
  value,
  styles,
  onChange,
  onBlur,
  onFocus,
  autoFocus = false,
  ...props
}) {
  const isTypePassword = type === "password";
  const [visible, setVisible] = useState(false);
  const inputRef = useRef();

  useAutofill(inputRef, onBlur);

  return (
    <>
      <input
        className={`${styles} ${isTypePassword ? "w-5/6" : "w-full"} autofill appearance-none
          rounded-lg ps-3 pe-3 outline-hidden dark:[color-scheme:dark]`}
        id={id}
        type={isTypePassword ? (visible ? "text" : type) : type}
        value={value}
        name={name}
        onChange={onChange}
        onBlur={onBlur}
        onFocus={onFocus}
        ref={inputRef}
        {...props}
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
