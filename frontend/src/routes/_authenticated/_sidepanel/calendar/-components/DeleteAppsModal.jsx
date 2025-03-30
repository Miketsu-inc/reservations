import Button from "@components/Button";
import Modal from "@components/Modal";
import { useToast } from "@lib/hooks";

export default function DeleteAppsModal({ event, isOpen, onClose, onDeleted }) {
  const { showToast } = useToast();

  async function deleteAppointmentHandler(e) {
    e.preventDefault();

    const deletion_reason = e.target.elements.deletion_reason.value;

    try {
      const response = await fetch(`/api/v1/appointments/${event.id}`, {
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
        showToast({ message: result.error.message, variant: "error" });
      } else {
        showToast({
          message: "Appointment deleted successfully",
          variant: "success",
        });

        onDeleted();
        onClose();
      }
    } catch (err) {
      showToast({ message: err.message, variant: "error" });
    }
  }

  return (
    <Modal zindex={50} isOpen={isOpen} onClose={onClose}>
      <form
        onSubmit={deleteAppointmentHandler}
        className="h-auto w-full p-2 md:w-80"
      >
        <div className="flex flex-col gap-4">
          <p className="text-lg">Delete appointment</p>
          <div className="flex flex-col gap-1 px-1">
            <p className="text-sm">Deletion reason</p>
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
        <div className="flex items-center justify-end gap-2 pt-2">
          <Button
            styles="p-2"
            buttonText="Cancel"
            variant="tertiary"
            type="button"
            onClick={onClose}
          />
          <Button
            styles="p-2"
            buttonText="Delete"
            variant="danger"
            type="submit"
          />
        </div>
      </form>
    </Modal>
  );
}
