import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
} from "recharts";

export default function RevenueChart({ data }) {
  return (
    <div className="flex h-full w-full flex-col gap-8">
      <p>Revenue</p>
      <div className="flex-1">
        <ResponsiveContainer height="100%" width="100%">
          <AreaChart
            accessibilityLayer
            data={data}
            margin={{ left: 4, right: 4, top: 4, bottom: 4 }}
          >
            <defs>
              <linearGradient id="fillRevenue" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="rgb(var(--primary))"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="rgb(var(--primary))"
                  stopOpacity={0.1}
                />
              </linearGradient>
            </defs>
            <CartesianGrid
              className="stroke-neutral-200 dark:stroke-neutral-800"
              vertical={false}
            />
            <XAxis
              dataKey="day"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              tick={<CustomTick dataLength={data.length} />}
            />
            <Tooltip
              cursor={false}
              content={<TooltipContent indicator="line" />}
            />
            <Area
              type="monotone"
              dataKey="value"
              stroke="rgb(var(--primary))"
              fillOpacity={0.8}
              fill="url(#fillRevenue)"
              activeDot={{ stroke: "rbg(var(--primary))" }}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

const CustomTick = ({ x, y, payload, index, dataLength }) => {
  if (dataLength >= 15) {
    if (index % 3 !== 0) return null;
  } else {
    if (index % 2 !== 0) return null;
  }

  return (
    <text
      x={x}
      y={y + 10}
      className="fill-gray-500 text-sm dark:fill-gray-400"
      textAnchor="middle"
    >
      {payload.value}
    </text>
  );
};

const TooltipContent = ({ active, payload, label }) => {
  if (active && payload && payload.length) {
    return (
      <div
        className="bg-bg_color flex h-fit min-w-32 flex-col rounded-lg border
          border-gray-200 p-2 text-xs shadow-xl dark:border-gray-800"
      >
        <p>{label}</p>
        {payload.map((item, index) => (
          <div key={index} className="flex flex-row justify-between">
            <div className="flex flex-row items-center gap-2">
              <div className="bg-primary h-2.5 w-2.5 shrink-0 rounded-[2px]"></div>
              <p className="text-gray-500 dark:text-gray-400">Revenue</p>
            </div>
            <p>{item.value}</p>
          </div>
        ))}
      </div>
    );
  }

  return null;
};
