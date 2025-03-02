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
