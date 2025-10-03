import Button from "@components/Button";
import Modal from "@components/Modal";
import CalendarIcon from "@icons/CalendarIcon";
import { formatToDateString, timeStringFromDate } from "@lib/datetime";
import { useToast } from "@lib/hooks";
import { preferencesQueryOptions } from "@lib/queries";
import { useQuery } from "@tanstack/react-query";

function dateFormatter(date, time_format) {
  if (!date) return "";
  return formatToDateString(date) + " " + timeStringFromDate(date, time_format);
}

export default function DragConfirmationModal({
  bookingData,
  isOpen,
  onClose,
  onMoved,
}) {
  const { showToast } = useToast();
  const { data: preferences } = useQuery(preferencesQueryOptions());

  async function submitButtonHandler(e) {
    e.preventDefault();

    try {
      const response = await fetch(
        `/api/v1/bookings/${bookingData.booking.extendedProps.id}`,
        {
          method: "PATCH",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify({
            merchant_note: bookingData.booking.extendedProps.merchant_note,
            from_date: bookingData.booking.start.toISOString(),
            to_date: bookingData.booking.end.toISOString(),
          }),
        }
      );
      if (!response.ok) {
        const result = await response.json();
        showToast({ message: result.error.message, variant: "error" });
        bookingData.revert();
      } else {
        showToast({
          message: "Successfully updated the booking",
          variant: "success",
        });

        onMoved();
      }
    } catch (err) {
      showToast({ message: err.message, variant: "error" });
      bookingData.revert();
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <form onSubmit={submitButtonHandler} className="h-auto w-full p-2">
        <div className="flex justify-center py-2">
          <p className="text-xl">Are you sure?</p>
        </div>
        <div className="just flex flex-col items-center py-4">
          <p className="max-w-72 pb-5 text-center">
            You are about to modify the date of{" "}
            <span className="font-semibold">
              {bookingData.booking.extendedProps?.first_name}'s
            </span>{" "}
            booking
          </p>
          <div className="flex flex-col items-center gap-2">
            <div className="flex flex-row items-center gap-2">
              <CalendarIcon styles="size-5" />
              <p>
                {dateFormatter(
                  bookingData.old_booking.start,
                  preferences?.time_format
                )}
              </p>
            </div>
            <p>to</p>
            <div className="flex flex-row items-center gap-2">
              <CalendarIcon styles="size-5" />
              <p>
                {dateFormatter(
                  bookingData.booking.start,
                  preferences?.time_format
                )}
              </p>
            </div>
          </div>
        </div>
        <div className="flex flex-row items-center justify-end gap-3">
          <Button
            variant="tertiary"
            name="cancel"
            styles="py-2 px-3"
            buttonText="Cancel"
            type="button"
            onClick={onClose}
          />
          <Button
            variant="primary"
            name="confirm"
            styles="py-2 px-3"
            buttonText="Confirm"
            type="submit"
          />
        </div>
      </form>
    </Modal>
  );
}
