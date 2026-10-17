export default function AvailableTimeSection({
  availableTimes,
  timeSection,
  onSelect,
  selectedStartsAt,
}) {
  return (
    <>
      {availableTimes.length > 0 ? (
        <div className="grid w-full grid-cols-2 gap-2 sm:grid-cols-5">
          {availableTimes.map((slot) => (
            <button
              key={`${timeSection}-${slot.starts_at}`}
              className={`cursor-pointer rounded-lg border py-2 text-sm ${
                selectedStartsAt === slot.starts_at
                  ? "bg-primary border-primary text-fg-on-primary"
                  : `bg-bg-raised border-border text-fg
                    hover:border-border-strong`
                }`}
              onClick={() => onSelect(slot)}
              type="button"
            >
              {slot.time}
            </button>
          ))}
        </div>
      ) : (
        <p className="text-fg-muted flex items-center justify-center text-sm">
          No available {timeSection} hours for this day
        </p>
      )}
    </>
  );
}
