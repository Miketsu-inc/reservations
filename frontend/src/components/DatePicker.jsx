import DatePickerIcon from "@icons/DatePickerIcon";
import { useClickOutside } from "@lib/hooks";
import { useRef, useState } from "react";
import { DayPicker } from "react-day-picker";

function formatDate(date) {
  const strs = date.toDateString().split(" ");
  strs.splice(0, 1);
  strs[1] = strs[1] + ",";
  return strs.join(" ");
}

export default function DatePicker({
  styles,
  defaultDate,
  palaceHolderText,
  disabledBefore,
  hideText = false,
  clearAfterClose = false,
  firstDayOfWeek = "Monday",
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
          className="focus:border-text_color w-full rounded-md border border-gray-400 px-3 py-2
            text-left text-gray-900 focus:outline-none dark:border-gray-500
            dark:bg-neutral-950 dark:focus:border-white"
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
        {showCalendar && (
          <div
            className="absolute z-10 w-fit rounded-md border border-gray-300 bg-white shadow-lg
              dark:border-gray-500 dark:bg-neutral-950"
          >
            <DayPicker
              mode="single"
              showOutsideDays={true}
              weekStartsOn={
                firstDayOfWeek === "Monday"
                  ? 1
                  : firstDayOfWeek === "Sunday"
                    ? 0
                    : undefined
              }
              selected={selectedDate}
              disabled={{ before: disabledBefore }}
              onSelect={(date) => {
                SetSelectedDate(date);
                onSelect(date);
              }}
              classNames={{
                month: "space-y-4 pb-2",
                month_caption: "flex justify-center items-center w-full pt-3",
                caption_label: "font-medium",
                nav: "absolute flex items-center justify-between w-full p-2",
                button_previous: "rounded-md hover:bg-hvr_gray cursor-pointer",
                button_next: "rounded-md hover:bg-hvr_gray cursor-pointer",
                weekdays: "flex px-2",
                weekday:
                  "w-9 font-normal text-[0.8rem] text-gray-600 dark:text-gray-400",
                week: "flex w-full mt-2 px-2",
                day: "h-9 w-9 inline-flex justify-center items-center text-sm rounded-md hover:bg-hvr_gray hover:text-text_color",
                selected:
                  "rounded-md bg-primary focus:bg-primary hover:bg-primary hover:text-white text-white",
                today: "bg-hvr_gray",
                outside: "text-gray-500",
                disabled:
                  "hover:bg-transparent text-gray-300! dark:text-gray-800!",
                hidden: "invisible",
                chevron:
                  "w-5 h-5 m-1 fill-gray-500 dark:fill-gray-300 hover:fill-text_color",
              }}
            />
          </div>
        )}
      </div>
    </>
  );
}
