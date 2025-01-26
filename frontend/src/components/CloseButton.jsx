import XIcon from "@icons/XIcon";

export default function CloseButton({ styles, onClick }) {
  return (
    <button
      className="cursor-pointer rounded-lg hover:bg-hvr_gray"
      name="close"
      type="button"
      onClick={onClick}
    >
      <XIcon styles={`${styles} w-8 h-8 fill-gray-700 dark:fill-gray-200`} />
    </button>
  );
}
