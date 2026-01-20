import { EditIcon, TrashBinIcon } from "@reservations/assets";

export default function TableActions({ onEdit, onDelete }) {
  return (
    <div className="flex h-full flex-row items-center justify-center">
      <button className="cursor-pointer" onClick={onEdit}>
        <EditIcon styles="size-4 mx-1" />
      </button>
      <button className="cursor-pointer" onClick={onDelete}>
        <TrashBinIcon styles="size-5 mx-1" />
      </button>
    </div>
  );
}
