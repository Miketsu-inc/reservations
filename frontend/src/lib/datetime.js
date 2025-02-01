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

export function calculateStartEndTime(view) {
  const now = new Date();
  let start, end;

  switch (view) {
    case "dayGridMonth":
      // This will need a correction if the user's calendar
      // starts with Sunday instead of Monday
      start = new Date(now.getFullYear(), now.getMonth(), 1);
      start = new Date(
        start.getFullYear(),
        start.getMonth(),
        1 - (start.getDay() - 1)
      );
      end = new Date(start.getFullYear(), start.getMonth() + 1, 0);
      end = new Date(
        end.getFullYear(),
        end.getMonth(),
        end.getDate() + (7 - end.getDay())
      );
      break;
    case "timeGridWeek":
    case "listWeek":
      // This will need a correction if the user's calendar
      // starts with Sunday instead of Monday
      start = new Date(
        now.getFullYear(),
        now.getMonth(),
        now.getDate() - now.getDay() + 1
      );
      end = new Date(
        start.getFullYear(),
        start.getMonth(),
        start.getDate() + 6
      );
      break;
    case "timeGridDay":
      start = now;
      end = now;
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

  switch (view) {
    // A week tolerance due to how the calendar displays dates
    // and this function can not known it ahead of time
    case "dayGridMonth": {
      const expectedPlusWeek = new Date(
        start.getFullYear(),
        start.getMonth() + 1,
        start.getDate() + 7
      );

      const expectedMinusWeek = new Date(
        start.getFullYear(),
        start.getMonth() + 1,
        start.getDate() - 7
      );

      return (
        end.getTime() >= expectedMinusWeek && end.getTime() <= expectedPlusWeek
      );
    }
    case "timeGridWeek":
    case "listWeek":
      return (
        end.getTime() ===
        new Date(
          start.getFullYear(),
          start.getMonth(),
          start.getDate() + 6
        ).getTime()
      );
    case "timeGridDay":
      return (
        end.getTime() ===
        new Date(
          start.getFullYear(),
          start.getMonth(),
          start.getDate()
        ).getTime()
      );

    default:
      return false;
  }
}
