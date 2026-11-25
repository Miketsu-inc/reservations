import { WEEKDAYS } from "@reservations/lib";

const DISPLAY_ORDER = [1, 2, 3, 4, 5, 6, 0];

function formatShifts(shifts) {
  if (!shifts || shifts.length === 0) return null;

  return shifts
    .map(
      (shift) =>
        `${shift.start_time.slice(0, 5)} - ${shift.end_time.slice(0, 5)}`
    )
    .join(" & ");
}

export default function BusinessHours({ hoursData }) {
  const todayIndex = new Date().getDay();

  return (
    <ul className="flex w-full flex-col gap-3">
      {DISPLAY_ORDER.map((dayIndex) => {
        const hours = formatShifts(hoursData?.[dayIndex]);
        const isClosed = hours === null;
        const textStyles = isClosed
          ? "text-fg-muted"
          : dayIndex === todayIndex
            ? "text-fg font-semibold"
            : "text-fg";

        return (
          <li key={dayIndex} className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <span
                className={`size-2 rounded-full
                ${isClosed ? "bg-bg-inactive" : "bg-success"}`}
              />
              <span className={textStyles}>{WEEKDAYS[dayIndex]}</span>
            </div>
            <span className={textStyles}>{hours ?? "Closed"}</span>
          </li>
        );
      })}
    </ul>
  );
}
