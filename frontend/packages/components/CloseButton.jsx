import { Cancel01Icon } from "@hugeicons/core-free-icons";
import { Icon } from ".";

export default function CloseButton({ styles, onClick }) {
  return (
    <button
      className="hover:bg-hvr_gray cursor-pointer rounded-lg"
      name="close"
      type="button"
      onClick={onClick}
    >
      <Icon
        icon={Cancel01Icon}
        styles={`${styles} size-8 text-gray-700 dark:text-gray-200`}
      />
    </button>
  );
}
