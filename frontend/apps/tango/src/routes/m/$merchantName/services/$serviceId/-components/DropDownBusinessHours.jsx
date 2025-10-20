import { BackArrowIcon, ClockIcon } from "@reservations/assets";
import { useMemo, useState } from "react";

const daysOfWeek = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
];

function parseTimeString(timeStr) {
  const [hours, minutes] = timeStr.split(":").map(Number);
  // if we implement business hours past midnight the date should set to +1 for the mext day's closing time
  const date = new Date();
  date.setHours(hours, minutes, 0, 0);
  return date;
}

function findNextOpenDay(businessHours, startIndex) {
  for (let i = 1; i <= 7; i++) {
    const nextDayIndex = (startIndex + i) % 7;
    const nextDay = daysOfWeek[nextDayIndex];
    const nextDayHours = businessHours[nextDay];
    if (nextDayHours !== "Closed") {
      return {
        day: nextDay,
        time: nextDayHours.split("–")[0].trim(),
      };
    }
  }
}

function calculateBusinessStatus(businessHours) {
  const todayIndex = new Date().getDay();
  const today = daysOfWeek[todayIndex];
  const todayHours = businessHours[today];

  const now = new Date();
  let isBusinessOpen = false;
  let closeTimeStr = null;
  let nextOpenTime = null;
  let nextOpenDay = null;
  let closingSoon = false;

  if (todayHours !== "Closed") {
    const [openStr, closeStr] = todayHours.split("–").map((s) => s.trim());
    closeTimeStr = closeStr;

    const openTime = parseTimeString(openStr);
    const closeTime = parseTimeString(closeStr);

    isBusinessOpen = now >= openTime && now < closeTime;

    const minutesUntilClose = (closeTime - now) / 1000 / 60;
    closingSoon = isBusinessOpen && minutesUntilClose <= 30;

    if (!isBusinessOpen) {
      if (now < openTime) {
        nextOpenTime = openStr;
        nextOpenDay = "today";
      } else {
        const nextOpen = findNextOpenDay(businessHours, todayIndex);
        nextOpenTime = nextOpen.time;
        nextOpenDay = nextOpen.day;
      }
    }
  } else {
    const nextOpen = findNextOpenDay(businessHours, todayIndex);
    nextOpenTime = nextOpen.time;
    nextOpenDay = nextOpen.day;
  }

  return {
    isBusinessOpen,
    closeTimeStr,
    nextOpenTime,
    nextOpenDay,
    closingSoon,
    today,
  };
}

export default function DropDownBusinessHours({ hoursData }) {
  const [isDropDownOpen, setIsDropDownOpen] = useState(false);

  const businessHours = useMemo(() => {
    const hours = {};
    daysOfWeek.forEach((day, index) => {
      const data = hoursData[index];
      if (data) {
        hours[day] =
          `${data.start_time.slice(0, 5)} – ${data.end_time.slice(0, 5)}`;
      } else {
        hours[day] = "Closed";
      }
    });
    return hours;
  }, [hoursData]);

  const {
    isBusinessOpen,
    closeTimeStr,
    nextOpenTime,
    nextOpenDay,
    closingSoon,
    today,
  } = useMemo(() => {
    return calculateBusinessStatus(businessHours);
  }, [businessHours]);

  return (
    <>
      <div className="flex w-full flex-col">
        <button
          onClick={() => setIsDropDownOpen(!isDropDownOpen)}
          className="flex w-full items-center justify-between"
        >
          <div className="flex items-center gap-3">
            <ClockIcon styles="fill-text_color size-5" />
            {isBusinessOpen ? (
              <span className="flex gap-1 text-base font-medium">
                <span
                  className="font-semibold text-green-600 dark:text-green-400"
                >
                  Open
                </span>
                – {closingSoon ? "closes soon at" : "closes at"}
                <span>{closeTimeStr}</span>
              </span>
            ) : (
              <span className="flex gap-1 text-base font-medium text-nowrap">
                <span className="text-orange-700 dark:text-orange-500">
                  Closed
                </span>
                – opens{" "}
                {nextOpenDay === "today"
                  ? `later today at ${nextOpenTime}`
                  : `on ${nextOpenDay} at ${nextOpenTime}`}
              </span>
            )}
          </div>
          <BackArrowIcon
            styles={`size-6 stroke-gray-600 transition-transform duration-200
              dark:text-gray-400 ${isDropDownOpen ? "rotate-90" : "-rotate-90"}`}
          />
        </button>

        <div
          className={`transition-[max-height,opacity] duration-200 ease-in-out
            ${
              isDropDownOpen
                ? "max-h-[1000px] opacity-100"
                : "max-h-0 overflow-hidden opacity-0"
            }`}
        >
          <ul className="mt-5 flex flex-col gap-3 text-base">
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
                        className={`size-3 rounded-full ${
                          isClosed
                            ? "bg-gray-400 dark:bg-gray-600"
                            : "bg-green-500 dark:bg-green-400"
                        }`}
                      ></span>
                      <span
                        className={`${
                          isClosed
                            ? "text-gray-500 dark:text-gray-400"
                            : isToday
                              ? "font-semibold"
                              : "text-text_color"
                        }`}
                      >
                        {day}
                      </span>
                    </div>
                    <span
                      className={`${
                        isClosed
                          ? "text-gray-500 dark:text-gray-400"
                          : isToday
                            ? "font-semibold"
                            : "text-text_color"
                      }`}
                    >
                      {hours}
                    </span>
                  </li>
                );
              })}
          </ul>
        </div>
      </div>
    </>
  );
}
