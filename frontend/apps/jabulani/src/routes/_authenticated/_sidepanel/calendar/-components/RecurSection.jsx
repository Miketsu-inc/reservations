import { RefreshIcon } from "@reservations/assets";
import { DatePicker, Input, Select, Switch } from "@reservations/components";
import {
  dayNameFromDate,
  getDaySuffix,
  timeStringFromDate,
} from "@reservations/lib";

function recurUntiText(startDate, endDate) {
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
      console.error("frequency does not match any of the expected values");
  }
}

export default function RecurSection({
  booking,
  recurData,
  updateRecurData,
  disabled,
  onSelectOpenChange,
  onDatePickerOpenChange,
}) {
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

  return (
    <div>
      <div
        className="border-text_color flex flex-row items-center justify-between"
      >
        <div className="flex flex-row items-center gap-2">
          <RefreshIcon styles="size-5" />
          <p>Recurring booking</p>
        </div>
        <Switch
          size="large"
          defaultValue={recurData.isRecurring}
          onSwitch={() =>
            updateRecurData({ isRecurring: !recurData.isRecurring })
          }
          disabled={disabled}
        />
      </div>
      {/* TODO: this should have overflow-hidden on it to hide the content while transitioning
                but it causes the dropdowns to not open. This could be solved by only applying
                overflow-hidden while transitioning or reworking the dropdowns */}
      <div
        className={`${
          recurData.isRecurring
            ? `${recurData.frequency !== "custom" ? "max-h-60" : "max-h-96"} p-2
              opacity-100`
            : "max-h-0 overflow-hidden p-0 opacity-0"
          } flex flex-col gap-3 transition-all duration-300 sm:w-86`}
      >
        <p className="max-w-5/6 text-sm">
          {`Repeats ${recurFreqText(booking.start, recurData.frequency)} at ${timeStringFromDate(booking.start)} - ${timeStringFromDate(booking.end)}
                      from ${recurUntiText(booking.start, recurData.endDate)}`}
        </p>
        <Select
          labelText="Frequency"
          required={false}
          styles="w-full"
          options={[
            { value: "daily", label: "Daily" },
            { value: "weekly", label: "Weekly" },
            { value: "monthly", label: "Monthly" },
            { value: "custom", label: "Custom" },
          ]}
          value={recurData.frequency}
          onSelect={(option) => updateRecurData({ frequency: option.value })}
          onOpenChange={(open) => onSelectOpenChange(open)}
        />
        {recurData.frequency === "custom" && (
          <div>
            <p className="pb-1 text-sm">Repeat days</p>
            <div className="flex w-full flex-row items-center gap-1 pb-3">
              {days.map(({ label, value }) => {
                const isSelected = recurData.days.includes(value);

                return (
                  <button
                    key={value}
                    onClick={() => toggleDay(value)}
                    className={`${
                      isSelected
                        ? "bg-primary"
                        : "hover:bg-primary bg-gray-300 dark:bg-gray-800"
                    } flex-1 cursor-pointer rounded-full p-2 text-center text-sm
                    transition-colors duration-300`}
                  >
                    {label}
                  </button>
                );
              })}
            </div>
            <Input
              styles="p-2"
              labelText="Interval"
              childrenSide="right"
              required={false}
              type="number"
              inputData={(data) => updateRecurData({ interval: data.value })}
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
                onOpenChange={(open) => onSelectOpenChange(open)}
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
          onOpenChange={(open) => onDatePickerOpenChange(open)}
        />
      </div>
    </div>
  );
}
