import {
  calculateStartEndTime,
  dateAndTimeStringsToLocalDate,
  dateStringToLocalDate,
  formatDuration,
  formatToDateString,
  getMonthFromCalendarStart,
  isDurationValid,
} from "@reservations/lib/datetime";
import { afterEach, describe, expect, it, vi } from "vitest";

afterEach(() => {
  vi.useRealTimers();
});

describe("dateStringToLocalDate", () => {
  it("parses a date as local calendar components", () => {
    const date = dateStringToLocalDate("2026-09-19");

    expect(date).not.toBeNull();
    expect(date.getFullYear()).toBe(2026);
    expect(date.getMonth()).toBe(8);
    expect(date.getDate()).toBe(19);
    expect(date.getHours()).toBe(0);
  });

  it.each(["2026-02-29", "2026-13-01", "2026-01-32", "19-09-2026", ""])(
    "rejects invalid date %s",
    (value) => {
      expect(dateStringToLocalDate(value)).toBeNull();
    }
  );
});

describe("dateAndTimeStringsToLocalDate", () => {
  it("combines local calendar date and time components", () => {
    const date = dateAndTimeStringsToLocalDate("2026-09-19", "14:30");

    expect(date).not.toBeNull();
    expect(date.getFullYear()).toBe(2026);
    expect(date.getMonth()).toBe(8);
    expect(date.getDate()).toBe(19);
    expect(date.getHours()).toBe(14);
    expect(date.getMinutes()).toBe(30);
  });

  it.each([
    ["2026-02-29", "14:30"],
    ["2026-09-19", "24:00"],
    ["2026-09-19", "14:60"],
    ["2026-09-19", "2:30"],
  ])("rejects invalid date/time values %s %s", (date, time) => {
    expect(dateAndTimeStringsToLocalDate(date, time)).toBeNull();
  });
});

describe("isDurationValid", () => {
  it("invalid dates", () => {
    expect(isDurationValid("week", "2025", "01-12")).toBe(false);
  });

  it("invalid view", () => {
    expect(isDurationValid("week", "2025-01-11", "2025-01-12")).toBe(false);
  });

  it.each([
    ["2025-02-30", "2025-03-03"],
    ["2025-13-01", "2026-01-02"],
    ["2025-1-01", "2025-01-02"],
  ])("rejects normalized or malformed dates %s", (start, end) => {
    expect(isDurationValid("timeGridDay", start, end)).toBe(false);
  });

  it("dayGridMonth a month", () => {
    expect(isDurationValid("dayGridMonth", "2025-01-01", "2025-02-01")).toBe(
      true
    );
  });

  it("dayGridMonth more than a month", () => {
    expect(isDurationValid("dayGridMonth", "2025-01-01", "2025-02-20")).toBe(
      false
    );
  });

  it("dayGridMonth less than a month", () => {
    expect(isDurationValid("dayGridMonth", "2025-01-18", "2025-02-01")).toBe(
      false
    );
  });

  it("timeGridWeek a week", () => {
    expect(isDurationValid("timeGridWeek", "2025-01-27", "2025-02-03")).toBe(
      true
    );
  });

  it("timeGridWeek more than a week", () => {
    expect(isDurationValid("timeGridWeek", "2025-01-27", "2025-02-04")).toBe(
      false
    );
  });

  it("timeGridWeek less than a week", () => {
    expect(isDurationValid("timeGridWeek", "2025-01-31", "2025-02-03")).toBe(
      false
    );
  });

  it("timeGridWeek during daylight savings switch", () => {
    expect(isDurationValid("timeGridWeek", "2026-03-23", "2026-03-30")).toBe(
      true
    );
  });

  it("timeGridDay a day", () => {
    expect(isDurationValid("timeGridDay", "2025-01-31", "2025-02-01")).toBe(
      true
    );
  });

  it("timeGridDay more than a day", () => {
    expect(isDurationValid("timeGridDay", "2025-01-31", "2025-02-02")).toBe(
      false
    );
  });
});

describe("calculateStartEndTime", () => {
  it("calculates a week containing a supplied date", () => {
    expect(
      calculateStartEndTime(
        "timeGridWeek",
        "Monday",
        new Date(2026, 8, 24)
      )
    ).toEqual({
      start: "2026-09-21",
      end: "2026-09-28",
    });
  });

  it("handles a month that starts on Sunday when weeks start on Monday", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2026, 1, 15));

    expect(calculateStartEndTime("dayGridMonth", "Monday")).toEqual({
      start: "2026-01-26",
      end: "2026-03-02",
    });
  });

  it("handles a month that ends on Sunday when weeks start on Monday", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2025, 7, 15));

    expect(calculateStartEndTime("dayGridMonth", "Monday")).toEqual({
      start: "2025-07-28",
      end: "2025-09-01",
    });
  });

  it("returns null for an unsupported view", () => {
    expect(calculateStartEndTime("unsupported", "Monday")).toBeNull();
  });
});

describe("formatDuration", () => {
  it("formats zero minutes", () => {
    expect(formatDuration(0)).toBe("0m");
  });

  it("does not leave trailing whitespace for whole hours", () => {
    expect(formatDuration(60)).toBe("1h");
  });
});

describe("getMonthFromCalendarStart", () => {
  // months indexing starts with 0 in js 0 ---> January
  it("date before start", () => {
    expect(getMonthFromCalendarStart("2025-01-27")).toBe(
      formatToDateString(new Date(2025, 1, 1))
    );
    expect(getMonthFromCalendarStart("2025-03-31")).toBe(
      formatToDateString(new Date(2025, 3, 1))
    );
    expect(getMonthFromCalendarStart("2025-05-26")).toBe(
      formatToDateString(new Date(2025, 5, 1))
    );
  });

  it("date at start", () => {
    expect(getMonthFromCalendarStart("2025-03-01")).toBe(
      formatToDateString(new Date(2025, 2, 1))
    );
    expect(getMonthFromCalendarStart("2025-09-01")).toBe(
      formatToDateString(new Date(2025, 8, 1))
    );
    expect(getMonthFromCalendarStart("2025-12-01")).toBe(
      formatToDateString(new Date(2025, 11, 1))
    );
  });
});
