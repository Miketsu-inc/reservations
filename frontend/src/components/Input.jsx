import { useState } from "react";
import EyeIcon from "../assets/EyeIcon";
import EyeSlashIcon from "../assets/EyeSlashIcon";

export default function Input(props) {
  const isTypePassword = props.type === "password";
  const [visible, setVisible] = useState(false);

  return (
    <>
      {isTypePassword ? (
        <>
          <input
            className={`${props.styles} left-1 w-5/6 bg-transparent p-2 text-customtxt outline-none
              autofill:p-1`}
            type={visible ? "text" : props.type}
            value={props.value}
            name={props.name}
            autoComplete={props.autoComplete}
            minLength={props.minLength}
            id={props.id}
            onChange={props.onChange}
            onBlur={props.onBlur}
          />
          <div>
            {visible ? (
              <EyeSlashIcon
                onClick={() => {
                  setVisible(!visible);
                }}
                styles={"fill-customtxt absolute -translate-y-1/2 right-4"}
                width={"20"}
                height={"20"}
                role={"button"}
              />
            ) : (
              <EyeIcon
                onClick={() => {
                  setVisible(!visible);
                }}
                styles={"fill-customtxt absolute -translate-y-1/2 right-4"}
                width={"20"}
                height={"20"}
                role={"button"}
              />
            )}
          </div>
        </>
      ) : (
        <input
          className={`${props.styles} w-full bg-transparent p-2 text-customtxt focus:outline-none`}
          name={props.name}
          aria-label={props.ariaLabel}
          type={props.type}
          value={props.value}
          autoComplete={props.autoComplete}
          minLength={props.minLength}
          id={props.id}
          onChange={props.onChange}
          onBlur={props.onBlur}
        />
      )}
    </>
  );
}
