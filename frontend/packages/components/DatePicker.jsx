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
  styles = "",
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
            <span
              className="text-fg flex items-center gap-1 pb-1.5 text-sm
                font-medium"
            >
              {labelText}
              {required !== false && (
                <span
                  aria-hidden="true"
                  className="text-fg-danger text-base leading-none"
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
                  bg-bg-input border-border-strong
                  hover:not-disabled:not-focus-visible:not-data-popup-open:border-border-hover
                  focus-visible:border-primary focus-visible:ring-primary-focus
                  w-full rounded-lg border px-3 py-2 text-left outline-hidden
                  focus-visible:ring-2 md:text-sm`}
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
                    <span className="text-fg h-5 flex-1 truncate">
                      {selectedDate
                        ? formatDate(selectedDate)
                        : placeholderText || "Pick a date"}
                    </span>
                  )}
                  {(!clearable || !selectedDate) && (
                    <Icon
                      icon={Calendar04Icon}
                      styles="text-fg-muted shrink-0 size-5"
                    />
                  )}
                </div>
              </button>
            </PopoverTrigger>
            {clearable && selectedDate && !preventUnselect && (
              <button
                aria-label="Clear selected date"
                className="hover:bg-hover text-fg-muted absolute top-1/2
                  right-1.5 flex size-7 -translate-y-1/2 cursor-pointer
                  items-center justify-center rounded-md"
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
