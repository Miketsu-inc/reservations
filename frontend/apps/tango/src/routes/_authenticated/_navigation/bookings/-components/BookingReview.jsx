import {
  BanIcon,
  Edit03Icon,
  SquareLock02Icon,
  StarIcon,
  TrashIcon,
} from "@hugeicons/core-free-icons";
import { Button, Card, Icon } from "@reservations/components";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import ReviewModal from "./ReviewModal";

const FIFTEEN_DAYS_MS = 15 * 24 * 60 * 60 * 1000;

function getReviewView(review, bookingId, isWithin15Days) {
  if (!review) return isWithin15Days ? "none" : "hidden";
  if (review.status === "removed") return "removed";
  if (review.booking_id === bookingId) return "own";
  if (review.status === "reported") return "hidden";
  return isWithin15Days ? "previous" : "hidden";
}

const CLOSED_MODAL = { open: false, mode: "create", rating: 5 };

export default function BookingReview({ booking }) {
  const queryClient = useQueryClient();
  const { showToast } = useToast();
  const [modal, setModal] = useState(CLOSED_MODAL);
  const [isDeleting, setIsDeleting] = useState(false);

  const toDate = booking?.to_date ? new Date(booking.to_date) : null;
  const isWithin15Days = toDate
    ? Date.now() - toDate.getTime() <= FIFTEEN_DAYS_MS
    : false;

  const review = booking.review;
  const view = getReviewView(review, booking.id, isWithin15Days);

  const openModal = (mode, rating = 5) =>
    setModal({ open: true, mode, rating });

  async function handleDelete() {
    if (!review?.id || isDeleting) return;

    setIsDeleting(true);
    try {
      const response = await fetch(`/api/v1/users/reviews/${review.id}`, {
        method: "DELETE",
        headers: { Accept: "application/json" },
      });

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json().catch(() => null);
        showToast({
          message: result?.error?.message,
          variant: "error",
        });
        return;
      }

      await queryClient.invalidateQueries({
        queryKey: ["user-booking-details", String(booking.id)],
      });
    } catch (err) {
      showToast({
        message: err.message,
        variant: "error",
      });
    } finally {
      setIsDeleting(false);
    }
  }

  if (view === "hidden") return null;

  return (
    <>
      <div className="flex flex-col gap-2 pt-2">
        {(view === "own" || view === "removed") && (
          <h3 className="text-fg text-sm sm:text-base">Your review</h3>
        )}

        {view === "removed" && <RemovedCard />}

        {view === "none" && (
          <PromptCard
            merchantName={booking.merchant_name}
            onRate={(rating) => openModal("create", rating)}
          />
        )}

        {view === "own" && (
          <OwnReviewCard
            review={review}
            canEdit={isWithin15Days}
            onEdit={() => openModal("edit", review?.rating || 5)}
            onDelete={handleDelete}
            isDeleting={isDeleting}
          />
        )}

        {view === "previous" && (
          <PreviousReviewCard
            serviceName={booking.service_name}
            employeeName={booking.employee_first_name}
            onUpdate={() => openModal("edit", review?.rating || 5)}
          />
        )}
      </div>

      <ReviewModal
        open={modal.open}
        onOpenChange={(open) => setModal((m) => ({ ...m, open }))}
        booking={booking}
        review={review}
        mode={modal.mode}
        initialRating={modal.rating}
      />
    </>
  );
}

function RemovedCard() {
  return (
    <Card styles="p-4 flex flex-col gap-3">
      <div className="flex items-center gap-2">
        <div
          className="bg-danger-subtle flex items-center justify-center
            rounded-lg p-1.5"
        >
          <Icon icon={BanIcon} styles="size-5 text-fg-danger" />
        </div>
        <h4 className="text-fg font-medium">Your review was removed</h4>
      </div>
      <p className="text-fg-muted text-sm">
        It was taken down by our moderators, for breaking the review guidelines.
        You can’t review this booking, but you can after your next visit
      </p>
    </Card>
  );
}

