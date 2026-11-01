import { ArrowDown01Icon, ArrowLeft01Icon } from "@hugeicons/core-free-icons";
import {
  Button,
  Icon,
  Popover,
  PopoverContent,
  PopoverTrigger,
  SmallCalendar,
  Toggle,
  ToggleGroup,
} from "@reservations/components";
import { useState } from "react";

// Intl joins ranges with an en dash
const RANGE_DASH = /\s*[–—]\s*/g;

export function formatCalendarTitle(start, last, showDays) {
  let options;
  if (showDays) {
    const year = new Date().getFullYear();
    const isThisYear =
      start.getFullYear() === year && last.getFullYear() === year;
    options = {
      month: "short",
      day: "numeric",
      year: isThisYear ? undefined : "numeric",
    };
  } else {
    const isSameMonth =
      start.getMonth() === last.getMonth() &&
      start.getFullYear() === last.getFullYear();
    options = { month: isSameMonth ? "long" : "short", year: "numeric" };
  }

  return new Intl.DateTimeFormat(undefined, options)
    .formatRange(start, last)
    .replace(RANGE_DASH, " - ");
}

function CalendarViewSwitcher({ ranges, value, onChange }) {
  return (
    <ToggleGroup variant="segmented" value={value} onValueChange={onChange}>
      {ranges.map((range) => (
        <Toggle
          key={range.value}
          value={range.value}
          aria-label={range.label}
          title={range.label}
        >
          <Icon icon={range.icon} styles="size-5" />
        </Toggle>
      ))}
    </ToggleGroup>
  );
}

function CalendarTitle({ visibleDates, firstDayOfWeek, onSelect }) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          aria-label={`${visibleDates?.title}, pick a date`}
          className="hover:bg-hover data-popup-open:bg-hover -mx-2 flex
            cursor-pointer items-center gap-1.5 rounded-lg px-2 py-1 text-xl
            font-semibold whitespace-nowrap md:text-2xl"
        >
          {visibleDates?.title}
          <Icon
            icon={ArrowDown01Icon}
            styles={`text-fg-muted size-5 shrink-0 transition-transform
              ${isOpen ? "rotate-180" : ""}`}
          />
        </button>
      </PopoverTrigger>
      <PopoverContent
        styles="w-fit p-0!"
        align="start"
        aria-label="Pick a date"
      >
        <SmallCalendar
          highlightedRange={visibleDates?.range}
          firstDayOfWeek={firstDayOfWeek}
          onSelect={(date) => {
            onSelect(date);
            setIsOpen(false);
          }}
        />
      </PopoverContent>
    </Popover>
  );
}

export default function CalendarToolbar({
  visibleDates,
  firstDayOfWeek,
  onSelectDate,
  onPrev,
  onToday,
  onNext,
  ranges,
  range,
  onRangeChange,
  children,
}) {
  return (
    <div className="flex shrink-0 flex-wrap items-center gap-x-4 gap-y-2 pb-4">
      <div className="flex flex-row items-center gap-1">
        <button
          className="text-fg-muted hover:bg-hover hover:text-fg cursor-pointer
            rounded-lg p-2"
          type="button"
          aria-label={visibleDates?.buttons.prev.hint}
          onClick={onPrev}
        >
          <Icon icon={ArrowLeft01Icon} styles="size-5" />
        </button>
        <Button
          variant="secondary"
          styles="px-3 py-2 text-sm"
          buttonText="Today"
          onClick={onToday}
        />
        <button
          className="text-fg-muted hover:bg-hover hover:text-fg cursor-pointer
            rounded-lg p-2"
          type="button"
          aria-label={visibleDates?.buttons.next.hint}
          onClick={onNext}
        >
          <Icon icon={ArrowLeft01Icon} styles="size-5 rotate-180" />
        </button>
      </div>
      <div className="order-first w-full md:order-0 md:w-auto">
        <CalendarTitle
          visibleDates={visibleDates}
          firstDayOfWeek={firstDayOfWeek}
          onSelect={onSelectDate}
        />
      </div>
      <div
        className="ml-auto flex min-w-0 flex-1 flex-row items-center justify-end
          gap-2 md:flex-none"
      >
        {ranges.length > 1 && (
          <CalendarViewSwitcher
            ranges={ranges}
            value={range}
            onChange={onRangeChange}
          />
        )}
        {children}
      </div>
    </div>
  );
}
