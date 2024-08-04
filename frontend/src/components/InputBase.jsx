import { useState } from "react";
import EyeIcon from "../assets/EyeIcon";
import EyeSlashIcon from "../assets/EyeSlashIcon";

export default function InputBase({
  id,
  name,
  type,
  value,
  autoComplete,
  styles,
  onChange,
  onBlur,
}) {
  const isTypePassword = type === "password";
  const [visible, setVisible] = useState(false);

  return (
    <>
      <input
        className={`${styles} ${isTypePassword ? "left-1 w-5/6 autofill:p-1" : "w-full"}
          bg-transparent p-2 outline-none`}
        // is this needed? wouldn't all non password inputs be text?
        type={isTypePassword ? (visible ? "text" : type) : type}
        value={value}
        name={name}
        id={id}
        autoComplete={autoComplete}
        onChange={onChange}
        onBlur={onBlur}
      />
      {isTypePassword ? (
        <div>
          {visible ? (
            <EyeSlashIcon
              onClick={() => {
                setVisible(!visible);
              }}
              styles="fill-customtxt absolute -translate-y-1/2 right-4"
            />
          ) : (
            <EyeIcon
              onClick={() => {
                setVisible(!visible);
              }}
              styles="fill-customtxt absolute -translate-y-1/2 right-4"
            />
          )}
        </div>
      ) : (
        <></>
      )}
    </>
  );
}
