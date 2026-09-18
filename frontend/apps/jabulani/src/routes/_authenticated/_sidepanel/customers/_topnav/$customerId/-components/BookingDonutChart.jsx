import { Cell, Pie, PieChart } from "recharts";

const COLORS = {
  upcoming: "rgb(var(--primary))",
  completed: "#16a34a",
  cancelled: "#dc2626",
  empty: "rgb(var(--hvr-gray))",
};

export default function BookingDonutChart({
  upcoming,
  completed,
  cancelled,
  total,
}) {
  const data = [
    { name: "Upcoming", value: upcoming, color: COLORS.upcoming },
    { name: "Completed", value: completed, color: COLORS.completed },
    { name: "Cancelled", value: cancelled, color: COLORS.cancelled },
  ];
  const hasBookings = data.some((item) => item.value > 0);
  const chartData = hasBookings
    ? data
    : [{ name: "No bookings", value: 1, color: COLORS.empty }];

  return (
    <div className="relative size-40">
      <PieChart
        width={160}
        height={160}
        margin={{ top: 8, right: 8, bottom: 8, left: 8 }}
      >
        <Pie
          data={chartData}
          dataKey="value"
          cx="50%"
          cy="50%"
          innerRadius={52}
          outerRadius={68}
          paddingAngle={hasBookings ? 2 : 0}
          stroke="none"
        >
          {chartData.map((item) => (
            <Cell key={item.name} fill={item.color} />
          ))}
        </Pie>
      </PieChart>
      <div
        className="pointer-events-none absolute inset-0 flex flex-col
          items-center justify-center"
      >
        <span className="text-2xl leading-none font-semibold">{total}</span>
        <span className="text-text_color/60 mt-1 text-xs">bookings</span>
      </div>
    </div>
  );
}
