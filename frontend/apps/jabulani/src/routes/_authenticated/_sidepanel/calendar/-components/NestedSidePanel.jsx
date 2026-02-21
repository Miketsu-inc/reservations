import { BackArrowIcon } from "@reservations/assets";

export default function NestedSidePanel({ onBack, isOpen, children, styles }) {
  return (
    <div
      className={`bg-layer_bg absolute inset-0 z-20 flex flex-col
        transition-transform duration-300 ease-in-out
        ${isOpen ? "translate-x-0" : "translate-x-full"}`}
    >
      <button
        className="mx-5 my-5 w-fit cursor-pointer rounded-lg p-1
          hover:bg-gray-400/20"
        type="button"
        onClick={onBack}
      >
        <BackArrowIcon styles={`${styles} stroke-text_color`} />
      </button>
      <div className="flex flex-1 flex-col overflow-hidden">{children}</div>
    </div>
  );
}
