import { DayPicker } from "@daypicker/react";

function getRangeModifiers({ from, to }) {
  if (from.toDateString() === to.toDateString()) return { range_single: from };

  return {
    range_start: from,
    range_middle: { after: from, before: to },
    range_end: to,
  };
}

export default function SmallCalendar({
  value,
  disabled,
  animate,
  onSelect,
  firstDayOfWeek,
  disabledSelectedModifier,
  disabledTodayStyling = false,
  onMonthChange,
  startMonth,
  endMonth,
  month,
  highlightedRange,
}) {
  return (
    <DayPicker
      mode="single"
      showOutsideDays={true}
      animate={animate}
      weekStartsOn={
        firstDayOfWeek === "Monday"
          ? 1
          : firstDayOfWeek === "Sunday"
            ? 0
            : undefined
      }
      selected={highlightedRange ? undefined : value}
      defaultMonth={
        month === undefined ? (highlightedRange?.from ?? value) : undefined
      }
      disabled={disabled}
      modifiers={{
        disabled_selected: disabledSelectedModifier,
        ...(highlightedRange && getRangeModifiers(highlightedRange)),
      }}
      modifiersClassNames={{
        disabled_selected:
          "rounded-md bg-primary-hover focus:bg-primary-hover! hover:bg-primary-hover! hover:text-fg-on-primary text-fg-on-primary",
        // Important, so they win over the today and hover styles.
        range_single:
          "rounded-md bg-primary-subtle! text-fg-primary! font-semibold",
        range_start:
          "rounded-r-none! bg-primary-subtle! text-fg-primary! font-semibold",
        range_middle:
          "rounded-none! bg-primary-subtle! text-fg-primary! font-semibold",
        range_end:
          "rounded-l-none! bg-primary-subtle! text-fg-primary! font-semibold",
      }}
      onSelect={(date) => onSelect(date)}
      onMonthChange={onMonthChange}
      month={month}
      startMonth={startMonth}
      endMonth={endMonth}
      classNames={{
        month: "space-y-4 pb-2",
        month_caption: "flex justify-center items-center w-full pt-3",
        caption_label: "font-medium",
        nav: "absolute flex items-center justify-between w-full p-2",
        weekdays: "flex px-2",
        weekday: "w-9 text-xs text-fg-muted",
        week: "flex w-full mt-2 px-2",
        day: "size-9 inline-flex justify-center items-center text-sm rounded-md hover:bg-hover hover:text-fg",
        day_button: "size-full",
        selected:
          "rounded-md bg-primary focus:bg-primary hover:bg-primary hover:text-fg-on-primary text-fg-on-primary",
        // The range band uses the usual today style, so today gets a dot.
        today: highlightedRange
          ? "relative text-fg-primary font-semibold after:absolute after:bottom-1 after:size-1 after:rounded-full after:bg-primary"
          : !disabledTodayStyling
            ? "bg-primary-subtle text-fg-primary font-semibold"
            : "",
        outside: "text-fg-muted",
        disabled: "hover:bg-transparent text-fg-disabled!",
        hidden: "invisible",
        chevron: "w-5 h-5 m-1 fill-fg-muted hover:fill-fg",
      }}
    />
  );
}
