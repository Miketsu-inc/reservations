import { Card, ServerError } from "@reservations/components";

export default function CustomerStats({ stats, isLoading, error }) {
  if (error) return <ServerError error={error.message} />;

  const values = stats
    ? [
        attendanceRate(stats),
        visitFrequency(stats),
        noShowRate(stats),
        formatMonthYear(stats.first_booking),
      ]
    : [];

  return (
    <div>
      <p className="mb-3 text-lg">Customer insights</p>
      <Card styles="h-auto!">
        <div className="grid grid-cols-2 gap-x-4 gap-y-5 sm:grid-cols-4">
          <Insight
            label="Attendance rate"
            loading={isLoading}
            value={values[0]}
          />
          <Insight
            label="Visit frequency"
            loading={isLoading}
            value={values[1]}
          />
          <Insight label="No-show rate" loading={isLoading} value={values[2]} />
          <Insight
            label="First booking"
            loading={isLoading}
            value={values[3]}
          />
        </div>
      </Card>
    </div>
  );
}

function Insight({ label, loading, value }) {
  return (
    <div className="flex min-w-0 flex-col gap-2">
      <span className="text-text_color/60 text-sm">{label}</span>
      {loading ? (
        <span className="bg-hvr_gray h-6 w-16 animate-pulse rounded" />
      ) : (
        <span className="truncate text-lg font-medium">{value ?? "—"}</span>
      )}
    </div>
  );
}

function attendanceRate(stats) {
  const resolved = resolvedBookings(stats);
  return resolved === 0
    ? "—"
    : `${Math.round((stats.times_completed / resolved) * 100)}%`;
}

function noShowRate(stats) {
  const resolved = resolvedBookings(stats);
  return resolved === 0
    ? "—"
    : `${Math.round((stats.times_no_show / resolved) * 100)}%`;
}

function visitFrequency(stats) {
  if (!stats.first_booking || stats.times_completed === 0) return "—";

  const firstBooking = new Date(stats.first_booking);
  const elapsedMonths = Math.max(
    1,
    (Date.now() - firstBooking.getTime()) / (1000 * 60 * 60 * 24 * 30.44)
  );
  const visitsPerMonth = stats.times_completed / elapsedMonths;

  return `${visitsPerMonth.toFixed(visitsPerMonth >= 10 ? 0 : 1)} / month`;
}

function resolvedBookings(stats) {
  return stats.times_completed + stats.times_cancelled_by_user;
}

function formatMonthYear(dateString) {
  if (!dateString) return "—";

  return new Date(dateString).toLocaleDateString([], {
    month: "short",
    year: "numeric",
  });
}
