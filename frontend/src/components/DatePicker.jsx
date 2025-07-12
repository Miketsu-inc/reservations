import DatePickerIcon from "@icons/DatePickerIcon";
import { useState } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "./Popover";
import SmallCalendar from "./SmallCalendar";

function formatDate(date) {
  return date.toLocaleDateString([], {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export default function DatePicker({
  styles,
  defaultDate,
  palaceHolderText,
  disabledBefore,
  disabled = false,
  hideText = false,
  clearAfterClose = false,
  firstDayOfWeek = "Monday",
  preventUnselect = false,
  resetOnUnselect = true,
  closeOnSelect = false,
  onOpenChange,
  onSelect,
}) {
  const [showCalendar, setShowCalendar] = useState(false);
  const [selectedDate, setSelectedDate] = useState(defaultDate);

  return (
    <>
      <Popover
        open={showCalendar}
        onOpenChange={(open) => {
          open ? setShowCalendar(true) : setShowCalendar(false);
          onOpenChange?.(open);
        }}
      >
        <PopoverTrigger disabled={disabled} asChild>
          <button
            className={`${styles} ${disabled ? "outline-none" : ""} border-input_border_color w-full
              rounded-lg border px-3 py-2 text-left`}
            type="button"
            onClick={() => {
              setShowCalendar(!showCalendar);
              if (clearAfterClose) setSelectedDate();
            }}
          >
            <div className="flex items-center justify-between">
              {!hideText && (
                <span className="text-text_color h-5 flex-1 truncate">
                  {selectedDate
                    ? formatDate(selectedDate)
                    : palaceHolderText || "Pick a date"}
                </span>
              )}
              <DatePickerIcon styles="stroke-gray-700 dark:stroke-gray-300 shrink-0 h-4 w-4" />
            </div>
          </button>
        </PopoverTrigger>
        <PopoverContent styles="w-fit !p-0">
          <SmallCalendar
            value={selectedDate}
            onSelect={(date) => {
              if (!date) {
                if (preventUnselect) return;
                if (resetOnUnselect) date = defaultDate;
              }

              setSelectedDate(date);
              onSelect(date);

              if (closeOnSelect) {
                setShowCalendar(!showCalendar);
              }
            }}
            firstDayOfWeek={firstDayOfWeek}
            disabled={{ before: disabledBefore }}
            disabledSelectedModifier={
              disabledBefore?.getTime() > selectedDate?.getTime()
                ? selectedDate
                : undefined
            }
          />
        </PopoverContent>
      </Popover>
    </>
  );
}
