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
        <div className="grid flex-1 grid-cols-1 gap-4 sm:grid-cols-3">
          {[0, 1, 2].map((item) => (
            <div
              key={item}
              className="bg-hvr_gray h-14 animate-pulse rounded"
            />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div>
      <p className="mb-4 text-lg">Customer overview</p>
      <div className="flex flex-col gap-6 md:flex-row md:items-center">
        <div
          className="flex flex-col items-center justify-center gap-4 sm:flex-row
            md:justify-start"
        >
          <div className="h-40 w-40 shrink-0">
            <BookingDonutChart
              upcoming={stats.times_upcoming}
              completed={stats.times_completed}
              cancelled={stats.times_cancelled_by_user}
              total={stats.times_booked}
            />
          </div>
          <div className="w-40 space-y-2 text-sm">
            <Legend
              color="bg-primary"
              label="Upcoming"
              value={stats.times_upcoming}
            />
            <Legend
              color="bg-green-600"
              label="Completed"
              value={stats.times_completed}
            />
            <Legend
              color="bg-red-600"
              label="Cancelled"
              value={stats.times_cancelled_by_user}
            />
          </div>
        </div>

        <div
          className="border-border_color grid flex-1 grid-cols-1 gap-5
            sm:grid-cols-3 md:border-l md:pl-6"
        >
          <Insight label="Completed value" value={stats.total_spent} />
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

function Legend({ color, label, value }) {
  return (
    <div className="flex items-center justify-between gap-4">
      <div className="flex min-w-0 items-center gap-2">
        <span className={`${color} size-2.5 shrink-0 rounded-xs`} />
        <span className="text-text_color/70 truncate">{label}</span>
      </div>
      <span className="font-medium">{value}</span>
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
