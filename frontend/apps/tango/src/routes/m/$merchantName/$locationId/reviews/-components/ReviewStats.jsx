import { StarIcon } from "@hugeicons/core-free-icons";
import { Icon, StarRating } from "@reservations/components";

export default function ReviewStats({ stats, selectedStar, onSelectStar }) {
  const total = stats.total_reviews;
  const average = stats.average_rating;

  const breakdown = [
    { star: 5, count: stats.count_5_star },
    { star: 4, count: stats.count_4_star },
    { star: 3, count: stats.count_3_star },
    { star: 2, count: stats.count_2_star },
    { star: 1, count: stats.count_1_star },
  ];

  return (
    <div
      className="flex flex-col gap-4 sm:flex-row sm:items-center sm:gap-16
        sm:px-4 lg:sticky lg:top-20 lg:flex-col lg:items-start lg:gap-6 lg:p-0"
    >
      <div className="flex flex-col gap-3 px-2 sm:items-center lg:items-start">
        <span className="text-6xl font-extrabold tracking-tight lg:text-7xl">
          {average.toFixed(1)}
        </span>
        <StarRating
          value={Math.round(average)}
          iconStyles="size-5"
          styles="gap-1.5"
        />
        <span className="text-fg-muted text-sm">
          {`${total} ${total === 1 ? "review" : "reviews"}`}
        </span>
      </div>

      <div className="flex w-full flex-1 flex-col gap-2.5 lg:max-w-xs">
        {breakdown.map(({ star, count }) => {
          const pct = total > 0 ? (count / total) * 100 : 0;
          const isSelected = selectedStar === star;

          return (
            <button
              key={star}
              type="button"
              onClick={() => onSelectStar(isSelected ? "all" : star)}
              className={`group flex w-full cursor-pointer items-center gap-3
              rounded-lg px-2 py-0.5 text-sm transition-colors ${
                isSelected
                  ? "bg-bg-raised ring-border-strong font-semibold ring-1"
                  : "hover:bg-hover"
              }`}
            >
              <span className="flex items-center gap-1.5">
                {star}
                <Icon
                  icon={StarIcon}
                  styles="size-3 fill-warning text-warning"
                />
              </span>
              <div
                className="bg-border-subtle h-2 flex-1 overflow-hidden
                  rounded-full"
              >
                <div
                  className="bg-warning h-full rounded-full transition-all
                    duration-300"
                  style={{ width: `${pct}%` }}
                />
              </div>
              <span className="text-fg-muted group-hover:text-fg">{count}</span>
            </button>
          );
        })}
      </div>
    </div>
  );
}
