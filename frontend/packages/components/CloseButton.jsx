import { XIcon } from "@reservations/assets";

export default function CloseButton({ styles, onClick }) {
  return (
    <button
      className="hover:bg-hvr_gray cursor-pointer rounded-lg"
      name="close"
      type="button"
      onClick={onClick}
    >
      <XIcon styles={`${styles} size-8 fill-gray-700 dark:fill-gray-200`} />
    </button>
  );
}
