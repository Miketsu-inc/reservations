import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";

const COLORS = {
  Completed: "#16a34a",
  Cancelled: "#dc2626",
  Upcoming: "rgb(var(--primary))",
  Empty: "rgb(var(--hvr-gray))",
};

export default function BookingDonutChart({ upcoming, cancelled, completed }) {
  const chartData = [
    { name: "Upcoming", value: upcoming },
    { name: "Cancelled", value: cancelled },
    { name: "Completed", value: completed },
  ];

  const total = chartData.reduce((sum, d) => sum + d.value, 0);
  // const completionRate = total > 0 ? Math.round((completed / total) * 100) : 0;

  const fallbackData = [{ name: "Empty", value: 1 }];

  return (
    <ResponsiveContainer width="100%" height="100%">
      <PieChart>
        <Pie
          data={total > 0 ? chartData : fallbackData}
          dataKey="value"
          nameKey="name"
          cx="50%"
          cy="50%"
          innerRadius={60}
          outerRadius={75}
          paddingAngle={total > 0 ? 2 : 0}
          stroke="none"
          isAnimationActive={true}
        >
          {(total > 0 ? chartData : fallbackData).map((entry, index) => (
            <Cell
              key={`cell-${index}`}
              fill={COLORS[entry.name] || COLORS.Empty}
            />
          ))}
          {/* <Label
            // content={<LabelContent completionRate={completionRate} />}
            position="center"
          /> */}
        </Pie>
        {total > 0 && <Tooltip content={<TooltipContent total={total} />} />}
      </PieChart>
    </ResponsiveContainer>
  );
}

const TooltipContent = ({ payload, total }) => {
  const data = payload[0]?.payload;
  const percentage = total > 0 ? Math.round((data?.value / total) * 100) : 0;

  return (
    <div
      className="bg-bg_color flex h-fit min-w-32 flex-col rounded-lg border
        border-gray-200 p-2 text-xs shadow-xl dark:border-gray-800"
    >
      <div
        className="text-text_color flex items-center justify-between
          font-semibold"
      >
        <div className="flex items-center gap-2">
          <span
            className="size-2.5 shrink-0 rounded-xs"
            style={{ backgroundColor: data?.fill }}
          />
          {data?.name}
        </div>
        <span>{data?.value}</span>
      </div>
      <div className="text-text_color/70 mt-1">Rate: {percentage}%</div>
    </div>
  );
};

// const LabelContent = ({ viewBox, completionRate }) => {
//   const { cx, cy } = viewBox;

//   return (
//     <g>
//       <text
//         x={cx + 3}
//         y={cy - 4}
//         textAnchor="middle"
//         className="fill-text_color text-2xl font-bold"
//       >
//         {`${completionRate}%`}
//       </text>
//       <text
//         x={cx + 3}
//         y={cy + 22}
//         textAnchor="middle"
//         className="fill-gray-500 text-xs dark:fill-gray-400"
//       >
//         Completion
//       </text>
//     </g>
//   );
// };
