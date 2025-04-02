import Button from "@components/Button";
import CloseButton from "@components/CloseButton";
import DatePicker from "@components/DatePicker";
import Modal from "@components/Modal";
import CalendarIcon from "@icons/CalendarIcon";
import {
  addTimeToDate,
  combineDateTimeLocal,
  timeStringFromDate,
  toISOStringWithLocalTime,
} from "@lib/datetime";
import { useToast } from "@lib/hooks";
import { useEffect, useState } from "react";
import AppointmentInfoSection from "./AppointmentInfoSection";
import DeleteAppsModal from "./DeleteAppsModal";
import NotesSection from "./NotesSection";
import RecurSection from "./RecurSection";

export default function CalendarModal({
  eventInfo,
  isOpen,
  onClose,
  onDeleted,
  onEdit,
}) {
  const [recurData, setRecurData] = useState({
    isRecurring: false,
    frequency: "weekly",
    endDate: new Date(
      eventInfo.start.getFullYear(),
      eventInfo.start.getMonth() + 1,
      eventInfo.start.getDate()
    ),
  });
  const [merchantNote, setMerchantNote] = useState("");
  const [hasUnsavedChanges, SetHasUnsavedChanges] = useState(false);
  const [cancelModalOpen, setCancelModalOpen] = useState(false);
  const [eventDatetime, setEventDatetime] = useState({
    date: eventInfo.start,
    start_time: timeStringFromDate(eventInfo.start),
  });

  const { showToast } = useToast();

  // startEditable is false when the end date is higher than the current date
  const disabled = !eventInfo.startEditable;

  useEffect(() => {
    setMerchantNote(eventInfo.extendedProps.merchant_note);
    setEventDatetime({
      date: eventInfo.start,
      start_time: timeStringFromDate(eventInfo.start),
    });
    setRecurData({
      isRecurring: false,
      frequency: "weekly",
      endDate: new Date(
        eventInfo.start.getFullYear(),
        eventInfo.start.getMonth() + 1,
        eventInfo.start.getDate()
      ),
    });
  }, [eventInfo]);

  function updateRecurData(data) {
    setRecurData((prev) => ({ ...prev, ...data }));
  }

  function updateMerchantNote(note) {
    setMerchantNote(note);

    if (note === eventInfo.extendedProps.merchant_note) {
      SetHasUnsavedChanges(false);
    } else {
      SetHasUnsavedChanges(true);
    }
  }

  async function saveButtonHandler() {
    const start_date = combineDateTimeLocal(
      eventDatetime.date,
      eventDatetime.start_time
    );

    try {
      const response = await fetch(`/api/v1/appointments/${eventInfo.id}`, {
        method: "PATCH",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          id: eventInfo.extendedProps.appointment_id,
          merchant_note: merchantNote,
          from_date: toISOStringWithLocalTime(start_date),
          to_date: toISOStringWithLocalTime(
            addTimeToDate(
              start_date,
              0,
              eventInfo.extendedProps.service_duration
            )
          ),
        }),
      });

      if (!response.ok) {
        const result = await response.json();
        showToast({ message: result.error.message, variant: "error" });
      } else {
        eventInfo.setExtendedProp("merchant_note", merchantNote);
        SetHasUnsavedChanges(false);
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
      <DeleteAppsModal
        event={eventInfo}
        isOpen={cancelModalOpen}
        onClose={() => setCancelModalOpen(false)}
        onDeleted={() => {
          onDeleted();
          onClose();
        }}
      />
      <Modal
        suspendCloseOnClickOutside={cancelModalOpen}
        isOpen={isOpen}
        onClose={onClose}
      >
        <div className="h-auto w-full md:w-96">
          <div className="flex flex-col gap-2 p-3">
            <div className="flex items-start justify-between pb-1">
              <div className="flex items-center gap-3 text-xl">
                <CalendarIcon styles="size-7 stroke-gray-700 dark:stroke-white" />
                <p>Appointment</p>
              </div>
              <CloseButton onClick={onClose} />
            </div>
            <AppointmentInfoSection event={eventInfo} />
            <div className="px-1">
              <div className="flex flex-row items-center justify-between">
                <div className="flex flex-col gap-1">
                  <p>Date</p>
                  <DatePicker
                    styles="w-48"
                    defaultDate={eventInfo.start}
                    disabledBefore={new Date()}
                    disabled={disabled}
                    onSelect={(date) => {
                      setEventDatetime((prev) => ({ ...prev, date: date }));

                      if (date.getTime() !== eventInfo.start.getTime()) {
                        SetHasUnsavedChanges(true);
                      } else {
                        SetHasUnsavedChanges(false);
                      }
                    }}
                  />
                </div>
                <div className="flex flex-col gap-1">
                  <p>Start time</p>
                  <input
                    className="h-10 w-32 dark:[color-scheme:dark]"
                    type="time"
                    value={eventDatetime.start_time}
                    disabled={disabled}
                    onChange={(e) => {
                      setEventDatetime((prev) => ({
                        ...prev,
                        start_time: e.target.value,
                      }));

                      if (
                        e.target.value !== timeStringFromDate(eventInfo.start)
                      ) {
                        SetHasUnsavedChanges(true);
                      } else {
                        SetHasUnsavedChanges(false);
                      }
                    }}
                  />
                </div>
              </div>
            </div>
            <RecurSection
              event={eventInfo}
              recurData={recurData}
              updateRecurData={updateRecurData}
              disabled={disabled}
            />
            <NotesSection
              event={eventInfo}
              merchantNote={merchantNote}
              updateMerchantNote={updateMerchantNote}
              disabled={disabled}
            />
            <div className="flex items-center justify-end gap-2 pt-2">
              <Button
                styles="p-2"
                buttonText="Delete"
                variant="danger"
                type="button"
                onClick={() => setCancelModalOpen(true)}
              />
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
