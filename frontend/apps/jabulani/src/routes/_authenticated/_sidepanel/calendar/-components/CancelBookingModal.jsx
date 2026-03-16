import { Button, CloseButton, Modal, Textarea } from "@reservations/components";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";

export default function CancelBookingModal({
  booking,
  onDeleted,
  isOpen,
  onClose,
}) {
  const { showToast } = useToast();

  async function deleteBookingHandler(e) {
    e.preventDefault();

    const deletion_reason = e.target.elements.deletion_reason.value;

    try {
      const response = await fetch(`/api/v1/bookings/merchant/${booking.id}`, {
        method: "DELETE",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          cancellation_reason: deletion_reason,
        }),
      });

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
      <form onSubmit={deleteBookingHandler} className="h-auto p-4 sm:w-110">
        <div className="flex flex-col gap-3">
          <div className="flex flex-col gap-2">
            <div className="flex items-center justify-between">
              <p className="text-lg font-medium">Cancel booking</p>
              <CloseButton onClick={onClose} />
            </div>
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
