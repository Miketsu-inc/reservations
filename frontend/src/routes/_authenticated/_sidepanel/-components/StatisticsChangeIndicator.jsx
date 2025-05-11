export default function StatisticsChangeIndicator({ styles, percent }) {
  const isNegative = percent < 0;

  return (
    <div
      className={`${isNegative ? "bg-red-600/25 dark:bg-red-600/15" : "bg-green-600/25 dark:bg-green-600/15"}
        flex w-fit flex-row items-center gap-2 rounded-lg px-2 py-1`}
    >
      <span
        className={`${styles} font-semibold text-nowrap
          ${isNegative ? "text-red-700 dark:text-red-500" : "text-green-700 dark:text-green-500"}`}
      >
        {`${!isNegative ? "+" : ""}${percent} %`}
      </span>
    </div>
  );
}
