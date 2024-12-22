import Button from "@components/Button";
import CalendarIcon from "@icons/CalendarIcon";
import ClockIcon from "@icons/ClockIcon";
import MessageIcon from "@icons/MessageIcon";
import PersonIcon from "@icons/PesonIcon";
import PhoneIcon from "@icons/PhoneIcon";
import { forwardRef, useEffect, useState } from "react";
import XIcon from "../../../../../assets/icons/XIcon";

export default forwardRef(function CalendarModal(
  { eventInfo, isOpen, close, error },
  ref
) {
  const [isSaved, setIsSaved] = useState(false);
  const [merchantComment, setMerchantComment] = useState("");

  useEffect(() => {
    if (eventInfo.extendedProps.merchant_comment !== undefined) {
      setMerchantComment(eventInfo.extendedProps.merchant_comment);
    }
  }, [eventInfo.extendedProps.merchant_comment]);

  useEffect(() => {
    if (
      isSaved &&
      merchantComment != eventInfo.extendedProps.merchant_comment
    ) {
      const sendRequest = async () => {
        try {
          const response = await fetch("/api/v1/appointments/modal", {
            method: "POST",
            headers: {
              Accept: "application/json",
              "content-type": "application/json",
            },
            body: JSON.stringify({
              id: eventInfo.id,
              merchant_comment: merchantComment,
            }),
          });

          if (!response.ok) {
            const result = await response.json();
            error(result.error.message);
          } else {
            eventInfo.setExtendedProp("merchant_comment", merchantComment);
            error("");
          }
        } catch (err) {
          error(err.message);
        } finally {
          setIsSaved(false);
        }
      };
      sendRequest();
    }
  }, [merchantComment, isSaved, error, eventInfo]);

  return (
    isOpen && (
      <div
        className="fixed inset-0 z-10 flex items-center justify-center bg-black bg-opacity-60 px-4
          transition-all"
      >
        <div
          ref={ref}
          className="h-auto w-full rounded-lg bg-layer_bg text-text_color shadow-lg sm:w-1/2 lg:w-1/3"
        >
          <div className="flex flex-col gap-2 rounded-lg p-3">
            <div className="flex items-center justify-between gap-10 border-b-2 border-gray-300 pb-1">
              <div className="flex items-center gap-3 pl-2 text-xl">
                <CalendarIcon styles="h-7 w-7 stroke-gray-700" />
                {eventInfo.start.getFullYear()}-
                {String(eventInfo.start.getMonth() + 1).padStart(2, "0")}-
                {String(eventInfo.start.getDate()).padStart(2, "0")}
              </div>
              <XIcon
                styles="hover:bg-hvr_gray w-8 h-8 rounded-lg fill-gray-500 cursor-pointer"
                onClick={close}
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
                    })}{" "}
                    -{" "}
                    {new Date(eventInfo.end).toLocaleTimeString([], {
                      hour: "2-digit",
                      minute: "2-digit",
                      hour12: false,
                      timeZone: "UTC",
                    })}
                  </span>
                </div>
                <span>{eventInfo.extendedProps.price} FT</span>
              </div>
            </div>
            <div
              className="mb-2 flex items-center justify-between rounded-lg border-l-4 border-blue-500
                bg-gray-200 p-3"
            >
              <div className="flex gap-3">
                <PersonIcon styles="fill-gray-600" />
                <span className="font-medium text-black">
                  {eventInfo.extendedProps.last_name}{" "}
                  {eventInfo.extendedProps.first_name}
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
                value={merchantComment}
                onChange={(e) => {
                  setMerchantComment(e.target.value);
                }}
                placeholder="Add notes about this appointment only you can see..."
                className="max-h-48 min-h-20 w-full rounded-lg border border-gray-300 bg-white px-2 py-2
                  text-sm text-blue-950 outline-none focus:border-blue-600"
              />
              <div className="flex items-center justify-end pb-2 pr-2">
                <Button
                  styles="bg-transparent text-sm text-blue-700 border border-blue-700 py-1 px-2
                    hover:border-blue-900 hover:text-blue-900"
                  buttonText="Save"
                  onClick={() => {
                    setIsSaved(true);
                  }}
                  type="button"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    )
  );
});
