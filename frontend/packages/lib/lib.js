import { JABULANI_URL, TANGO_URL } from "./constants";
import {
  dateStringToLocalDate,
  formatToDateString,
  isoToDateString,
} from "./datetime";

export function invalidateLocalStorageAuth(responseCode) {
  if (responseCode === 401) {
    localStorage.setItem("loggedIn", false);
  }
}

export function fillStatisticsWithDate(data, fromDateStr, toDateStr) {
  const filledStats = [];

  const fromDate = dateStringToLocalDate(isoToDateString(fromDateStr));
  const toDate = dateStringToLocalDate(isoToDateString(toDateStr));
  if (!fromDate || !toDate || fromDate > toDate) return filledStats;

  const convertedMap = new Map(
    data.map(({ day, value }) => [isoToDateString(day), value])
  );

  for (
    const date = new Date(fromDate);
    date <= toDate;
    date.setDate(date.getDate() + 1)
  ) {
    const value = convertedMap.get(formatToDateString(date)) ?? 0;

    filledStats.push({
      value,
      day: date.toLocaleDateString([], {
        month: "2-digit",
        day: "2-digit",
      }),
    });
  }

  return filledStats;
}

export function getDisplayPrice(price, type) {
  if (type === "free") {
    return "Free";
  }

  if (type === "from") {
    return `from ${price}`;
  }

  return price;
}

export function getSafeRedirectUrl(value) {
  if (!value) return null;

  try {
    const target = new URL(value, window.location.origin);

    const allowedOrigins = new Set([
      new URL(JABULANI_URL).origin,
      new URL(TANGO_URL).origin,
    ]);

    return allowedOrigins.has(target.origin) ? target.href : null;
  } catch {
    return null;
  }
}
