import { TradeUpIcon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";

export default function StatisticsChangeIndicator({ styles, percent }) {
  const isNegative = percent < 0;

  return (
    <div className="flex w-fit flex-row items-center gap-2">
      <Icon
        icon={TradeUpIcon}
        styles={`size-5 ${
          isNegative
            ? "text-red-700 dark:text-red-500 rotate-x-180"
            : "text-green-700 dark:text-green-500"
          }`}
      />
      <span
        className={`${styles} font-semibold text-nowrap ${
          isNegative
            ? "text-red-700 dark:text-red-500"
            : "text-green-700 dark:text-green-500"
          }`}
      >
        {`${!isNegative ? "+" : ""}${percent} %`}
      </span>
    </div>
  );
}
