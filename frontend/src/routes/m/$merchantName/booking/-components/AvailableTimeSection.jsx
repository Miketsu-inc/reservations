export default function AvailableTimeSection({
  availableTimes,
  timeSection,
  clickedHour,
  selectedHour,
}) {
  return (
    <>
      {availableTimes.length > 0 ? (
        <div className="grid w-full grid-cols-2 gap-3 rounded-md sm:grid-cols-5">
          {availableTimes.map((hour, index) => (
            <button
              key={`${timeSection}-${index}`}
              className={`bg-accent/90 hover:bg-accent/80 cursor-pointer rounded-md py-1 font-bold
                text-black transition-all ${selectedHour === hour ? "ring-2 ring-blue-500" : ""}`}
              onClick={clickedHour}
              value={hour}
              type="button"
            >
              {hour}
            </button>
          ))}
        </div>
      ) : (
        <p className="text-md flex items-center justify-center font-bold">
          No available {timeSection} hours for this day
        </p>
      )}
    </>
  );
}
