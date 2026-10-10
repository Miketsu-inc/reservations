import { Calendar04Icon, Cancel01Icon } from "@hugeicons/core-free-icons";
import { useState } from "react";
import Icon from "./Icon.jsx";
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
  value,
  placeholderText,
  disabledBefore,
  disabledAfter,
  labelText,
  required,
  disabled = false,
  hideText = false,
  clearable = false,
  clearAfterClose = false,
  firstDayOfWeek = "Monday",
  preventUnselect = false,
  resetOnUnselect = true,
  closeOnSelect = false,
  onOpenChange,
  onSelect,
}) {
  const [showCalendar, setShowCalendar] = useState(false);
  const [internalDate, setInternalDate] = useState();
  const isControlled = value !== undefined;
  const selectedDate = isControlled ? value : internalDate;

  function handleOpenChange(open) {
    setShowCalendar(open);
    if (!open && clearAfterClose && !isControlled) setInternalDate(undefined);
    onOpenChange?.(open);
  }

  function clearSelectedDate(event) {
    event.preventDefault();
    event.stopPropagation();

    setInternalDate(undefined);
    onSelect?.(undefined);
  }

  return (
    <>
      <Popover open={showCalendar} onOpenChange={handleOpenChange}>
        <div className="w-full">
          {labelText && (
            <span className="flex items-center gap-1 pb-1 text-sm">
              {labelText}
              {required !== false && (
                <span
                  aria-hidden="true"
                  className="text-base leading-none text-red-500"
                >
                  *
                </span>
              )}
            </span>
          )}
          <div className="relative">
            <PopoverTrigger disabled={disabled} asChild>
              <button
                className={`${styles} ${disabled ? "outline-none" : ""}
                  border-input_border_color w-full rounded-lg border px-3 py-2
                  text-left`}
                type="button"
                disabled={disabled}
                aria-label={
                  hideText
                    ? selectedDate
                      ? `Selected date: ${formatDate(selectedDate)}`
                      : placeholderText || "Pick a date"
                    : undefined
                }
              >
                <div className="flex items-center justify-between">
                  {!hideText && (
                    <span className="text-text_color h-5 flex-1 truncate">
                      {selectedDate
                        ? formatDate(selectedDate)
                        : placeholderText || "Pick a date"}
                    </span>
                  )}
                  {(!clearable || !selectedDate) && (
                    <Icon
                      icon={Calendar04Icon}
                      styles="text-gray-700 dark:text-gray-300 shrink-0 size-5"
                    />
                  )}
                </div>
              </button>
            </PopoverTrigger>
            {clearable && selectedDate && !preventUnselect && (
              <button
                aria-label="Clear selected date"
                className="hover:bg-hvr_gray absolute top-1/2 right-1.5 flex
                  size-7 -translate-y-1/2 cursor-pointer items-center
                  justify-center rounded-md text-gray-700 dark:text-gray-300"
                disabled={disabled}
                type="button"
                onClick={clearSelectedDate}
              >
                <Icon icon={Cancel01Icon} styles="size-4" />
              </button>
            )}
          </div>
        </div>
        <PopoverContent styles="w-fit p-0!">
          <SmallCalendar
            value={selectedDate}
            onSelect={(date) => {
              if (!date) {
                if (preventUnselect) return;
                if (resetOnUnselect) date = value;
              }

              if (!isControlled) {
                setInternalDate(date);
              }
              onSelect?.(date);

              if (closeOnSelect) {
                handleOpenChange(false);
              }
            }}
            firstDayOfWeek={firstDayOfWeek}
            disabled={[{ before: disabledBefore }, { after: disabledAfter }]}
            disabledSelectedModifier={
              disabledBefore?.getTime() > value?.getTime() ? value : undefined
            }
          />
        </PopoverContent>
      </Popover>
    </>
  );
}
