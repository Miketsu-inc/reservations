import {
  Avatar,
  Button,
  Loading,
  ServerError,
  StarRating,
} from "@reservations/components";
import {
  formatRelativeDate,
  formatShortName,
  publicReviewsQueryOptions,
  toAvatar,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";

const PREVIEW_LIMIT = 4;
const PREVIEW_FILTERS = { sort: "most_relevant", limit: PREVIEW_LIMIT };

export default function ReviewsSection({ merchantName, merchantInfo, stats }) {
  const locationId = merchantInfo.location_id;
  const total = stats.total_reviews;

  const { data, isLoading, isError, error } = useQuery(
    publicReviewsQueryOptions(merchantName, locationId, PREVIEW_FILTERS)
  );

  if (isError) {
    return <ServerError error={error.message} />;
  }

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div className="flex flex-col gap-8">
      <div className="flex items-center gap-4">
        <span className="text-5xl font-extrabold">
          {stats.average_rating.toFixed(1)}
        </span>
        <div className="mt-3 flex flex-col gap-1">
          <StarRating value={Math.round(stats.average_rating)} />
          <span className="text-fg-muted text-sm">
            {total} {total === 1 ? "review" : "reviews"}
          </span>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-x-10 gap-y-8 lg:grid-cols-2">
        {data.reviews.map((review) => (
          <Review key={review.id} review={review} />
        ))}
      </div>

      {total > PREVIEW_LIMIT && (
        <Link
          to="/m/$merchantName/$locationId/reviews"
          params={{ merchantName, locationId: String(locationId) }}
          className="w-fit"
        >
          <Button
            variant="secondary"
            buttonText={`See all ${total} reviews`}
            styles="px-6 py-2.5 text-sm"
          />
        </Link>
      )}
    </div>
  );
}

function Review({ review }) {
  const customer = toAvatar(review.customer);
  const serviceInfo = [
    review.service_name,
    review.employee && `with ${formatShortName(review.employee)}`,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <article className="flex flex-col gap-3">
      <div className="flex items-center gap-3">
        <Avatar
          styles="size-14! shrink-0"
          img={customer.img}
          initials={customer.initials}
        />
        <div className="flex min-w-0 flex-1 flex-col gap-2">
          <div className="flex items-center justify-between gap-2">
            <span className="truncate font-semibold">
              {formatShortName(review.customer, "Client")}
            </span>
            <span className="text-fg-muted shrink-0 text-xs">
              {formatRelativeDate(review.updated_at ?? review.created_at)}
            </span>
          </div>
          <StarRating
            value={review.rating}
            iconStyles="size-3.5"
            styles="gap-0.5"
          />
        </div>
      </div>

      {review.comment && (
        <p className="line-clamp-3 text-sm leading-relaxed">{review.comment}</p>
      )}
      {serviceInfo && (
        <p className="text-fg-muted truncate text-xs">{serviceInfo}</p>
      )}
    </article>
  );
}
