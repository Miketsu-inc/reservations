export default function AvailableTimeSection({
  availableTimes,
  timeSection,
  onSelect,
  selectedStartsAt,
}) {
  return (
    <>
      {availableTimes.length > 0 ? (
        <div className="grid w-full grid-cols-2 gap-3 rounded-md sm:grid-cols-5">
          {availableTimes.map((slot) => (
            <button
              key={`${timeSection}-${slot.starts_at}`}
              className={`bg-layer_bg border-border_color text-text_color
                cursor-pointer rounded-md border py-1.5 transition-all
                hover:bg-gray-50 dark:hover:bg-gray-200/5
                ${selectedStartsAt === slot.starts_at ? "ring-primary ring-2" : ""}`}
              onClick={() => onSelect(slot)}
              type="button"
            >
              {slot.time}
            </button>
          ))}
        </div>
      ) : (
        <p className="text-md flex items-center justify-center">
          No available {timeSection} hours for this day
        </p>
      )}
    </>
  );
}
