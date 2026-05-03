import { CalendarIcon, ClockIcon } from "@reservations/assets";
import { useAuth } from "@reservations/jabulani/lib";
import { preferencesQueryOptions, timeStringFromDate } from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";

function monthDateFormat(date) {
  return date.toLocaleDateString([], {
    weekday: "short",
    month: "short",
    day: "numeric",
  });
}

const statusMap = {
  completed: {
    getLabel: () => "Completed",
    style: "bg-green-600/20 text-green-700 dark:text-green-500",
  },
  booked: {
    getLabel: () => "Booked",
    style: "bg-primary/10 text-primary",
  },
  confirmed: {
    getLabel: () => "Confirmed",
    style: "bg-primary/25 text-primary font-bold",
  },
  cancelled: {
    getLabel: (name) => `Cancelled by ${name || "User"}`,
    style: "bg-red-600/10 text-red-600",
  },
  "no-show": {
    getLabel: () => "No-show",
    style: "bg-red-600/15 text-red-500",
  },
};

export default function BookingItem({ booking, customerName }) {
  const { merchantId } = useAuth();
  const { data: preferences } = useQuery(preferencesQueryOptions(merchantId));

  const currentStatus = statusMap[booking.status];
  const statusLabel = currentStatus.getLabel(customerName);
  const statusStyle = currentStatus.style;

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-center justify-between gap-4">
        <h4 className="text-text_color font-semibold">
          {booking.service_name}
        </h4>
        <span
          className={`rounded-xl px-3 py-1.5 text-xs font-medium ${statusStyle}`}
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
