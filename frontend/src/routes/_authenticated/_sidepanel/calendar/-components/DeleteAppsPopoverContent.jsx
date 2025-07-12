import Button from "@components/Button";
import { useToast } from "@lib/hooks";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { PopoverClose } from "@radix-ui/react-popover";
import { useRef } from "react";

export default function DeleteAppsPopoverContent({ event, onDeleted }) {
  const { showToast } = useToast();
  const closeButtonRef = useRef(null);

  async function deleteAppointmentHandler(e) {
    e.preventDefault();

    const deletion_reason = e.target.elements.deletion_reason.value;

    try {
      const response = await fetch(`/api/v1/appointments/${event.group_id}`, {
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
          message: "Appointment deleted successfully",
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
    <form
      onSubmit={deleteAppointmentHandler}
      className="h-auto w-72 p-2 sm:w-80"
    >
      <div className="flex flex-col gap-4">
        <p className="text-lg">Delete appointment</p>
        <p className="py-2 text-sm">
          You can give a cancellation reason here, which will be included in the
          cancellation email sent to the customer.
        </p>
        <div className="flex flex-col gap-1 px-1">
          <p className="">Deletion reason</p>
          <textarea
            id="deletion_reason"
            name="deletion reason"
            placeholder="Add your reason here..."
            className="bg-bg_color text-text_color max-h-20 min-h-20 w-full rounded-lg border
              border-gray-400 p-2 text-sm outline-hidden focus:border-gray-900
              dark:focus:border-gray-200"
          />
        </div>
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
