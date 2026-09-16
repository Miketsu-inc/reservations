import { Card, ServerError } from "@reservations/components";

export default function CustomerStats({ stats, isLoading, error }) {
  if (error) return <ServerError error={error.message} />;

  return (
    <div>
      <p className="mb-3 text-lg">Overview</p>
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <Stat
          label="Total bookings"
          loading={isLoading}
          value={stats?.times_booked}
        />
        <Stat
          label="Completed"
          loading={isLoading}
          value={stats?.times_completed}
        />
        <Stat
          label="Upcoming"
          loading={isLoading}
          value={stats?.times_upcoming}
        />
        <Stat
          label="Cancelled"
          loading={isLoading}
          value={stats?.times_cancelled_by_user}
        />
      </div>
      {!isLoading && stats && (
        <div
          className="text-text_color/60 mt-3 flex flex-wrap gap-x-5 gap-y-1
            text-sm"
        >
          <span>
            Attendance rate:{" "}
            <strong className="text-text_color font-medium">
              {attendanceRate(stats)}
            </strong>
          </span>
          <span>
            No-shows:{" "}
            <strong className="text-text_color font-medium">
              {stats.times_no_show}
            </strong>
          </span>
          <span>
            First booking:{" "}
            <strong className="text-text_color font-medium">
              {formatMonthYear(stats.first_booking)}
            </strong>
          </span>
        </div>
      )}
    </div>
  );
}

function Stat({ label, loading, value }) {
  return (
    <Card styles="h-auto! p-3!">
      <div className="flex flex-col gap-2">
        <span className="text-text_color/70 text-sm">{label}</span>
        {loading ? (
          <span className="bg-hvr_gray h-6 w-8 animate-pulse rounded" />
        ) : (
          <span className="text-lg font-medium">{value ?? 0}</span>
        )}
      </div>
    </Card>
  );
}

function attendanceRate(stats) {
  const resolved = stats.times_completed + stats.times_cancelled_by_user;
  return resolved === 0
    ? "—"
    : `${Math.round((stats.times_completed / resolved) * 100)}%`;
}

function formatMonthYear(dateString) {
  if (!dateString) return "—";

  return new Date(dateString).toLocaleDateString([], {
    month: "short",
    year: "numeric",
  });
}
