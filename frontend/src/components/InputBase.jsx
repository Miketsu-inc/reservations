import { useRef, useState } from "react";
import EyeIcon from "../assets/EyeIcon";
import EyeSlashIcon from "../assets/EyeSlashIcon";
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
}) {
  const isTypePassword = type === "password";
  const [visible, setVisible] = useState(false);
  const inputRef = useRef();

  useAutofill(inputRef, onBlur);

  return (
    <>
      <input
        className={`${styles} ${isTypePassword ? "w-5/6" : "w-full"} bg-transparent p-2 outline-none
          dark:[color-scheme:dark]`}
        // is this needed? wouldn't all non password inputs be text?
        type={isTypePassword ? (visible ? "text" : type) : type}
        value={value}
        name={name}
        id={id}
        autoComplete={autoComplete}
        onChange={onChange}
        onBlur={onBlur}
        onFocus={onFocus}
        ref={inputRef}
      />
      {isTypePassword ? (
        <div>
          {visible ? (
            <EyeSlashIcon
              onClick={() => {
                setVisible(!visible);
              }}
              styles="fill-text_color absolute -translate-y-1/2 right-4"
            />
          ) : (
            <EyeIcon
              onClick={() => {
                setVisible(!visible);
              }}
              styles="fill-text_color absolute -translate-y-1/2 right-4"
            />
          )}
        </div>
      ) : (
        <></>
      )}
    </>
  );
}
