import Button from "@components/Button";
import Modal from "@components/Modal";
import CalendarIcon from "@icons/CalendarIcon";
import ClockIcon from "@icons/ClockIcon";
import MessageIcon from "@icons/MessageIcon";
import PersonIcon from "@icons/PersonIcon";
import PhoneIcon from "@icons/PhoneIcon";
import XIcon from "@icons/XIcon";
import { useState } from "react";

export default function CalendarModal({
  eventInfo,
  isOpen,
  onClose,
  setError,
}) {
  const [merchantComment, setMerchantComment] = useState(
    eventInfo.extendedProps.merchant_comment
  );
  const [hasUnsavedChanges, SetHasUnsavedChanges] = useState(false);

  async function saveButtonHandler() {
    if (merchantComment === eventInfo.extendedProps.merchant_comment) {
      return;
    }

    try {
      const response = await fetch("/api/v1/appointments/merchant-comment", {
        method: "PATCH",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          id: eventInfo.extendedProps.appointment_id,
          merchant_comment: merchantComment,
        }),
      });

      if (!response.ok) {
        const result = await response.json();
        setError(result.error.message);
      } else {
        eventInfo.setExtendedProp("merchant_comment", merchantComment);
        SetHasUnsavedChanges(false);
        setError("");
      }
    } catch (err) {
      setError(err.message);
    }
  }

  function merchantCommentChangeHandler(e) {
    setMerchantComment(e.target.value);

    if (e.target.value === eventInfo.extendedProps.merchant_comment) {
      SetHasUnsavedChanges(false);
    } else {
      SetHasUnsavedChanges(true);
    }
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <div className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-60 px-4">
        <div className="h-auto w-full rounded-lg bg-layer_bg text-text_color shadow-lg sm:w-1/2 lg:w-1/3">
          <div className="flex flex-col gap-2 rounded-lg p-3">
            <div className="flex items-center justify-between gap-10 border-b-2 border-gray-300 pb-1">
              <div className="flex items-center gap-3 pl-2 text-xl">
                <CalendarIcon styles="h-7 w-7 stroke-gray-700 dark:stroke-white" />
                {eventInfo.start.getFullYear()}-
                {String(eventInfo.start.getMonth() + 1).padStart(2, "0")}-
                {String(eventInfo.start.getDate()).padStart(2, "0")}
              </div>
              <XIcon
                styles="hover:bg-hvr_gray w-8 h-8 rounded-lg fill-gray-500 cursor-pointer
                  dark:fill-white"
                onClick={onClose}
              />
            </div>
            <div
              className="mb-2 mt-1 flex flex-col gap-3 rounded-lg bg-primary/70 p-3 font-semibold
                text-white"
            >
              {eventInfo.title}
              <div className="flex justify-between pr-2 text-sm text-gray-200">
                <div className="flex items-center gap-3">
                  <ClockIcon styles="fill-white" />
                  <span className="text-center">
                    {new Date(eventInfo.start).toLocaleTimeString([], {
                      hour: "2-digit",
                      minute: "2-digit",
                      hour12: false,
                      timeZone: "UTC",
                    })}
                    {" - "}
                    {new Date(eventInfo.end).toLocaleTimeString([], {
                      hour: "2-digit",
                      minute: "2-digit",
                      hour12: false,
                      timeZone: "UTC",
                    })}
                  </span>
                </div>
                <span>
                  {parseFloat(eventInfo.extendedProps.price).toLocaleString()}{" "}
                  HUF
                </span>
              </div>
            </div>
            <div
              className="mb-2 flex items-center justify-between rounded-lg border-l-4 border-blue-500
                bg-gray-200 p-3"
            >
              <div className="flex gap-3">
                <PersonIcon styles="fill-gray-600" />
                <span className="font-medium text-black">
                  {`${eventInfo.extendedProps.last_name} ${eventInfo.extendedProps.first_name}`}
                </span>
              </div>
              <div className="flex items-center gap-2">
                <PhoneIcon styles="fill-gray-600" />
                <span className="font-[0.6rem] text-gray-600">
                  {eventInfo.extendedProps.phone_number}
                </span>
              </div>
            </div>
            {/* Customer Note */}
            {eventInfo.extendedProps.user_comment && (
              <div className="mb-2 rounded-lg border border-gray-300 bg-gray-200 p-3">
                <div className="mb-1 flex items-center gap-2 text-sm text-gray-600">
                  <MessageIcon styles="fill-gray-600" />
                  <span>Customer Note</span>
                </div>
                <p
                  className="w-full rounded-lg border border-gray-300 bg-gray-100 px-3 py-2 text-sm
                    text-gray-700"
                >
                  {eventInfo.extendedProps.user_comment}
                </p>
              </div>
            )}

            {/* Merchant Note */}
            <div className="space-y-2 rounded-lg border border-blue-300 bg-blue-100 px-3 pt-3">
              <div className="flex items-center gap-2 text-sm text-blue-950">
                <MessageIcon styles="fill-blue-950" />
                <span>Your Notes</span>
              </div>
              <textarea
                name="merchant comment"
                value={merchantComment}
                onChange={merchantCommentChangeHandler}
                placeholder="Add notes about this appointment only you can see..."
                className="max-h-48 min-h-20 w-full rounded-lg border border-gray-300 bg-white p-2 text-sm
                  text-blue-950 outline-none focus:border-blue-600"
              />
            </div>
            <div className="flex items-center justify-end pt-2">
              <Button
                styles={`${
                  !hasUnsavedChanges
                    ? ""
                    : `hover:border-blue-600 hover:text-blue-600 dark:hover:border-blue-400
                      dark:hover:text-blue-400`
                  } bg-transparent text-sm text-blue-400 dark:text-blue-600 border border-blue-400
                  dark:border-blue-600 py-2 px-2 hover:bg-transparent min-w-16`}
                buttonText={hasUnsavedChanges ? "Save" : "Saved"}
                disabled={!hasUnsavedChanges}
                onClick={saveButtonHandler}
                type="button"
              />
            </div>
          </div>
        </div>
      </div>
    </Modal>
  );
}
