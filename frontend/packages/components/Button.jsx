import { LoaderIcon } from "@reservations/assets";
import { forwardRef } from "react";

const variants = {
  primary: `bg-primary text-fg-on-primary shadow-card border-transparent
    hover:not-disabled:bg-primary-hover active:not-disabled:bg-primary-active`,
  secondary: `bg-primary-subtle text-fg-primary border-transparent
    hover:not-disabled:bg-primary-subtle-hover`,
  tertiary: `bg-bg-input text-fg border-border-strong shadow-card
    hover:not-disabled:bg-hover`,
  danger: `bg-danger text-white shadow-card border-transparent
    hover:not-disabled:bg-danger-hover`,
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
      className={`${styles} ${variants[variant]} focus-visible:outline-primary
        flex items-center justify-center rounded-lg border
        focus-visible:outline-2 focus-visible:outline-offset-2
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
