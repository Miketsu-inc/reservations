import ArrowTrendIcon from "@icons/ArrowTrendIcon";

export default function StatisticsChangeIndicator({ styles, percent }) {
  const isNegative = percent < 0;

  return (
    <div className="flex w-fit flex-row items-center gap-2">
      <ArrowTrendIcon
        styles={`size-4
          ${isNegative ? "fill-red-700 dark:fill-red-500 rotate-x-180" : "fill-green-700 dark:fill-green-500"}`}
      />
      <span
        className={`${styles} font-semibold text-nowrap
          ${isNegative ? "text-red-700 dark:text-red-500" : "text-green-700 dark:text-green-500"}`}
      >
        {`${!isNegative ? "+" : ""}${percent} %`}
      </span>
    </div>
  );
}
