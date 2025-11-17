import { CalendarIcon } from "@reservations/assets";
import {
  Button,
  CloseButton,
  DatePicker,
  Modal,
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@reservations/components";
import {
  addTimeToDate,
  combineDateTimeLocal,
  timeStringFromDate,
  useToast,
} from "@reservations/lib";
import { useState } from "react";
import BookingInfoSection from "./BookingInfoSection";
import DeleteBookingPopoverContent from "./DeleteBookingPopoverContent";
import NotesSection from "./NotesSection";
// import RecurSection from "./RecurSection";

export default function CalendarModal({
  bookingInfo,
  isOpen,
  onClose,
  onDeleted,
  onEdit,
}) {
  // const [recurData, setRecurData] = useState({
  //   isRecurring: false,
  //   frequency: "weekly",
  //   endDate: new Date(
  //     bookingInfo.start.getFullYear(),
  //     bookingInfo.start.getMonth() + 1,
  //     bookingInfo.start.getDate()
  //   ),
  // });
  const [merchantNote, setMerchantNote] = useState(
    bookingInfo.extendedProps.merchant_note
  );
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [isDeletePopoverOpen, setIsDeletePopoverOpen] = useState(false);
  const [isDatepickerOpen, setIsDatepickerOpen] = useState(false);
  const [bookingDatetime, setBookingDatetime] = useState({
    date: bookingInfo.start,
    start_time: timeStringFromDate(bookingInfo.start).split(" ")[0],
  });

  const { showToast } = useToast();

  // startEditable is false when the end date is higher than the current date
  const disabled = !bookingInfo.startEditable;

  // function updateRecurData(data) {
  //   setRecurData((prev) => ({ ...prev, ...data }));
  // }

  function updateMerchantNote(note) {
    setMerchantNote(note);

    if (note === bookingInfo.extendedProps.merchant_note) {
      setHasUnsavedChanges(false);
    } else {
      setHasUnsavedChanges(true);
    }
  }

  async function saveButtonHandler() {
    const start_date = combineDateTimeLocal(
      bookingDatetime.date,
      bookingDatetime.start_time
    );

    try {
      const response = await fetch(
        `/api/v1/bookings/${bookingInfo.extendedProps.id}`,
        {
          method: "PATCH",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify({
            id: bookingInfo.extendedProps.id,
            merchant_note: merchantNote,
            from_date: start_date.toISOString(),
            to_date: addTimeToDate(
              start_date,
              0,
              bookingInfo.extendedProps.service_duration
            ).toISOString(),
          }),
        }
      );

      if (!response.ok) {
        const result = await response.json();
        showToast({ message: result.error.message, variant: "error" });
      } else {
        bookingInfo.setExtendedProp("merchant_note", merchantNote);
        setHasUnsavedChanges(false);
        showToast({
          message: "Successfully updated the booking",
          variant: "success",
        });

        onEdit();
        onClose();
      }
    } catch (err) {
      showToast({ message: err.message, variant: "error" });
    }
  }

  return (
    <>
      <Modal
        styles="w-full sm:w-fit"
        isOpen={isOpen}
        onClose={onClose}
        disableFocusTrap={true}
        suspendCloseOnClickOutside={isDeletePopoverOpen || isDatepickerOpen}
      >
        <div className="h-auto w-full">
          <div className="flex flex-col gap-3 p-3">
            <div className="flex items-start justify-between pb-1">
              <div className="flex items-center gap-3 text-xl">
                <CalendarIcon styles="size-7 stroke-gray-700 dark:stroke-white" />
                <p>Booking</p>
              </div>
              <CloseButton onClick={onClose} />
            </div>
            <BookingInfoSection booking={bookingInfo} />
            <div className="px-1">
              <div className="flex flex-row items-center justify-between gap-4">
                <div className="flex flex-col gap-1">
                  <p>Date</p>
                  <DatePicker
                    styles="sm:w-40 w-36"
                    value={bookingInfo.start}
                    disabledBefore={new Date()}
                    disabled={disabled}
                    onSelect={(date) => {
                      setBookingDatetime((prev) => ({ ...prev, date: date }));

                      if (date.getTime() !== bookingInfo.start.getTime()) {
                        setHasUnsavedChanges(true);
                      } else {
                        setHasUnsavedChanges(false);
                      }
                    }}
                    onOpenChange={(open) => setIsDatepickerOpen(open)}
                  />
                </div>
                <div className="flex flex-col gap-1">
                  <p>Start time</p>
                  <input
                    className="h-10 w-32 dark:scheme-dark"
                    type="time"
                    value={bookingDatetime.start_time}
                    disabled={disabled}
                    onChange={(e) => {
                      setBookingDatetime((prev) => ({
                        ...prev,
                        start_time: e.target.value,
                      }));

                      if (
                        e.target.value !==
                        timeStringFromDate(bookingInfo.start).split(" ")[0]
                      ) {
                        setHasUnsavedChanges(true);
                      } else {
                        setHasUnsavedChanges(false);
                      }
                    }}
                  />
                </div>
              </div>
            </div>
            {/* <RecurSection
              booking={bookingInfo}
              recurData={recurData}
              updateRecurData={updateRecurData}
              disabled={disabled}
            /> */}
            <NotesSection
              booking={bookingInfo}
              merchantNote={merchantNote}
              updateMerchantNote={updateMerchantNote}
              disabled={disabled}
            />
            <div className="flex items-center justify-end gap-2 pt-2">
              <Popover
                open={isDeletePopoverOpen}
                onOpenChange={setIsDeletePopoverOpen}
              >
                <PopoverTrigger asChild>
                  <Button
                    styles="p-2"
                    buttonText="Delete"
                    variant="danger"
                    type="button"
                  />
                </PopoverTrigger>
                <PopoverContent side="top" styles="w-fit">
                  <DeleteBookingPopoverContent
                    booking={bookingInfo.extendedProps}
                    onDeleted={() => {
                      onDeleted();
                      onClose();
                    }}
                  />
                </PopoverContent>
              </Popover>
              <Button
                styles="p-2"
                buttonText="Cancel"
                variant="tertiary"
                type="button"
                onClick={onClose}
              />
              <Button
                styles="p-2"
                variant="primary"
                buttonText={hasUnsavedChanges && !disabled ? "Save" : "Saved"}
                disabled={!hasUnsavedChanges || disabled}
                onClick={saveButtonHandler}
                type="button"
              />
            </div>
          </div>
        </div>
      </Modal>
    </>
  );
}
