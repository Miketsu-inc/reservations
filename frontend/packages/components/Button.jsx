import { LoaderIcon } from "@reservations/assets";
import { forwardRef } from "react";

const variants = {
  primary:
    "bg-primary hover:not-disabled:bg-primary-hover text-white shadow-md",
  secondary:
    "bg-transparent text-fg-primary hover:not-disabled:text-primary-hover border-2 border-primary hover:not-disabled:border-primary-hover",
  tertiary:
    "bg-transparent hover:not-disabled:bg-hover text-fg shadow-none border-2 border-border",
  danger: " bg-danger hover:not-disabled:bg-danger-hover text-white shadow-md",
};

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
    ...props
  },
  ref
) {
  return (
    <button
      ref={ref}
      onClick={onClick}
      className={`${styles} ${variants[variant]} flex items-center
        justify-center rounded-lg focus-visible:outline-1
        ${childSide === "right" ? "flex-row-reverse" : ""}
        ${isLoading || disabled ? "opacity-50 transition-opacity duration-300" : "cursor-pointer"}`}
      name={name}
      type={type}
      disabled={isLoading || disabled}
      aria-busy={isLoading || undefined}
      {...props}
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
