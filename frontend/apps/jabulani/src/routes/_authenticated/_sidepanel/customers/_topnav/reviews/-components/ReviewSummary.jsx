import { StarIcon } from "@hugeicons/core-free-icons";
import { Card, Icon, Loading, ServerError } from "@reservations/components";

export default function ReviewSummary({
  stats,
  isLoading,
  isError,
  error,
  selectedRating,
  onSelectRating,
}) {
  if (isError) {
    return (
      <Card styles="p-6">
        <ServerError
          error={error?.message || "Failed to load review statistics"}
        />
      </Card>
    );
  }

  if (isLoading || !stats) {
    return (
      <Card styles="p-6 flex items-center justify-center min-h-40">
        <Loading />
      </Card>
    );
  }

  const total = stats.total_reviews;
  const average = Number(stats.average_rating || 0);

  const breakdown = [
    { star: 5, count: stats.count_5_star },
    { star: 4, count: stats.count_4_star },
    { star: 3, count: stats.count_3_star },
    { star: 2, count: stats.count_2_star },
    { star: 1, count: stats.count_1_star },
  ];

  return (
    <Card styles="p-6">
      <div className="grid gap-6 md:grid-cols-12 md:gap-8">
        <div className="flex flex-col items-center justify-center md:col-span-4">
          <span className="text-fg text-5xl font-bold">
            {average.toFixed(1)}
          </span>
          <div className="mt-2 flex gap-1">
            {[1, 2, 3, 4, 5].map((star) => (
              <Icon
                key={star}
                icon={StarIcon}
                styles={`size-5 ${
                  star <= Math.round(average)
                    ? "fill-warning text-warning"
                    : "text-border-strong"
                }`}
              />
            ))}
          </div>
          <p className="text-fg-muted mt-2 text-xs">
            {total} {total === 1 ? "review" : "reviews"}
          </p>
        </div>

        <div className="flex flex-col gap-2 md:col-span-8">
          {breakdown.map(({ star, count }) => {
            const pct = total > 0 ? (count / total) * 100 : 0;
            const isSelected = String(selectedRating) === String(star);

            return (
              <button
                key={star}
                type="button"
                onClick={() => onSelectRating?.(isSelected ? "all" : star)}
                className={`flex cursor-pointer items-center gap-4 rounded-lg
                px-2 py-1 text-xs
                ${isSelected ? "bg-hover font-semibold" : "hover:bg-hover"}`}
              >
                <span className="text-fg flex items-center justify-end gap-1.5">
                  {star}
                  <Icon
                    icon={StarIcon}
                    styles="size-3 fill-warning text-warning"
                  />
                </span>

                <div
                  className="bg-border-subtle h-2.5 flex-1 overflow-hidden
                    rounded-full"
                >
                  <div
                    className="bg-warning h-full w-(--w) rounded-full
                      transition-[width] duration-700 starting:w-0"
                    style={{ "--w": `${pct}%` }}
                  />
                </div>

                <span className="text-fg-muted">{count}</span>
              </button>
            );
          })}
        </div>
      </div>
    </Card>
  );
}
