import { LoaderIcon } from "@reservations/assets";
import React from "react";

const Button = React.forwardRef(function Button(
  {
    children,
    name,
    type,
    styles,
    onClick,
    buttonText,
    isLoading,
    disabled,
    variant = "primary",
  },
  ref
) {
  const variants = {
    primary: "bg-primary hover:bg-hvr_primary text-white shadow-md",
    secondary:
      "bg-transparent text-primary hover:text-hvr_primary border-2 border-primary hover:border-hvr_primary",
    tertiary:
      "bg-transparent hover:bg-gray-300 dark:hover:bg-gray-800 text-text_color shadow-none border-2 border-gray-300 dark:border-gray-800",
    danger:
      "dark:hover:bg-red-800 dark:bg-red-700 bg-red-500 hover:bg-red-600 text-white shadow-md",
  };

  return (
    <button
      ref={ref}
      onClick={onClick}
      className={`${styles} ${variants[variant]} rounded-lg
        focus-visible:outline-1
        ${isLoading || disabled ? "opacity-50 transition-opacity" : "cursor-pointer"}`}
      name={name}
      type={type}
      disabled={isLoading || disabled}
    >
      {isLoading ? (
        <div className="flex items-center justify-center">
          <span className="pr-4 pl-5">{buttonText}</span>
          <LoaderIcon styles="-ml-1 mr-3 h-5 w-5" />
        </div>
      ) : children ? (
        <div className="flex items-center justify-center">
          <span>{children}</span>
          <span>{buttonText}</span>
        </div>
      ) : (
        buttonText
      )}
    </button>
  );
});

export default Button;
