import Button from "@components/Button";
import CloseButton from "@components/CloseButton";
import DatePicker from "@components/DatePicker";
import Modal from "@components/Modal";
import { Popover, PopoverContent, PopoverTrigger } from "@components/Popover";
import CalendarIcon from "@icons/CalendarIcon";
import {
  addTimeToDate,
  combineDateTimeLocal,
  timeStringFromDate,
} from "@lib/datetime";
import { useToast } from "@lib/hooks";
import { useEffect, useState } from "react";
import AppointmentInfoSection from "./AppointmentInfoSection";
import DeleteAppsPopoverContent from "./DeleteAppsPopoverContent";
import NotesSection from "./NotesSection";
// import RecurSection from "./RecurSection";

export default function CalendarModal({
  eventInfo,
  isOpen,
  onClose,
  onDeleted,
  onEdit,
}) {
  // const [recurData, setRecurData] = useState({
  //   isRecurring: false,
  //   frequency: "weekly",
  //   endDate: new Date(
  //     eventInfo.start.getFullYear(),
  //     eventInfo.start.getMonth() + 1,
  //     eventInfo.start.getDate()
  //   ),
  // });
  const [merchantNote, setMerchantNote] = useState("");
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [isDeletePopoverOpen, setIsDeletePopoverOpen] = useState(false);
  const [isDatepickerOpen, setIsDatepickerOpen] = useState(false);
  const [eventDatetime, setEventDatetime] = useState({
    date: eventInfo.start,
    start_time: timeStringFromDate(eventInfo.start).split(" ")[0],
  });

  const { showToast } = useToast();

  // startEditable is false when the end date is higher than the current date
  const disabled = !eventInfo.startEditable;

  useEffect(() => {
    setMerchantNote(eventInfo.extendedProps.merchant_note);
    setEventDatetime({
      date: eventInfo.start,
      start_time: timeStringFromDate(eventInfo.start).split(" ")[0],
    });
    // setRecurData({
    //   isRecurring: false,
    //   frequency: "weekly",
    //   endDate: new Date(
    //     eventInfo.start.getFullYear(),
    //     eventInfo.start.getMonth() + 1,
    //     eventInfo.start.getDate()
    //   ),
    // });
  }, [eventInfo]);

  // function updateRecurData(data) {
  //   setRecurData((prev) => ({ ...prev, ...data }));
  // }

  function updateMerchantNote(note) {
    setMerchantNote(note);

    if (note === eventInfo.extendedProps.merchant_note) {
      setHasUnsavedChanges(false);
    } else {
      setHasUnsavedChanges(true);
    }
  }

  async function saveButtonHandler() {
    const start_date = combineDateTimeLocal(
      eventDatetime.date,
      eventDatetime.start_time
    );

    try {
      const response = await fetch(
        `/api/v1/appointments/${eventInfo.extendedProps.group_id}`,
        {
          method: "PATCH",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify({
            id: eventInfo.extendedProps.group_id,
            merchant_note: merchantNote,
            from_date: start_date.toISOString(),
            to_date: addTimeToDate(
              start_date,
              0,
              eventInfo.extendedProps.service_duration
            ).toISOString(),
          }),
        }
      );

      if (!response.ok) {
        const result = await response.json();
        showToast({ message: result.error.message, variant: "error" });
      } else {
        eventInfo.setExtendedProp("merchant_note", merchantNote);
        setHasUnsavedChanges(false);
        showToast({
          message: "Successfully updated the appointment",
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
                <p>Appointment</p>
              </div>
              <CloseButton onClick={onClose} />
            </div>
            <AppointmentInfoSection event={eventInfo} />
            <div className="px-1">
              <div className="flex flex-row items-center justify-between gap-4">
                <div className="flex flex-col gap-1">
                  <p>Date</p>
                  <DatePicker
                    styles="sm:w-40 w-36"
                    defaultDate={eventInfo.start}
                    disabledBefore={new Date()}
                    disabled={disabled}
                    onSelect={(date) => {
                      setEventDatetime((prev) => ({ ...prev, date: date }));

                      if (date.getTime() !== eventInfo.start.getTime()) {
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
                    value={eventDatetime.start_time}
                    disabled={disabled}
                    onChange={(e) => {
                      setEventDatetime((prev) => ({
                        ...prev,
                        start_time: e.target.value,
                      }));

                      if (
                        e.target.value !==
                        timeStringFromDate(eventInfo.start).split(" ")[0]
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
              event={eventInfo}
              recurData={recurData}
              updateRecurData={updateRecurData}
              disabled={disabled}
            /> */}
            <NotesSection
              event={eventInfo}
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
                  <DeleteAppsPopoverContent
                    event={eventInfo.extendedProps}
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
