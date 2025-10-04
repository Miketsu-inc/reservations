import Button from "@components/Button";

export default function DangerZoneItem({
  title,
  description,
  buttonText,
  onClick,
}) {
  return (
    <div
      className="flex flex-col items-start justify-between gap-3 sm:flex-row
        sm:items-center sm:gap-0"
    >
      <div className="flex flex-col">
        <span className="font-semibold">{title}</span>
        <span className="text-text_color/70">{description}</span>
      </div>
      <Button
        onClick={onClick}
        variant="danger"
        styles="py-1 px-2 w-fit"
        buttonText={buttonText}
      />
    </div>
  );
}
