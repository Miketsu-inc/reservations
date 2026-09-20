import { LoaderIcon } from "@reservations/assets";
import { forwardRef } from "react";

const Button = forwardRef(function Button(
  {
    children,
    childSide = "left",
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
    primary:
      "bg-primary hover:not-disabled:bg-hvr_primary text-white shadow-md",
    secondary:
      "bg-transparent text-primary hover:not-disabled:text-hvr_primary border-2 border-primary hover:not-disabled:border-hvr_primary",
    tertiary:
      "bg-transparent hover:not-disabled:bg-gray-300 dark:hover:not-disabled:bg-gray-800 text-text_color shadow-none border-2 border-gray-300 dark:border-gray-800",
    danger:
      "dark:hover:not-disabled:bg-red-800 dark:bg-red-700 bg-red-500 hover:not-disabled:bg-red-600 text-white shadow-md",
  };

  return (
    <button
      ref={ref}
      onClick={onClick}
      className={`${styles} ${variants[variant]} flex items-center
        justify-center rounded-lg focus-visible:outline-1
        ${childSide == "right" ? "flex-row-reverse" : ""}
        ${isLoading || disabled ? "opacity-50 transition-opacity duration-300" : "cursor-pointer"}`}
      name={name}
      type={type}
      disabled={isLoading || disabled}
    >
      {isLoading ? (
        <>
          <span className="pr-4 pl-5">{buttonText}</span>
          <LoaderIcon styles="-ml-1 mr-3 h-5 w-5" />
        </>
      ) : children ? (
        <>
          <span>{children}</span>
          <span>{buttonText}</span>
        </>
      ) : (
        buttonText
      )}
    </button>
  );
});

export default Button;
