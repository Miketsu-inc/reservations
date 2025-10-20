import { PopoverClose } from "@radix-ui/react-popover";
import { Button, Textarea } from "@reservations/components";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { useRef } from "react";

export default function DeleteBookingPopoverContent({ booking, onDeleted }) {
  const { showToast } = useToast();
  const closeButtonRef = useRef(null);

  async function deleteBookingHandler(e) {
    e.preventDefault();

    const deletion_reason = e.target.elements.deletion_reason.value;

    try {
      const response = await fetch(`/api/v1/bookings/${booking.id}`, {
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

        // close popover programmatically
        closeButtonRef.current.click();
      }
    } catch (err) {
      showToast({ message: err.message, variant: "error" });
    }
  }

  return (
    <form onSubmit={deleteBookingHandler} className="h-auto w-72 p-2 sm:w-80">
      <div className="flex flex-col gap-4">
        <p className="text-lg">Delete booking</p>
        <p className="py-2 text-sm">
          You can give a cancellation reason here, which will be included in the
          cancellation email sent to the customer.
        </p>
        <Textarea
          styles="p-2 max-h-20 min-h-20 text-sm"
          id="deletion_reason"
          name="deletion reason"
          labelText="Deletion reason"
          required={false}
          placeholder="About this cutomer..."
        />
      </div>
      <div className="flex justify-end pt-2">
        <PopoverClose asChild>
          <button
            ref={closeButtonRef}
            className="hidden"
            aria-hidden="true"
          ></button>
        </PopoverClose>
        <Button
          styles="p-2"
          buttonText="Delete"
          variant="danger"
          type="submit"
        />
      </div>
    </form>
  );
}
