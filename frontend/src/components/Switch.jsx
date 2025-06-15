import { useEffect, useState } from "react";

export default function Switch({
  size = "medium",
  defaultValue = false,
  variant = "default",
  disabled = false,
  onSwitch,
}) {
  const [isOn, setIsOn] = useState(defaultValue);

  useEffect(() => {
    setIsOn(defaultValue);
  }, [defaultValue]);

  return (
    <button
      aria-checked={isOn}
      role="switch"
      type="button"
      onClick={() => {
        if (!disabled) {
          setIsOn(!isOn);
          onSwitch(isOn);
        }
      }}
      className={`${isOn ? `${variant === "monochrome" ? "bg-black dark:bg-white" : "bg-primary"}` : "bg-gray-300 dark:bg-gray-800"}
        ${size === "small" ? "pr-4" : size === "medium" ? "pr-5" : size === "large" ? "pr-6" : ""}
        w-fit cursor-pointer rounded-full py-1 pl-1 outline-gray-400 transition-colors
        duration-100 focus-visible:outline-2 dark:outline-white`}
    >
      <div
        className={`${isOn ? `${size === "small" ? "translate-x-3" : size === "medium" ? "translate-x-4" : size === "large" ? "translate-x-5" : ""}` : "translate-x-0"}
          ${size === "small" ? "size-3" : size === "medium" ? "size-4" : size === "large" ? "size-5" : ""}
          pointer-events-none rounded-full bg-white transition-transform duration-100
          outline-none dark:bg-black`}
      ></div>
    </button>
  );
}
