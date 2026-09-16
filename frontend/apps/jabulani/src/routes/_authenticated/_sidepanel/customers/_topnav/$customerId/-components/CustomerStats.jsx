import {
  Calendar02Icon,
  Cancel01Icon,
  CheckmarkCircle02Icon,
  User03Icon,
} from "@hugeicons/core-free-icons";
import { Card, Icon, ServerError } from "@reservations/components";

export default function CustomerStats({ stats, isLoading, error }) {
  if (error) return <ServerError error={error.message} />;

  if (isLoading || !stats) {
    return (
      <Card styles="animate-pulse">
        <div className="bg-hvr_gray mb-4 h-6 w-40 rounded" />
        <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
          {[0, 1, 2, 3].map((item) => (
            <div key={item} className="bg-hvr_gray h-24 rounded-lg" />
          ))}
        </div>
      </Card>
    );
  }

  const resolvedBookings =
    stats.times_completed + stats.times_cancelled_by_user;
  const attendanceRate =
    resolvedBookings > 0
      ? Math.round((stats.times_completed / resolvedBookings) * 100)
      : null;

  return (
    <Card styles="p-0! overflow-hidden">
      <div className="border-border_color border-b px-4 py-4 sm:px-6">
        <h2 className="text-text_color text-lg font-semibold">
          Customer insights
        </h2>
        <p className="text-text_color/55 mt-0.5 text-sm">
          Booking activity and attendance at a glance
        </p>
      </div>
      <div className="p-4 sm:p-6">
        <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
          <StatCard
            icon={User03Icon}
            label="Total bookings"
            value={stats.times_booked}
          />
          <StatCard
            icon={CheckmarkCircle02Icon}
            label="Completed"
            styles="text-green-600"
            value={stats.times_completed}
          />
          <StatCard
            icon={Calendar02Icon}
            label="Upcoming"
            styles="text-primary"
            value={stats.times_upcoming}
          />
          <StatCard
            icon={Cancel01Icon}
            label="Cancelled / no-show"
            styles="text-red-600"
            value={stats.times_cancelled_by_user}
          />
        </div>

        <div
          className="border-border_color mt-5 grid gap-4 border-t pt-5
            sm:grid-cols-3"
        >
          <Insight
            label="Attendance rate"
            value={attendanceRate === null ? "—" : `${attendanceRate}%`}
          />
          <Insight label="No-shows" value={stats.times_no_show} />
          <Insight
            label="First booking"
            value={formatMonthYear(stats.first_booking)}
          />
        </div>
      </div>
    </Card>
  );
}

function StatCard({ icon, label, styles = "text-text_color", value }) {
  return (
    <div
      className="bg-bg_color border-border_color rounded-lg border p-3 sm:p-4"
    >
      <div className="flex items-center justify-between gap-2">
        <span className={`${styles} text-2xl font-bold sm:text-3xl`}>
          {value}
        </span>
        <span className="bg-layer_bg rounded-lg p-2">
          <Icon icon={icon} styles={`size-5 ${styles}`} />
        </span>
      </div>
      <p className="text-text_color/60 mt-2 text-xs sm:text-sm">{label}</p>
    </div>
  );
}

function Insight({ label, value }) {
  return (
    <div>
      <p
        className="text-text_color/50 text-xs font-medium tracking-wide
          uppercase"
      >
        {label}
      </p>
      <p className="text-text_color mt-1 font-semibold">{value}</p>
    </div>
  );
}

function formatMonthYear(dateString) {
  if (!dateString) return "—";

  return new Date(dateString).toLocaleDateString([], {
    month: "short",
    year: "numeric",
  });
}
