import {
  calculateStartEndTime,
  dateAndTimeStringsToLocalDate,
  dateStringToLocalDate,
  formatDuration,
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
    expect(isDurationValid("month", "2025-01-11", "2025-02-11")).toBe(false);
  });

  it.each([
    ["2025-02-30", "2025-03-03"],
    ["2025-13-01", "2026-01-02"],
    ["2025-1-01", "2025-01-02"],
  ])("rejects normalized or malformed dates %s", (start, end) => {
    expect(isDurationValid("day", start, end)).toBe(false);
  });

  it("week a week", () => {
    expect(isDurationValid("week", "2025-01-27", "2025-02-03")).toBe(true);
  });

  it("week more than a week", () => {
    expect(isDurationValid("week", "2025-01-27", "2025-02-04")).toBe(false);
  });

  it("week less than a week", () => {
    expect(isDurationValid("week", "2025-01-31", "2025-02-03")).toBe(false);
  });

  it("week during daylight savings switch", () => {
    expect(isDurationValid("week", "2026-03-23", "2026-03-30")).toBe(true);
  });

  it("day a day", () => {
    expect(isDurationValid("day", "2025-01-31", "2025-02-01")).toBe(true);
  });

  it("day more than a day", () => {
    expect(isDurationValid("day", "2025-01-31", "2025-02-02")).toBe(false);
  });

  it("2day two days", () => {
    expect(isDurationValid("2day", "2025-01-31", "2025-02-02")).toBe(true);
  });

  it("2day a single day", () => {
    expect(isDurationValid("2day", "2025-01-31", "2025-02-01")).toBe(false);
  });
});

describe("calculateStartEndTime", () => {
  it("calculates a week containing a supplied date", () => {
    expect(
      calculateStartEndTime("week", "Monday", new Date(2026, 8, 24))
    ).toEqual({
      start: "2026-09-21",
      end: "2026-09-28",
    });
  });

  it("starts two day views on the supplied date", () => {
    expect(
      calculateStartEndTime("2day", "Monday", new Date(2026, 9, 31))
    ).toEqual({
      start: "2026-10-31",
      end: "2026-11-02",
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
