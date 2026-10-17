import { Cancel01Icon } from "@hugeicons/core-free-icons";
import Icon from "./Icon.jsx";

export default function CloseButton({ styles, onClick, ...props }) {
  return (
    <button
      className="text-fg-muted hover:bg-hover hover:text-fg cursor-pointer
        rounded-lg p-1 transition-colors"
      name="close"
      type="button"
      onClick={onClick}
      aria-label="Close"
      {...props}
    >
      <Icon icon={Cancel01Icon} styles={`${styles} size-6`} />
    </button>
  );
}
