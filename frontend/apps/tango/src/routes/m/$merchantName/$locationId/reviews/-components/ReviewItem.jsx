import { Avatar, StarRating } from "@reservations/components";
import {
  formatRelativeDate,
  formatShortName,
  toAvatar,
} from "@reservations/lib";

export default function ReviewItem({ review, merchantInfo }) {
  const customer = toAvatar(review.customer);
  const serviceInfo = [
    review.service_name,
    review.employee && `with ${formatShortName(review.employee)}`,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <article className="flex flex-col gap-4 py-6 first:pt-0 last:pb-0">
      <div className="flex gap-3.5">
        <div className="flex w-11 shrink-0 flex-col items-center gap-2 sm:w-13">
          <Avatar
            styles="size-11! sm:size-13! text-sm!"
            img={customer.img}
            initials={customer.initials}
          />
          {review.reply && (
            <div className="border-border-strong flex-1 border-l" />
          )}
        </div>

        <div className="flex min-w-0 flex-1 flex-col gap-3">
          <div className="flex items-start justify-between gap-3">
            <div className="flex flex-col gap-1.5">
              <span className="text-base font-semibold">
                {formatShortName(review.customer, "Client")}
              </span>
              <StarRating
                value={review.rating}
                iconStyles="size-4"
                styles="gap-0.5"
              />
            </div>

            <span className="text-fg-muted shrink-0 text-sm">
              {formatRelativeDate(review.updated_at ?? review.created_at)}
            </span>
          </div>

          {(review.comment || serviceInfo) && (
            <div className="flex flex-col gap-2">
              {review.comment && (
                <p className="text-base leading-relaxed whitespace-pre-wrap">
                  {review.comment}
                </p>
              )}
              {serviceInfo && (
                <p className="text-fg-muted text-sm">{serviceInfo}</p>
              )}
            </div>
          )}
        </div>
      </div>

      {review.reply && (
        <div className="flex gap-3.5">
          <div className="relative w-11 shrink-0 sm:w-13">
            <div
              className="border-border-strong absolute -top-4
                left-[calc(50%-0.5px)] h-8 w-[calc(50%+0.5px+0.625rem)]
                rounded-bl-xl border-b border-l"
            />
          </div>

          <div className="flex min-w-0 flex-1 items-start gap-3">
            <Avatar
              styles="size-8! text-xs! shrink-0"
              img={merchantInfo.logo_url}
              initials={merchantInfo.merchant_name.charAt(0)}
            />
            <div className="flex min-w-0 flex-col gap-0.5">
              <span className="text-sm font-semibold">
                {merchantInfo.merchant_name}
              </span>
              <p className="text-sm leading-relaxed whitespace-pre-wrap">
                {review.reply.comment}
              </p>
            </div>
          </div>
        </div>
      )}
    </article>
  );
}
