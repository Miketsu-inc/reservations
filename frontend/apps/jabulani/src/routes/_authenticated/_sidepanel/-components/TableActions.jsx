import { Delete02Icon, Edit03Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";

export default function TableActions({ onEdit, onDelete }) {
  return (
    <div className="flex h-full flex-row items-center justify-center gap-0.5">
      <button
        className="text-fg-muted hover:bg-hover hover:text-fg cursor-pointer
          rounded-lg p-2 transition-colors"
        onClick={onEdit}
      >
        <Icon icon={Edit03Icon} styles="size-4.5" />
      </button>
      <button
        className="hover:bg-danger-subtle cursor-pointer rounded-lg p-2
          transition-colors"
        onClick={onDelete}
      >
        <Icon icon={Delete02Icon} styles="size-4.5 text-fg-danger" />
      </button>
    </div>
  );
}
