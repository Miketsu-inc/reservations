export function formatToDateString(date) {
  return (
    date.getFullYear() +
    "-" +
    String(date.getMonth() + 1).padStart(2, "0") +
    "-" +
    String(date.getDate()).padStart(2, "0")
  );
}

export function isoToDateString(dateStr) {
  return dateStr.split("T")[0];
}

export function calculateStartEndTime(view, firstDayOfWeek) {
  if (!firstDayOfWeek) {
    firstDayOfWeek = "Monday";
  }

  const now = new Date();
  const offset = firstDayOfWeek === "Monday" ? 1 : 0;
  let start, end;

  switch (view) {
    case "dayGridMonth": {
      const monthStart = new Date(now.getFullYear(), now.getMonth(), 1);
      start = new Date(
        monthStart.getFullYear(),
        monthStart.getMonth(),
        offset - (monthStart.getDay() - 1)
      );
      const monthEnd = new Date(
        monthStart.getFullYear(),
        monthStart.getMonth() + 1,
        0
      );
      end = new Date(
        monthEnd.getFullYear(),
        monthEnd.getMonth(),
        monthEnd.getDate() + (7 - monthEnd.getDay() + offset)
      );
      break;
    }
    case "timeGridWeek":
    case "listWeek":
      if (now.getDay() === 0) {
        if (firstDayOfWeek === "Monday") {
          end = new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1);
          start = new Date(
            end.getFullYear(),
            end.getMonth(),
            end.getDate() - 7
          );
        } else if (firstDayOfWeek === "Sunday") {
          start = now;
          end = new Date(
            start.getFullYear(),
            start.getMonth(),
            start.getDate() + 7
          );
        }
      } else {
        start = new Date(
          now.getFullYear(),
          now.getMonth(),
          now.getDate() + offset - now.getDay()
        );
        end = new Date(
          start.getFullYear(),
          start.getMonth(),
          start.getDate() + 7
        );
      }
      break;
    case "timeGridDay":
      start = now;
      end = new Date(
        start.getFullYear(),
        start.getMonth(),
        start.getDate() + 1
      );
      break;
  }

  return {
    start: formatToDateString(start),
    end: formatToDateString(end),
  };
}

export function isDurationValid(view, startStr, endStr) {
  const parseDate = (dateStr) => {
    const [year, month, day] = dateStr.split("-").map(Number);
    return new Date(year, month - 1, day);
  };

  if (startStr === undefined || endStr === undefined) return false;

  const start = parseDate(startStr);
  const end = parseDate(endStr);

  if (isNaN(start) || isNaN(end)) return false;

  const diff = end.getTime() - start.getTime();
  if (diff < 0) return false;

  const days = diff / (1000 * 60 * 60 * 24);

  switch (view) {
    // The maximum number of weeks displayed in the calendar in a month is 6
    case "dayGridMonth":
      return days >= 28 && days <= 42;
    case "timeGridWeek":
    case "listWeek":
      return days === 7;
    case "timeGridDay":
      return days === 1;
    default:
      return false;
  }
}

export function getMonthFromCalendarStart(dateStr) {
  const date = new Date(dateStr);

  if (date.getDate() <= 7) {
    return formatToDateString(date);
  }

  return formatToDateString(
    new Date(date.getFullYear(), date.getMonth() + 1, 1)
  );
}

export function getDaySuffix(day) {
  if (day >= 11 && day <= 13) return `${day}th`; // Special case for 11th, 12th, 13th

  switch (day % 10) {
    case 1:
      return `${day}st`;
    case 2:
      return `${day}nd`;
    case 3:
      return `${day}rd`;
    default:
      return `${day}th`;
  }
}

export function timeStringFromDate(date, time_format = "24-hour") {
  const hour12 =
    time_format === "12-hour"
      ? true
      : time_format === "24-hour"
        ? false
        : undefined;

  return date.toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
    hour12: hour12,
  });
}

export function dayNameFromDate(date) {
  return date.toLocaleDateString([], { weekday: "long" });
}

export function combineDateTimeLocal(date, timeStr) {
  const year = date.getFullYear();
  const month = date.getMonth();
  const day = date.getDate();

  const [hours, minutes] = timeStr.split(":").map(Number);

  return new Date(year, month, day, hours, minutes);
}

export function addTimeToDate(date, hours = 0, minutes = 0) {
  const newDate = new Date(date);

  newDate.setHours(newDate.getHours() + hours);
  newDate.setMinutes(newDate.getMinutes() + minutes);

  return newDate;
}

export function formatDuration(duration) {
  const minutes = duration % 60;
  let hours = 0;

  if (duration >= 60) {
    hours = Math.floor(duration / 60);
  }

  return `${hours > 0 ? `${hours}h ` : ""}${minutes > 0 ? `${minutes}m` : ""}`;
}

export function GetDayPickerWindow(month, firstDayOfWeek = "Monday") {
  const weekStartsOn = firstDayOfWeek === "Monday" ? 1 : 0;

  const year = month.getFullYear();
  const m = month.getMonth();

  const monthStart = new Date(year, m, 1);
  const monthEnd = new Date(year, m + 1, 0);

  const startOffset = (monthStart.getDay() - weekStartsOn + 7) % 7;
  const endOffset = (weekStartsOn + 6 - monthEnd.getDay() + 7) % 7;

  const start = new Date(year, m, 1 - startOffset);
  const end = new Date(year, m + 1, endOffset);

  return {
    start: formatToDateString(start),
    end: formatToDateString(end),
  };
}

export function GenerateTimeOptions(time_format) {
  const options = [];

  for (let hour = 0; hour < 24; hour++) {
    for (let minute of [0, 30]) {
      const value = `${hour.toString().padStart(2, "0")}:${minute === 0 ? "00" : "30"}`;

      let label;
      if (time_format === "12-hour") {
        const period = hour >= 12 ? "PM" : "AM";
        const hour12 = hour % 12 || 12;
        label = `${hour12}:${minute === 0 ? "00" : "30"} ${period}`;
      } else {
        label = `${hour}:${minute === 0 ? "00" : "30"}`;
      }

      options.push({ label, value });
    }
  }
  return options;
}
