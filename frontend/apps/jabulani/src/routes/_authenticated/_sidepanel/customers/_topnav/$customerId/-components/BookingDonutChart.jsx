import { getBookingStatusColorStyles } from "@reservations/jabulani/lib";
import { Pie, PieChart, Sector, Tooltip } from "recharts";

function BookingStatusSector({ isActive, payload, ...props }) {
  const status = payload?.status;

  return (
    <Sector
      {...props}
      className={status ? getBookingStatusColorStyles(status) : undefined}
      fill={status ? "currentColor" : "rgb(var(--hvr-gray))"}
      outerRadius={
        isActive && status ? props.outerRadius + 3 : props.outerRadius
      }
    />
  );
}

function BookingStatusTooltip({ payload }) {
  const status = payload?.[0]?.payload;
  const hasStatus = Boolean(status?.status);

  return (
    <div
      aria-hidden={!hasStatus}
      className={`border-border_color bg-layer_bg flex h-9 w-36 items-center
        gap-3 rounded-lg border px-3 text-sm shadow-md
        ${hasStatus ? "" : "invisible"}`}
    >
      <span
        className={`${
          hasStatus ? getBookingStatusColorStyles(status.status) : ""
        } size-2.5
          shrink-0 rounded-xs bg-current`}
      />
      <span className="min-w-0 flex-1 truncate">{status?.label}</span>
      <span className="font-semibold">{status?.value}</span>
    </div>
  );
}

export default function BookingDonutChart({ statuses }) {
  const hasBookings = statuses.some((item) => item.value > 0);
  const total = statuses.reduce((sum, item) => sum + item.value, 0);
  const chartData = hasBookings
    ? statuses
    : [{ label: "No bookings", value: 1 }];

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
          shape={BookingStatusSector}
        />
        <Tooltip
          allowEscapeViewBox={{ x: true, y: true }}
          cursor={false}
          offset={14}
          wrapperStyle={{ zIndex: 20 }}
          content={<BookingStatusTooltip />}
        />
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
