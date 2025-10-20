import { RefreshIcon } from "@reservations/assets";
import { DatePicker, Select, Switch } from "@reservations/components";
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
    default:
      console.error("frequency does not match any of the expected values");
  }
}

export default function RecurSection({
  booking,
  recurData,
  updateRecurData,
  disabled,
}) {
  return (
    <>
      <div
        className="border-text_color flex flex-row items-center justify-between
          px-1 pt-2"
      >
        <div className="flex flex-row items-center gap-2">
          <RefreshIcon styles="size-5" />
          <p>Recurring booking?</p>
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
            ? "max-h-52 p-2 opacity-100"
            : "max-h-0 overflow-hidden p-0 opacity-0"
          } flex flex-col gap-3 transition-all duration-300 sm:w-86`}
      >
        <p className="max-w-5/6 text-sm">
          {`Recurs ${recurFreqText(booking.start, recurData.frequency)} at ${timeStringFromDate(booking.start)} - ${timeStringFromDate(booking.end)}
                      from ${recurUntiText(booking.start, recurData.endDate)}`}
        </p>
        <Select
          styles="w-full"
          options={[
            { value: "daily", label: "Daily" },
            { value: "weekly", label: "Weekly" },
            { value: "monthly", label: "Monthly" },
          ]}
          value={recurData.frequency}
          onSelect={(option) => updateRecurData({ frequency: option.value })}
        />
        <p className="text-sm">Recur until</p>
        <DatePicker
          styles="w-full"
          defaultDate={
            new Date(
              booking.start.getFullYear(),
              booking.start.getMonth() + 1,
              booking.start.getDate()
            )
          }
          disabledBefore={new Date()}
          onSelect={(date) => updateRecurData({ endDate: date })}
        />
      </div>
    </>
  );
}
