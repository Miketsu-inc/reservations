import { Button, CloseButton, Modal, Textarea } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { useState } from "react";

const options = [
  {
    id: "this",
    label: "This booking",
    description: "Only this occurrence will be cancelled.",
  },
  {
    id: "future",
    label: "All future occurrences",
    description: "This and all future occurrences will be cancelled.",
  },
];

export default function CancelBookingModal({
  bookingId,
  onDeleted,
  isOpen,
  onClose,
  isRecurring,
}) {
  const { showToast } = useToast();
  const { merchantId } = useAuth();
  const [selected, setSelected] = useState("this");

  async function deleteBookingHandler(e) {
    e.preventDefault();

    const deletion_reason = e.target.elements.deletion_reason.value;
    const cancelFuture = selected === "future";

    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/bookings/${bookingId}`,
        {
          method: "DELETE",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify({
            cancel_future: cancelFuture,
            cancellation_reason: deletion_reason,
          }),
        }
      );

      if (!response.ok) {
        const result = await response.json();
        invalidateLocalStorageAuth(response.status);
        showToast({ message: result.error.message, variant: "error" });
      } else {
        showToast({
          message: "Booking deleted successfully",
          variant: "success",
        });

        onDeleted();
      }
    } catch (err) {
      showToast({ message: err.message, variant: "error" });
    }
  }

  return (
    <Modal
      styles="w-full sm:w-80"
      isOpen={isOpen}
      onClose={onClose}
      zindex={60}
      disableFocusTrap={true}
    >
      <form onSubmit={deleteBookingHandler} className="h-auto p-6 sm:w-130">
        <div className="flex flex-col gap-3">
          <div className="flex flex-col gap-5">
            <div className="flex items-center justify-between">
              <p className="text-lg font-medium">Cancel booking</p>
              <CloseButton onClick={onClose} />
            </div>
            {isRecurring && (
              <div className="flex flex-col gap-3 sm:flex-row">
                {options.map((opt) => {
                  const active = selected === opt.id;
                  return (
                    <button
                      key={opt.id}
                      type="button"
                      onClick={() => setSelected(opt.id)}
                      className={`flex flex-1 cursor-pointer flex-col gap-2
                      rounded-lg border px-4 py-4 text-left transition-all ${
                        active
                          ? "border-primary bg-primary/5 "
                          : `border-input_border_color hover:bg-gray-50
                            dark:hover:bg-gray-700/10`
                      }`}
                    >
                      <span className="flex flex-col gap-1">
                        <span
                          className={"text-text_color/90 text-sm font-semibold"}
                        >
                          {opt.label}
                        </span>
                        <span
                          className="text-xs leading-relaxed text-gray-500
                            dark:text-gray-400"
                        >
                          {opt.description}
                        </span>
                      </span>
                    </button>
                  );
                })}
              </div>
            )}
            <p className="py-2 text-sm">
              You can give a cancellation reason here, which will be included in
              the cancellation email sent to the customer.
            </p>
          </div>
          <Textarea
            styles="p-2 max-h-24 min-h-24 text-sm"
            id="deletion_reason"
            name="deletion reason"
            labelText="Deletion reason"
            required={false}
            placeholder="About this cutomer..."
          />
        </div>
        <div className="flex justify-end pt-4">
          <Button
            styles="px-4 py-1"
            buttonText="Cancel"
            variant="danger"
            type="submit"
          />
        </div>
      </form>
    </Modal>
  );
}