function PromptCard({ merchantName, onRate }) {
  return (
    <Card styles="p-4 flex flex-col gap-3">
      <div className="flex flex-col gap-1">
        <h4 className="text-fg font-medium">How was your visit?</h4>
        <p className="text-fg-muted text-sm">
          Your feedback helps {merchantName} and other customers.
        </p>
      </div>
      <div className="pt-1">
        <StarRating onChange={onRate} iconStyles="size-7" styles="gap-1.5" />
      </div>
    </Card>
  );
}

function OwnReviewCard({ review, canEdit, onEdit, onDelete, isDeleting }) {
  const isReported = review.status === "reported";

  return (
    <Card styles="p-4 flex flex-col gap-3">
      <div className="flex items-center justify-between gap-4">
        <StarRating value={review.rating || 0} />

        {isReported ? (
          <div
            className="bg-warning-subtle text-fg-warning flex items-center gap-2
              rounded-md px-2 py-1.5 text-xs"
          >
            <Icon icon={SquareLock02Icon} styles="size-3.5" />
            <span>Under moderation</span>
          </div>
        ) : (
          <div className="flex items-center gap-2">
            {canEdit && (
              <Button
                variant="secondary"
                onClick={onEdit}
                styles="px-2 py-0.5 gap-2"
                buttonText="Edit"
                childSide="right"
              >
                <Icon icon={Edit03Icon} styles="size-3.5" />
              </Button>
            )}
            <Button
              styles="p-1"
              variant="danger"
              onClick={onDelete}
              disabled={isDeleting}
              isLoading={isDeleting}
              aria-label="Delete review"
            >
              <Icon icon={TrashIcon} styles="size-4" />
            </Button>
          </div>
        )}
      </div>

      {review.comment && (
        <div className="text-fg pb-2 text-sm whitespace-pre-wrap">
          {review.comment}
        </div>
      )}

      {isReported && (
        <span className="text-fg-muted text-xs">
          This review was reported and can’t be edited or deleted while our team
          looks into it.
        </span>
      )}
    </Card>
  );
}

function PreviousReviewCard({ serviceName, employeeName, onUpdate }) {
  return (
    <Card styles="p-4 flex flex-col gap-3">
      <div className="flex flex-col gap-1">
        <h4 className="text-fg text-base font-medium">
          You've reviewed this before
        </h4>
        <p className="text-fg-muted text-sm">
          Visited again? Update your review for{" "}
          <span className="font-medium">{serviceName}</span>
          with <span className="font-medium">{employeeName}</span>.
        </p>
      </div>
      <div className="pt-1">
        <Button
          variant="secondary"
          styles="w-fit text-sm px-4 py-2"
          buttonText="Update review"
          type="button"
          onClick={onUpdate}
        />
      </div>
    </Card>
  );
}

export function StarRating({
  value = 0,
  onChange,
  iconStyles = "size-5",
  styles = "gap-1",
}) {
  if (!onChange) {
    return (
      <div className={`flex items-center ${styles}`}>
        {[1, 2, 3, 4, 5].map((star) => (
          <Icon
            key={star}
            icon={StarIcon}
            styles={`${iconStyles} ${
              star <= value
                ? "fill-warning text-warning"
                : "text-border-strong fill-transparent"
            }`}
          />
        ))}
      </div>
    );
  }

  return (
    <div className={`flex items-center ${styles}`}>
      {[1, 2, 3, 4, 5].map((star) => (
        <button
          key={star}
          type="button"
          onClick={() => onChange(star)}
          className="cursor-pointer p-1"
          aria-label={`${star} star${star > 1 ? "s" : ""}`}
        >
          <Icon
            icon={StarIcon}
            styles={`${iconStyles} ${
              star <= value
                ? "fill-warning text-warning"
                : "text-border-strong fill-transparent"
            }`}
          />
        </button>
      ))}
    </div>
  );
}
