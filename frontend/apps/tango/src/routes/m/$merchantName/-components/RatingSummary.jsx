import { StarIcon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";

export default function RatingSummary({ average, styles, children }) {
  return (
    <span className={`flex items-center gap-1 ${styles}`}>
      <Icon icon={StarIcon} styles="size-[1.1em] fill-warning text-warning" />
      <span className={`text-fg font-medium ${styles}`}>
        {average.toFixed(1)}
      </span>
      {children}
    </span>
  );
}
