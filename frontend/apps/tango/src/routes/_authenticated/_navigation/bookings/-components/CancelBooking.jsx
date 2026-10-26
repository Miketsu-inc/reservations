import {
  ArrowRight01Icon,
  CalendarRemove01Icon,
} from "@hugeicons/core-free-icons";
import { Card, DeleteModal, Icon } from "@reservations/components";
import {
  CANCEL_DEADLINE_OPTIONS,
  formatDuration,
  invalidateLocalStorageAuth,
  useToast,
} from "@reservations/lib";
import { useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

export default function CancelBooking({ booking, isUpcoming }) {
  const [isCancelModalOpen, setIsCancelModalOpen] = useState(false);
  const { showToast } = useToast();
  const queryClient = useQueryClient();

  const fromDate = new Date(booking.from_date);
  const now = new Date();

  const cancelDeadline = booking.cancel_deadline || 0;
  const cancelDeadlineTime = new Date(
    fromDate.getTime() - cancelDeadline * 60000
  );
  const canCancel = isUpcoming && now < cancelDeadlineTime;

  const cancelDeadlineLabel =
    CANCEL_DEADLINE_OPTIONS.find((opt) => opt.value === booking.cancel_deadline)
      ?.label ?? formatDuration(booking.cancel_deadline);

  async function handleCancel() {
    try {
      const response = await fetch(`/api/v1/users/bookings/${booking.id}`, {
        method: "DELETE",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
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

      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ["user-booking-details", String(booking.id)],
        }),
        queryClient.invalidateQueries({ queryKey: ["user-bookings"] }),
        queryClient.invalidateQueries({ queryKey: ["user-booking-counts"] }),
      ]);
    } catch (err) {
      showToast({
        message: err.message,
        variant: "error",
      });
    } finally {
      setIsCancelModalOpen(false);
    }
  }

  return (
    <div className="flex flex-col gap-2 pt-2">
      <h3 className="text-fg text-sm sm:text-base">Cancellation</h3>
      <Card styles="p-0! flex flex-col overflow-hidden">
        <div className="flex flex-col gap-1.5 p-4 sm:p-5">
          <p className="text-fg-muted text-sm leading-relaxed">
            {canCancel
              ? cancelDeadline > 0
                ? `Please cancel at least ${cancelDeadlineLabel} before your appointment.`
                : "Free cancellation anytime before your appointment begins."
              : "The cancellation deadline has passed. This booking can no longer be cancelled."}
          </p>
        </div>

        {canCancel && (
          <button
            type="button"
            onClick={() => setIsCancelModalOpen(true)}
            className="hover:bg-hover border-border group flex w-full
              cursor-pointer items-center justify-between border-t px-4 py-3
              text-left transition-colors sm:px-5 sm:py-3.5"
          >
            <div className="flex items-center gap-3">
              <Icon
                icon={CalendarRemove01Icon}
                styles="size-5.5 text-fg-danger"
              />

              <span className="text-fg-danger text-sm">Cancel booking</span>
            </div>
            <Icon
              icon={ArrowRight01Icon}
              styles="size-4 text-fg-muted group-hover:text-fg
                transition-colors"
            />
          </button>
        )}
      </Card>

      <DeleteModal
        open={isCancelModalOpen}
        onOpenChange={setIsCancelModalOpen}
        onDelete={handleCancel}
      />
    </div>
  );
}
