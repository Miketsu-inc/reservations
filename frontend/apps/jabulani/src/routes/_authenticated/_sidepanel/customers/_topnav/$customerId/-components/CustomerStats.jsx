import { ServerError } from "@reservations/components";
import BookingDonutChart from "./BookingDonutChart";

export default function CustomerStats({ stats, isLoading, error }) {
  if (error) return <ServerError error={error.message} />;

  if (isLoading || !stats) {
    return (
      <div className="flex flex-col gap-5 md:flex-row md:items-center">
        <div
          className="bg-hvr_gray size-40 animate-pulse self-center rounded-full"
        />
        <div className="grid flex-1 grid-cols-2 gap-4 lg:grid-cols-4">
          {[0, 1, 2, 3].map((item) => (
            <div
              key={item}
              className="bg-hvr_gray h-14 animate-pulse rounded"
            />
          ))}
        </div>
      </div>
    );
  }

  const statuses = [
    { status: "booked", label: "Booked", value: stats.times_booked },
    {
      status: "confirmed",
      label: "Confirmed",
      value: stats.times_confirmed,
    },
    {
      status: "completed",
      label: "Completed",
      value: stats.times_completed,
    },
    {
      status: "cancelled",
      label: "Cancelled",
      value: stats.times_cancelled,
    },
    { status: "no-show", label: "No-show", value: stats.times_no_show },
  ];
  return (
    <div>
      <p className="mb-4 text-xl">Overview</p>
      <div className="flex flex-col gap-6 md:flex-row md:items-center">
        <div className="h-40 w-40 shrink-0 self-center md:self-auto">
          <BookingDonutChart statuses={statuses} total={stats.total_bookings} />
        </div>

        <div
          className="border-border_color grid flex-1 grid-cols-2 gap-x-6 gap-y-5
            border-t pt-6 md:border-t-0 md:border-l md:pt-0 md:pl-6
            lg:grid-cols-4"
        >
          <Insight
            label="First booking"
            value={formatFirstBookingDate(stats.first_booking)}
          />
          <Insight
            label="Completed value"
            value={formatCompletedValues(stats.completed_values)}
          />
          <Insight
            label="Favorite service"
            value={stats.favorite_service ?? "—"}
          />
          <Insight
            label="Next booking"
            value={formatBookingDate(stats.next_booking)}
          />
        </div>
      </div>
    </div>
  );
}

function Insight({ label, value }) {
  return (
    <div className="min-w-0">
      <p className="text-text_color/60 text-sm">{label}</p>
      <p className="mt-1 truncate font-medium" title={value}>
        {value}
      </p>
    </div>
  );
}

function formatBookingDate(dateString) {
  if (!dateString) return "None scheduled";

  return new Date(dateString).toLocaleDateString([], {
    weekday: "short",
    month: "short",
    day: "numeric",
  });
}

function formatFirstBookingDate(dateString) {
  if (!dateString) return "No bookings yet";

  return new Date(dateString).toLocaleDateString([], {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function formatCompletedValues(values) {
  if (!values?.length) return "—";

  return values.map(({ formatted_value }) => formatted_value).join(" + ");
}
