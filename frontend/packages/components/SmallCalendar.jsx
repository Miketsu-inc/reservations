import { DayPicker } from "@daypicker/react";

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
      selected={value}
      defaultMonth={month === undefined ? value : undefined}
      disabled={disabled}
      modifiers={{
        disabled_selected: disabledSelectedModifier,
      }}
      modifiersClassNames={{
        disabled_selected:
          "rounded-md bg-primary-hover focus:bg-primary-hover! hover:bg-primary-hover! hover:text-fg-on-primary text-fg-on-primary",
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
        today: !disabledTodayStyling
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
