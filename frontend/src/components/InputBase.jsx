import { useState } from "react";
import EyeIcon from "../assets/EyeIcon";
import EyeSlashIcon from "../assets/EyeSlashIcon";

export default function InputBase(props) {
  const isTypePassword = props.type === "password";
  const [visible, setVisible] = useState(false);

  return (
    <>
      <input
        className={`${props.styles} ${isTypePassword ? "left-1 w-5/6 autofill:p-1" : "w-full"}
          bg-transparent p-2 text-customtxt outline-none`}
        // is this needed? wouldn't all non password inputs be text?
        type={isTypePassword ? (visible ? "text" : props.type) : props.type}
        value={props.value}
        name={props.name}
        id={props.id}
        autoComplete={props.autoComplete}
        onChange={props.onChange}
        onBlur={props.onBlur}
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
