import { DayPicker } from "react-day-picker";

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
      disabled={disabled}
      modifiers={{
        disabled_selected: disabledSelectedModifier,
      }}
      modifiersClassNames={{
        disabled_selected:
          "rounded-md bg-primary/80 focus:bg-primary/80! hover:bg-primary/80! hover:text-white text-white",
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
        weekday:
          "w-9 font-normal text-[0.8rem] text-gray-600 dark:text-gray-400",
        week: "flex w-full mt-2 px-2",
        day: "h-9 w-9 inline-flex justify-center items-center text-sm rounded-md hover:bg-hvr_gray hover:text-text_color",
        selected:
          "rounded-md bg-primary focus:bg-primary hover:bg-primary hover:text-white text-white",
        today: !disabledTodayStyling ? "bg-hvr_gray" : "",
        outside: "text-white",
        disabled: "hover:bg-transparent text-gray-300! dark:text-gray-800!",
        hidden: "invisible",
        chevron:
          "w-5 h-5 m-1 fill-gray-500 dark:fill-gray-300 hover:fill-text_color",
      }}
    />
  );
}
