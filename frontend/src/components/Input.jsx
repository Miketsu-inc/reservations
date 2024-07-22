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
            className={`${props.styles} p-2 left-1 w-5/6 bg-transparent outline-none`}
            type={visible ? "text" : props.type}
            value={props.value}
            name={props.name}
            required={props.required}
            autoComplete={props.autoComplete}
            minLength={props.minLength}
            id={props.id}
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
          className={`${props.styles} p-2 w-full bg-transparent focus:outline-none`}
          name={props.name}
          aria-label={props.ariaLabel}
          type={props.type}
          value={props.value}
          required={props.required}
          autoComplete={props.autoComplete}
          minLength={props.minLength}
          id={props.id}
        />
      )}
    </>
  );
}
