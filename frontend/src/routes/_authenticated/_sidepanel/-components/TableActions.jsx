import EditIcon from "@icons/EditIcon";
import TrashBinIcon from "@icons/TrashBinIcon";

export default function TableActions({ onEdit, onDelete }) {
  return (
    <div className="flex h-full flex-row items-center justify-center gap-2">
      <EditIcon onClick={onEdit} styles="cursor-pointer w-4 h-4" />
      <TrashBinIcon onClick={onDelete} styles="cursor-pointer w-5 h-5" />
    </div>
  );
}
