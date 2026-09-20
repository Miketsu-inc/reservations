import { fillStatisticsWithDate } from "@reservations/lib/lib";
import { describe, expect, it } from "vitest";

describe("fillStatisticsWithDate", () => {
  it("fills missing calendar dates without shifting ISO date-only values", () => {
    const result = fillStatisticsWithDate(
      [
        { day: "2026-03-28", value: 3 },
        { day: "2026-03-30T12:00:00Z", value: 5 },
      ],
      "2026-03-28",
      "2026-03-30"
    );

    expect(result.map(({ value }) => value)).toEqual([3, 0, 5]);
  });

  it("returns an empty result for invalid or reversed ranges", () => {
    expect(fillStatisticsWithDate([], "invalid", "2026-01-01")).toEqual([]);
    expect(fillStatisticsWithDate([], "2026-01-02", "2026-01-01")).toEqual([]);
  });
});
