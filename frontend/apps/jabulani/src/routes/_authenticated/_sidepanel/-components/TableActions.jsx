import { Delete02Icon, Edit03Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";

export default function TableActions({ onEdit, onDelete }) {
  return (
    <div className="flex h-full flex-row items-center justify-center gap-2">
      <button className="cursor-pointer" onClick={onEdit}>
        <Icon icon={Edit03Icon} styles="size-5" />
      </button>
      <button className="cursor-pointer" onClick={onDelete}>
        <Icon
          icon={Delete02Icon}
          styles="size-5 text-red-600 dark:text-red-500"
        />
      </button>
    </div>
  );
}
