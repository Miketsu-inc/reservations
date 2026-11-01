import {
  Button,
  ResponsiveDialog,
  ResponsiveDialogClose,
  ResponsiveDialogContent,
  Textarea,
} from "@reservations/components";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { StarRating } from "./BookingReview";

export default function ReviewModal({
  open,
  onOpenChange,
  booking,
  review,
  mode,
  initialRating,
}) {
  return (
    <ResponsiveDialog open={open} onOpenChange={onOpenChange}>
      <ResponsiveDialogContent
        styles="w-full lg:min-w-lg max-w-xl p-4 lg:p-6 flex flex-col gap-5"
      >
        <ReviewForm
          booking={booking}
          review={review}
          isUpdate={mode === "edit"}
          initialRating={initialRating}
          onClose={() => onOpenChange(false)}
        />
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}

function ReviewForm({ booking, review, isUpdate, initialRating, onClose }) {
  const queryClient = useQueryClient();
  const { showToast } = useToast();

  const [rating, setRating] = useState(
    isUpdate ? (review?.rating ?? 5) : initialRating
  );
  const [comment, setComment] = useState(
    isUpdate ? (review?.comment ?? "") : ""
  );
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(e) {
    e.preventDefault();
    if (!rating || isSubmitting) return;

    setIsSubmitting(true);
    try {
      const url = isUpdate
        ? `/api/v1/users/reviews/${review.id}`
        : `/api/v1/users/bookings/${booking.id}/review`;

      const method = isUpdate ? "PUT" : "POST";
      const payload = {
        rating,
        comment: comment.trim(),
        ...(isUpdate && { booking_id: booking.id }),
      };

      const response = await fetch(url, {
        method,
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        showToast({
          message: result.error.message,
          variant: "error",
        });
        return;
      }

      onClose();

      await queryClient.invalidateQueries({
        queryKey: ["user-booking-details", String(booking.id)],
      });
    } catch (err) {
      showToast({
        message: err.message,
        variant: "error",
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex w-full flex-col gap-4">
      <div className="flex flex-col gap-1">
        <h3 className="text-fg text-xl font-semibold">
          {isUpdate ? "Update your review" : "Leave a review"}
        </h3>
        {booking.service_name && booking.employee?.first_name && (
          <p className="text-fg-muted text-sm">{`${booking.service_name} with ${booking.employee.first_name}`}</p>
        )}
      </div>

      <div className="flex flex-col items-center gap-2 py-6">
        <StarRating
          value={rating}
          onChange={setRating}
          iconStyles="size-8 sm:size-9"
          styles="gap-2"
        />
      </div>

      <Textarea
        id="reviewComment"
        name="comment"
        rows={4}
        maxLength={300}
        value={comment}
        inputData={(data) => setComment(data.value)}
        placeholder="Share details of your experience (optional)"
        styles="p-3.5"
        required={false}
      />

      <div className="flex items-center justify-end gap-3 pt-2">
        <ResponsiveDialogClose asChild>
          <Button
            variant="secondary"
            styles="px-4 py-1.5 text-sm"
            buttonText="Cancel"
            type="button"
            disabled={isSubmitting}
          />
        </ResponsiveDialogClose>

        <Button
          variant="primary"
          styles="px-4 py-1.5 text-sm"
          buttonText={isUpdate ? "Save changes" : "Submit"}
          type="submit"
          isLoading={isSubmitting}
          disabled={!rating || isSubmitting}
        />
      </div>
    </form>
  );
}
