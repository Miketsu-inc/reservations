import LoaderIcon from "../assets/icons/LoaderIcon";

export default function Button({
  children,
  name,
  type,
  styles,
  onClick,
  buttonText,
  isLoading,
}) {
  return (
    <button
      onClick={onClick}
      className={`${styles} rounded-lg bg-primary py-2 font-medium shadow-md
        ${isLoading ? "opacity-50 transition-opacity" : ""}`}
      name={name}
      type={type}
      disabled={isLoading ? true : false}
    >
      {isLoading ? (
        <div className="flex items-center justify-center">
          <span className="pl-5 pr-4">{buttonText}</span>
          <LoaderIcon styles="-ml-1 mr-3 h-5 w-5" />
        </div>
      ) : children ? (
        <div className="flex items-center justify-center">
          <span className="pr-3">{children}</span>
          <span>{buttonText}</span>
        </div>
      ) : (
        buttonText
      )}
    </button>
  );
}
