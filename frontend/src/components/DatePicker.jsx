import DatePickerIcon from "@icons/DatePickerIcon";
import { useClickOutside } from "@lib/hooks";
import { useRef, useState } from "react";
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
  onSelect,
}) {
  const datePickerRef = useRef();
  const [showCalendar, setShowCalendar] = useState(false);
  const [selectedDate, SetSelectedDate] = useState(defaultDate);
  useClickOutside(datePickerRef, () => {
    setShowCalendar(false);
    if (clearAfterClose) SetSelectedDate();
  });

  return (
    <>
      <div ref={datePickerRef} className={`${styles}`}>
        <button
          className={`${!disabled ? "focus:border-text_color dark:focus:border-white" : ""} w-full
            rounded-md border border-gray-400 px-3 py-2 text-left text-gray-900
            focus:outline-none dark:border-gray-500 dark:bg-neutral-950`}
          type="button"
          onClick={() => {
            setShowCalendar(!showCalendar);
            if (clearAfterClose) SetSelectedDate();
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
            <DatePickerIcon styles="stroke-text_color shrink-0 h-4 w-4" />
          </div>
        </button>
        {showCalendar && !disabled && (
          <div className="relative top-1.5">
            <div
              className="absolute z-50 w-fit rounded-md border border-gray-300 bg-white shadow-lg
                dark:border-gray-500 dark:bg-neutral-950"
            >
              <SmallCalendar
                value={selectedDate}
                onSelect={(date) => {
                  if (!date) {
                    if (preventUnselect) return;
                    if (resetOnUnselect) date = defaultDate;
                  }

                  SetSelectedDate(date);
                  onSelect(date);
                }}
                firstDayOfWeek={firstDayOfWeek}
                disabled={{ before: disabledBefore }}
                disabledSelectedModifier={
                  disabledBefore?.getTime() > selectedDate?.getTime()
                    ? selectedDate
                    : undefined
                }
              />
            </div>
          </div>
        )}
      </div>
    </>
  );
}
