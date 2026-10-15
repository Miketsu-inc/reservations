import { useState } from "react";

export default function Switch({
  size = "medium",
  defaultValue = false,
  variant = "default",
  disabled = false,
  onSwitch,
  styles = "",
  ...props
}) {
  const [isOn, setIsOn] = useState(defaultValue);

  return (
    <button
      aria-checked={isOn}
      role="switch"
      type="button"
      disabled={disabled}
      onClick={() => {
        if (!disabled) {
          setIsOn(!isOn);
          onSwitch(!isOn);
        }
      }}
      className={`${
        isOn
          ? `${variant === "monochrome" ? "bg-fg" : "bg-primary"}`
          : "bg-bg-inactive"
        }
        ${size === "small" ? "pr-4" : size === "medium" ? "pr-5" : size === "large" ? "pr-6" : ""}
        ${styles} outline-border-strong w-fit cursor-pointer rounded-full py-1
        pl-1 transition-colors duration-100 focus-visible:outline-2
        disabled:cursor-not-allowed disabled:opacity-50`}
      {...props}
    >
      <div
        className={`${isOn ? `${size === "small" ? "translate-x-3" : size === "medium" ? "translate-x-4" : size === "large" ? "translate-x-5" : ""}` : "translate-x-0"}
          ${size === "small" ? "size-3" : size === "medium" ? "size-4" : size === "large" ? "size-5" : ""}
          bg-bg pointer-events-none rounded-full transition-transform
          duration-100 outline-none`}
      ></div>
    </button>
  );
}
