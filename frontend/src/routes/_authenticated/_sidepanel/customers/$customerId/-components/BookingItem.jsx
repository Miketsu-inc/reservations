import CalendarIcon from "@icons/CalendarIcon";
import ClockIcon from "@icons/ClockIcon";
import { timeStringFromDate } from "@lib/datetime";
import { preferencesQueryOptions } from "@lib/queries";
import { useQuery } from "@tanstack/react-query";

function monthDateFormat(date) {
  return date.toLocaleDateString([], {
    weekday: "short",
    month: "short",
    day: "numeric",
  });
}

export default function BookingItem({ booking, customerName }) {
  const now = new Date();
  const toDate = new Date(booking.to_date);
  const { data: preferences } = useQuery(preferencesQueryOptions());

  let statusLabel = "Completed";
  let statusStyle = "bg-green-600/20 text-green-600";

  if (booking.is_cancelled) {
    statusLabel = `Cancelled by ${customerName}`;
    statusStyle = "bg-red-600/20 text-red-600";
  } else if (toDate > now) {
    statusLabel = "Upcoming";
    statusStyle = "bg-primary/20 text-primary";
  }

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-center justify-between gap-4">
        <h4 className="text-text_color font-semibold">
          {booking.service_name}
        </h4>
        <span
          className={`rounded-full px-2 py-1 text-xs font-medium ${statusStyle}`}
        >
          {statusLabel}
        </span>
      </div>
      <div
        className="flex flex-wrap items-center gap-4 text-sm text-gray-500
          dark:text-gray-400"
      >
        <span className="flex items-center gap-2">
          <CalendarIcon styles="size-4 text-gray-500 dark:text-gray-400" />
          <span className="mt-0.5 text-sm">{`${monthDateFormat(new Date(booking.from_date))}`}</span>
        </span>
        <span className="flex items-center gap-2">
          <ClockIcon styles="size-4 fill-gray-500 dark:fill-gray-400" />
          <span className="mt-0.5 text-sm">{`${timeStringFromDate(new Date(booking.from_date), preferences?.time_format)} - ${timeStringFromDate(new Date(booking.to_date), preferences?.time_format)}`}</span>
        </span>
      </div>
    </div>
  );
}
