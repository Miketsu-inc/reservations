import {
  Comment01Icon,
  Flag01Icon,
  MoreVerticalIcon,
  StarIcon,
} from "@hugeicons/core-free-icons";
import {
  Avatar,
  Button,
  Card,
  Icon,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
  Textarea,
} from "@reservations/components";
import { toAvatar } from "@reservations/lib";
import { useState } from "react";

function formatReviewDate(dateString) {
  if (!dateString) return "";
  const date = new Date(dateString);
  const now = new Date();
  const diffDays = Math.floor((now - date) / (1000 * 60 * 60 * 24));
  if (diffDays === 0) return "Today";
  if (diffDays === 1) return "Yesterday";
  if (diffDays < 7) return `${diffDays} days ago`;
  if (diffDays < 30) return `${Math.floor(diffDays / 7)}w ago`;
  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    year: date.getFullYear() !== now.getFullYear() ? "numeric" : undefined,
  }).format(date);
}

export default function ReviewCard({
  review,
  onReply,
  onDeleteReply,
  onReport,
}) {
  const [isReplying, setIsReplying] = useState(false);
  const [replyText, setReplyText] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const customer = toAvatar(review.customer);
  const employee = review.employee ? toAvatar(review.employee) : null;
  const replyAuthor = review.reply?.employee
    ? toAvatar(review.reply.employee)
    : null;
  const isReported = review.status === "reported";

  function startReplying() {
    setReplyText(review.reply?.comment ?? "");
    setIsReplying(true);
  }

  async function handleSendReply() {
    const comment = replyText.trim();
    if (!comment) return;
    setIsSubmitting(true);
    try {
      await onReply(review.id, comment);
      setIsReplying(false);
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <Card styles="p-5 flex flex-col gap-3">
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-center gap-3">
          <Avatar
            styles="size-11! text-sm! shrink-0"
            img={customer.img}
            initials={customer.initials}
          />
          <div>
            <div className="flex items-center gap-2">
              <span
                className="max-w-24 truncate text-sm font-semibold
                  md:max-w-full"
              >
                {customer.name}
              </span>
              {isReported && (
                <span
                  className="bg-warning-subtle text-warning hidden rounded-lg
                    px-3 py-0.5 text-xs md:block"
                >
                  Reported
                </span>
              )}
            </div>
            <span className="text-fg-muted text-xs">
              {formatReviewDate(review.created_at)}
            </span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <div className="flex gap-1">
            {[1, 2, 3, 4, 5].map((star) => (
              <Icon
                key={star}
                icon={StarIcon}
                styles={`size-4 sm:size-5 ${
                  star <= review.rating
                    ? "fill-warning text-warning"
                    : "text-border-strong"
                }`}
              />
            ))}
          </div>

          <Popover>
            <PopoverTrigger asChild>
              <button
                type="button"
                className="text-fg-muted hover:bg-hover cursor-pointer
                  rounded-lg p-1"
                aria-label="Review actions"
              >
                <Icon icon={MoreVerticalIcon} styles="size-5" />
              </button>
            </PopoverTrigger>
            <PopoverContent align="end">
              {isReported ? (
                <p className="text-fg-muted p-2 text-sm">
                  Reported - under review
                </p>
              ) : (
                <PopoverClose asChild>
                  <button
                    type="button"
                    onClick={() => onReport(review)}
                    className="text-fg-danger hover:bg-hover flex w-full
                      items-center gap-2 rounded-md p-2 text-sm"
                  >
                    <Icon icon={Flag01Icon} styles="size-4" />
                    Report review
                  </button>
                </PopoverClose>
              )}
            </PopoverContent>
          </Popover>
        </div>
      </div>

      {review.comment && (
        <p className="text-sm leading-relaxed whitespace-pre-wrap">
          {review.comment}
        </p>
      )}

      <p className="text-fg-muted text-xs">
        {review.service_name} {employee && `with ${employee.name}`}
      </p>

      {isReplying ? (
        <div
          className="bg-bg-sunken/40 border-border flex flex-col gap-3
            rounded-lg border p-3.5"
        >
          <Textarea
            styles="p-2.5 text-sm min-h-20"
            placeholder="Thank the client or address their feedback publicly..."
            maxLength={300}
            value={replyText}
            inputData={(e) => setReplyText(e.value)}
          />
          <div className="flex justify-end gap-2">
            <Button
              variant="secondary"
              styles="px-3 py-1.5 text-xs"
              buttonText="Cancel"
              type="button"
              onClick={() => setIsReplying(false)}
            />
            <Button
              variant="primary"
              styles="px-3 py-1.5 text-xs"
              buttonText={review.reply ? "Save changes" : "Post reply"}
              type="button"
              isLoading={isSubmitting}
              onClick={handleSendReply}
            />
          </div>
        </div>
      ) : review.reply ? (
        <div
          className="bg-bg-sunken/40 border-border flex flex-col gap-1.5
            rounded-lg border px-3 py-3 sm:px-5"
        >
          <div className="flex items-center gap-2.5 text-xs">
            <Avatar
              styles="size-6! text-[9px]! shrink-0"
              img={replyAuthor?.img}
              initials={replyAuthor?.initials}
            />
            <span className="font-semibold">Your reply</span>
            <span className="text-fg-muted">
              {formatReviewDate(review.reply.created_at)}
            </span>
          </div>
          <p className="text-sm leading-relaxed whitespace-pre-wrap">
            {review.reply.comment}
          </p>
          <div className="text-fg-muted flex gap-3 text-xs">
            <button
              type="button"
              onClick={startReplying}
              className="hover:text-fg cursor-pointer"
            >
              Edit
            </button>
            <button
              type="button"
              onClick={() => onDeleteReply(review.id)}
              className="hover:text-fg-danger cursor-pointer"
            >
              Delete
            </button>
          </div>
        </div>
      ) : (
        <Button
          variant="secondary"
          styles="px-3 py-1.5 text-xs gap-1.5 w-fit"
          buttonText="Reply"
          type="button"
          onClick={startReplying}
        >
          <Icon icon={Comment01Icon} styles="size-3.5" />
        </Button>
      )}
    </Card>
  );
}
