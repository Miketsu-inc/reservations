import { Cancel01Icon } from "@hugeicons/core-free-icons";
import Icon from "./Icon.jsx";

export default function CloseButton({ styles, onClick, ...props }) {
  return (
    <button
      className="hover:bg-hover cursor-pointer rounded-lg"
      name="close"
      type="button"
      onClick={onClick}
      aria-label="Close"
      {...props}
    >
      <Icon icon={Cancel01Icon} styles={`${styles} size-8 text-fg-muted`} />
    </button>
  );
}
