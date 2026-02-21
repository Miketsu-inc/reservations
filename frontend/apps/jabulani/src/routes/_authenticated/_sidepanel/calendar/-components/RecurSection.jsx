import { Button, DatePicker, Input, Select } from "@reservations/components";
import {
  dayNameFromDate,
  getDaySuffix,
  timeStringFromDate,
} from "@reservations/lib";
import { useState } from "react";

function recurUntiText(startDate, endDate) {
  if (!startDate || !endDate) return "";

  const showYear =
    startDate.getFullYear() === endDate.getFullYear() ? false : true;

  const startDateStr = startDate.toLocaleDateString([], {
    year: showYear ? "numeric" : undefined,
    month: "short",
    day: "numeric",
  });

  const endDateStr = endDate.toLocaleDateString([], {
    year: showYear ? "numeric" : undefined,
    month: "short",
    day: "numeric",
  });

  return `${startDateStr} to ${endDateStr}`;
}

function recurFreqText(startDate, freq) {
  switch (freq) {
    case "daily":
      return "daily";
    case "weekly":
      return `every ${dayNameFromDate(startDate)}`;
    case "monthly":
      return `on the ${getDaySuffix(startDate.getDate())} of each month`;
    case "custom":
      return "";
    default:
      return "does not repeat";
  }
}

const currentDate = new Date();
const defaultRecurData = {
  isRecurring: false,
  frequency: "weekly",
  endDate: new Date(currentDate.setMonth(currentDate.getMonth() + 1)),
  interval: 1,
  intervalUnit: "weeks",
  days: [],
};

export default function RecurSection({ booking, recurringData, onSave }) {
  const [recurData, setRecurData] = useState(recurringData || defaultRecurData);

  function updateRecurData(data) {
    setRecurData((prev) => ({ ...prev, ...data }));
  }

  const days = [
    { label: "Mo", value: "MO" },
    { label: "Tu", value: "TU" },
    { label: "We", value: "WE" },
    { label: "Th", value: "TH" },
    { label: "Fr", value: "FR" },
    { label: "Sa", value: "SA" },
    { label: "Su", value: "SU" },
  ];

  function toggleDay(value) {
    const newDays = recurData.days.includes(value)
      ? recurData.days.filter((d) => d !== value)
      : [...recurData.days, value];

    updateRecurData({ days: newDays });
  }

  function handleFreqSelect(option) {
    if (option.value === "not-repeat") {
      updateRecurData({ isRecurring: false });
    } else {
      updateRecurData({ isRecurring: true, frequency: option.value });
    }
  }

  return (
    <div className="relative h-full w-full">
      <div className="flex h-full flex-1 flex-col gap-6 px-6">
        <p className="text-2xl font-semibold">Recurring Rules</p>

        <div className="flex h-full flex-1 flex-col gap-5">
          <Select
            labelText="Frequency"
            required={false}
            styles="w-full"
            options={[
              { value: "not-repeat", label: "Doesn't repeat" },
              { value: "daily", label: "Daily" },
              { value: "weekly", label: "Weekly" },
              { value: "monthly", label: "Monthly" },
              { value: "custom", label: "Custom" },
            ]}
            value={
              !recurData?.isRecurring ? "not-repeat" : recurData?.frequency
            }
            onSelect={handleFreqSelect}
          />

          <div
            className={`flex flex-col overflow-hidden transition-all
              duration-300 ease-in-out ${
                recurData?.isRecurring
                  ? "max-h-125 translate-y-0 opacity-100"
                  : "max-h-0 -translate-y-2 opacity-0"
              }`}
          >
            <div className="flex flex-col gap-3">
              <p className="text-text_color/70 px-1">
                {`Repeats ${recurFreqText(
                  booking.start,
                  recurData?.frequency
                )} at ${timeStringFromDate(booking.start)} - ${timeStringFromDate(
                  booking.end
                )}
              from ${recurUntiText(booking?.start, recurData?.endDate)}`}
              </p>

              {recurData?.frequency === "custom" && (
                <div className="flex flex-col gap-3 pt-2">
                  <div>
                    <p className="pb-1 text-sm">Repeat days</p>
                    <div
                      className="flex w-full flex-row items-center gap-1
                        md:gap-2"
                    >
                      {days.map(({ label, value }) => {
                        const isSelected = recurData?.days?.includes(value);
                        return (
                          <button
                            key={value}
                            onClick={() => toggleDay(value)}
                            className={`${
                              isSelected
                                ? "border-primary bg-primary/10 text-primary"
                                : "border-border_color hover:border-gray-400"
                            } text-text_color/70 size-10 flex-1 cursor-pointer
                            rounded-full border-2 p-2 text-center text-sm
                            font-medium transition-all duration-200 md:size-12
                            md:text-base`}
                          >
                            {label}
                          </button>
                        );
                      })}
                    </div>
                  </div>
                  <Input
                    styles="p-2"
                    labelText="Interval"
                    childrenSide="right"
                    required={false}
                    type="number"
                    inputData={(data) =>
                      updateRecurData({ interval: data.value })
                    }
                    value={recurData.interval}
                  >
                    <Select
                      styles="rounded-l-none min-w-24!"
                      options={[
                        { value: "days", label: "days" },
                        { value: "weeks", label: "weeks" },
                      ]}
                      onSelect={(option) =>
                        updateRecurData({ intervalUnit: option.value })
                      }
                      value={recurData.intervalUnit}
                    />
                  </Input>
                </div>
              )}

              <DatePicker
                labelText="End date"
                required={false}
                styles="w-full"
                value={recurData.endDate}
                disabledBefore={new Date()}
                onSelect={(date) => updateRecurData({ endDate: date })}
              />
            </div>
          </div>
        </div>
        <div
          className="border-border_color bg-layer_bg items center fixed right-0
            bottom-0 left-0 flex w-full border-t px-6 py-4"
        >
          <Button
            styles="py-2 px-4 w-full"
            variant="primary"
            name="createButton"
            buttonText="Save"
            onClick={() => onSave(recurData)}
          />
        </div>
      </div>
    </div>
  );
}
