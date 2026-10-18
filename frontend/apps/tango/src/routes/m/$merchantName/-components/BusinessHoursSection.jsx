import { useMemo } from "react";

const daysOfWeek = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
];

export default function BusinessHoursSection({ hoursData }) {
  const today = daysOfWeek[new Date().getDay()];

  const businessHours = useMemo(() => {
    if (!hoursData) return {};

    const hours = {};
    daysOfWeek.forEach((day, index) => {
      const dailyShifts = hoursData[index];

      if (dailyShifts && dailyShifts.length > 0) {
        hours[day] = dailyShifts
          .map(
            (shift) =>
              `${shift.start_time.slice(0, 5)} - ${shift.end_time.slice(0, 5)}`
          )
          .join(" & ");
      } else {
        hours[day] = "Closed";
      }
    });
    return hours;
  }, [hoursData]);

  return (
    <ul className="flex w-full flex-col gap-3">
      {Object.entries(businessHours)
        .sort(([a], [b]) => {
          const order = [
            "Monday",
            "Tuesday",
            "Wednesday",
            "Thursday",
            "Friday",
            "Saturday",
            "Sunday",
          ];
          return order.indexOf(a) - order.indexOf(b);
        })
        .map(([day, hours]) => {
          const isToday = day === today;
          const isClosed = hours === "Closed";

          return (
            <li key={day} className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <span
                  className={`size-2 rounded-full ${
                    isClosed ? "bg-bg-inactive" : "bg-success"
                  }`}
                ></span>
                <span
                  className={`${
                    isClosed
                      ? "text-fg-muted"
                      : isToday
                        ? "text-fg font-semibold"
                        : "text-fg"
                  }`}
                >
                  {day}
                </span>
              </div>
              <span
                className={`${
                  isClosed
                    ? "text-fg-muted"
                    : isToday
                      ? "text-fg font-semibold"
                      : "text-fg"
                }`}
              >
                {hours}
              </span>
            </li>
          );
        })}
    </ul>
  );
}
