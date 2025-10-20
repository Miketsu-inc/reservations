import { EditIcon, TrashBinIcon } from "@reservations/assets";

export default function TableActions({ onEdit, onDelete }) {
  return (
    <div className="flex h-full flex-row items-center justify-center">
      <button className="cursor-pointer" onClick={onEdit}>
        <EditIcon styles="w-4 h-4 mx-1" />
      </button>
      <button className="cursor-pointer" onClick={onDelete}>
        <TrashBinIcon styles="w-5 h-5 mx-1" />
      </button>
    </div>
  );
}
