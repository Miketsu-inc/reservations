import { Button } from "@reservations/components";

export default function DangerZoneItem({
  title,
  description,
  buttonText,
  onClick,
  action,
}) {
  return (
    <div
      className="flex flex-col items-start justify-between gap-3 sm:flex-row
        sm:items-center sm:gap-6"
    >
      <div className="flex flex-col">
        <span className="font-medium">{title}</span>
        <span className="text-text_color/65 mt-1 text-sm">{description}</span>
      </div>
      {action ?? (
        <Button
          onClick={onClick}
          variant="danger"
          styles="w-fit px-3 py-2 text-nowrap"
          buttonText={buttonText}
        />
      )}
    </div>
  );
}
