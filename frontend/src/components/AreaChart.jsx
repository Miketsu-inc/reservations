import {
  Area,
  AreaChart as ReAreaChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

export default function AreaChart({ data }) {
  return (
    <ResponsiveContainer height="100%" width="100%">
      <ReAreaChart
        data={data}
        margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
      >
        <defs>
          <linearGradient id="colorUv" x1="0" y1="0" x2="0" y2="1">
            <stop offset="10%" stopColor="#8884d8" stopOpacity={0.8} />
            <stop offset="90%" stopColor="#8884d8" stopOpacity={0} />
          </linearGradient>
        </defs>
        <XAxis interval="preserveStartEnd" dataKey="day" />
        <YAxis />
        <Tooltip />
        <Area
          type="monotone"
          dataKey="value"
          stroke="#8884d8"
          fillOpacity={1}
          fill="url(#colorUv)"
          // animationDuration={0}
        />
      </ReAreaChart>
    </ResponsiveContainer>
  );
}
