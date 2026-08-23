import { ViewIcon, ViewOffSlashIcon } from "@hugeicons/core-free-icons";
import { useAutofill } from "@reservations/lib";
import { useRef, useState } from "react";
import Icon from "./Icon.jsx";

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

// TODO: Temporary solution to avoid import cycles until we decide wether to
// get rid of FloatingLabelInput and put these stylings on InputBase
export function StyledInputBase({ styles = "", ...props }) {
  return (
    <InputBase
      {...props}
      styles={`${styles} peer border bg-layer_bg outline-hidden
        placeholder-stone-500 dark:placeholder-zinc-400
        transition-[border-color,box-shadow] ease-in-out duration-150
        border-input_border_color focus:border-primary focus:ring-4
        focus:ring-primary/30 disabled:text-text_color/70
        disabled:border-input_border_color/60 disabled:bg-gray-200/60
        disabled:dark:bg-gray-700/20 p-2`}
    />
  );
}
