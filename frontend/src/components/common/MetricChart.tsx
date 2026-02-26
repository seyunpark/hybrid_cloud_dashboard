import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';

interface MetricDataPoint {
  time: string;
  value: number;
}

interface MetricChartProps {
  data: MetricDataPoint[];
  title: string;
  color?: string;
  unit?: string;
  height?: number;
}

export function MetricChart({
  data,
  title,
  color = '#3b82f6',
  unit = '',
  height = 200,
}: MetricChartProps) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4">
      <h3 className="mb-2 text-sm font-medium text-gray-700">{title}</h3>
      {data.length === 0 ? (
        <div
          className="flex items-center justify-center text-sm text-gray-400"
          style={{ height }}
        >
          No data available
        </div>
      ) : (
        <ResponsiveContainer width="100%" height={height}>
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
            <XAxis
              dataKey="time"
              tick={{ fontSize: 11 }}
              stroke="#9ca3af"
            />
            <YAxis
              tick={{ fontSize: 11 }}
              stroke="#9ca3af"
              tickFormatter={(v: number) => `${v}${unit}`}
            />
            <Tooltip
              formatter={(value) => [`${value}${unit}`, title]}
            />
            <Line
              type="monotone"
              dataKey="value"
              stroke={color}
              strokeWidth={2}
              dot={false}
              activeDot={{ r: 4 }}
            />
          </LineChart>
        </ResponsiveContainer>
      )}
    </div>
  );
}
