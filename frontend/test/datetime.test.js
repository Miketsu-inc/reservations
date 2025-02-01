import { isDurationValid } from "@lib/datetime";
import { describe, expect, it } from "vitest";

describe("isDurationValid", () => {
  it("invalid dates", () => {
    expect(isDurationValid("week", "2025", "01-12")).toBe(false);
  });

  it("invalid view", () => {
    expect(isDurationValid("week", "2025-01-11", "2025-01-12")).toBe(false);
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
    expect(isDurationValid("timeGridWeek", "2025-01-27", "2025-02-02")).toBe(
      true
    );
  });

  it("timeGridWeek more than a week", () => {
    expect(isDurationValid("timeGridWeek", "2025-01-27", "2025-02-03")).toBe(
      false
    );
  });

  it("timeGridWeek less than a week", () => {
    expect(isDurationValid("timeGridWeek", "2025-01-31", "2025-02-02")).toBe(
      false
    );
  });

  it("timeGridDay a day", () => {
    expect(isDurationValid("timeGridDay", "2025-01-31", "2025-01-31")).toBe(
      true
    );
  });

  it("timeGridDay more than a day", () => {
    expect(isDurationValid("timeGridDay", "2025-01-31", "2025-02-01")).toBe(
      false
    );
  });
});
