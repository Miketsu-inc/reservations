import { ViewIcon, ViewOffSlashIcon } from "@hugeicons/core-free-icons";
import { useAutofill } from "@reservations/lib";
import { useRef, useState } from "react";
import { Icon } from ".";

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

  const input = (
    <input
      className={`${styles} ${isTypePassword ? "pr-12" : ""} autofill w-full
        appearance-none rounded-lg ps-3 pe-3 outline-hidden dark:scheme-dark`}
      id={id}
      type={isTypePassword ? (visible ? "text" : type) : type}
      value={value}
      name={name}
      onChange={onChange}
      onBlur={onBlur}
      onFocus={onFocus}
      ref={inputRef}
      autoFocus={autoFocus}
      {...props}
    />
  );

  if (!isTypePassword) return input;

  return (
    <div className="relative w-full">
      {input}
      {isTypePassword ? (
        <button
          type="button"
          className="absolute top-1/2 right-4 -translate-y-1/2 cursor-pointer"
          onClick={() => setVisible(!visible)}
          onKeyDown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              setVisible(!visible);
            }
          }}
        >
          <Icon icon={ViewIcon} altIcon={ViewOffSlashIcon} showAlt={visible} />
        </button>
      ) : (
        <></>
      )}
    </div>
  );
}
