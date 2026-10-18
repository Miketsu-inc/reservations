import { useCallback } from "react";

function getPhaseBackground(phases, start, end) {
  if (end <= start) return null;
  const stops = [];
  let offset = 0;

  for (const phase of phases) {
    const from = Math.max(start, offset);
    offset += phase.duration;
    const to = Math.min(end, offset);
    if (to <= from) continue;
    const color =
      phase.phase_type === "wait"
        ? "color-mix(in srgb, var(--fc-event-color) 30%, transparent)"
        : "var(--fc-event-color)";
    stops.push(
      `${color} ${((from - start) / (end - start)) * 100}% ${((to - start) / (end - start)) * 100}%`
    );
  }

  return stops.length
    ? `linear-gradient(to bottom, ${stops.join(", ")})`
    : null;
}

export default function CalendarEventContent({ eventInfo }) {
  const { event, timeText, view } = eventInfo;
  const viewType = view.type;
  const isListView = viewType.startsWith("list");
  const isMonthView = viewType === "dayGridMonth";
  const setBackground = useCallback(
    (label) => {
      const phases = event.extendedProps.phases;
      if (
        !label ||
        event.allDay ||
        !viewType.startsWith("timeGrid") ||
        !phases?.some((phase) => phase.phase_type === "wait")
      )
        return;

      const element = label.closest(".calendar-event");
      const date = element?.closest("[data-date]")?.dataset.date;
      if (!date) return;

      // Clip overnight bookings to the portion visible in this day's column.
      const dayStart = new Date(`${date}T00:00:00`);
      const dayEnd = new Date(dayStart);
      dayEnd.setDate(dayEnd.getDate() + 1);
      element.style.setProperty(
        "background",
        getPhaseBackground(
          phases,
          Math.max(0, +dayStart - +event.start) / 60_000,
          Math.min(+event.end - +event.start, +dayEnd - +event.start) / 60_000
        ) ?? "",
        "important"
      );
      return () => element.style.removeProperty("background");
    },
    [event, viewType]
  );

  if (isListView) {
    return (
      <div className="flex size-full min-w-0 items-center gap-4">
        {timeText && (
          <span className="text-fg-muted w-28 shrink-0 text-sm">
            {timeText}
          </span>
        )}
        <span className="min-w-0 truncate font-medium">{event.title}</span>
      </div>
    );
  }

  return (
    <div
      ref={setBackground}
      className={`flex size-full min-w-0 overflow-hidden
        ${isMonthView ? "flex-row items-center gap-1" : "flex-col"}`}
    >
      {timeText && (
        <span className="shrink-0 text-xs leading-tight font-medium">
          {timeText}
        </span>
      )}
      <span className="min-w-0 truncate text-xs leading-tight font-semibold">
        {event.title}
      </span>
    </div>
  );
}
