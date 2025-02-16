export default function BookingRing({ booked, cancelled }) {
  const strokeWidth = 7;
  const radius = 32 / 2 - strokeWidth / 2;
  const total = booked + cancelled;
  const circumference = 2 * Math.PI * radius;
  const bookedStroke = (booked / total) * circumference;
  const cancelledStroke = (cancelled / total) * circumference;

  return (
    <svg className="h-5 w-5 -rotate-90" viewBox="0 0 32 32">
      <circle
        className={`fill-none stroke-green-600 stroke-[${strokeWidth}]`}
        cx="16"
        cy="16"
        r={radius}
        strokeDasharray={`${bookedStroke} ${circumference - bookedStroke}`}
        strokeDashoffset="0"
      />
      <circle
        className={`fill-none stroke-red-600 stroke-[${strokeWidth}]`}
        cx="16"
        cy="16"
        r={radius}
        strokeDasharray={`${cancelledStroke} ${circumference - cancelledStroke}`}
        strokeDashoffset={`-${bookedStroke}`}
      />
    </svg>
  );
}
