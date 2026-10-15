import { TradeUpIcon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";

export default function StatisticsChangeIndicator({ styles, percent }) {
  const isNegative = percent < 0;

  return (
    <div className="flex w-fit flex-row items-center gap-2">
      <Icon
        icon={TradeUpIcon}
        styles={`size-5 ${
          isNegative ? "text-fg-danger rotate-x-180" : "text-fg-success"
        }`}
      />
      <span
        className={`${styles} font-semibold text-nowrap ${
          isNegative ? "text-fg-danger" : "text-fg-success"
        }`}
      >
        {`${!isNegative ? "+" : ""}${percent} %`}
      </span>
    </div>
  );
}
