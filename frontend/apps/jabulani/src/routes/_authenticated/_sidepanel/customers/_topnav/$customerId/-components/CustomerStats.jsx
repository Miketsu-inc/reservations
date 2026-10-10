import { Loading, ServerError } from "@reservations/components";
import { BOOKING_STATUS_OPTIONS, useAuth } from "@reservations/jabulani/lib";
import { useQuery } from "@tanstack/react-query";
import BookingDonutChart from "./BookingDonutChart";
import { customerStatsQueryOptions } from "./customerQueries";

export default function CustomerStats({ customerId }) {
  const { merchantId } = useAuth();
  const {
    data: stats,
    isLoading,
    error,
  } = useQuery(customerStatsQueryOptions(merchantId, customerId));

  return (
    <section>
      <h2 className="mb-4 text-xl">Overview</h2>
      {error ? (
        <ServerError error={error.message} />
      ) : isLoading || !stats ? (
        <Loading />
      ) : (
        <CustomerStatsContent stats={stats} />
      )}
    </section>
  );
}

function CustomerStatsContent({ stats }) {
  const counts = {
    booked: stats.times_booked,
    confirmed: stats.times_confirmed,
    completed: stats.times_completed,
    cancelled: stats.times_cancelled,
    "no-show": stats.times_no_show,
  };
  const statuses = BOOKING_STATUS_OPTIONS.map(({ value: status, label }) => ({
    status,
    label,
    value: counts[status],
  }));

  return (
    <div className="flex flex-col gap-6 md:flex-row md:items-center">
      <div className="h-40 w-40 shrink-0 self-center md:self-auto">
        <BookingDonutChart statuses={statuses} />
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
