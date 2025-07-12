import BackArrowIcon from "@icons/BackArrowIcon";
import MessageIcon from "@icons/MessageIcon";
import { useState } from "react";

export default function NotesSection({
  event,
  merchantNote,
  updateMerchantNote,
  disabled,
}) {
  const [areNotesHidden, setAreNotesHidden] = useState(true);

  return (
    <>
      <div className="bg-secondary rounded-lg border border-gray-400 px-3 py-2">
        <div className="flex w-full flex-row items-center">
          <div className="text-text_color flex items-center gap-2 text-sm">
            <MessageIcon styles="fill-text_color size-4" />
            <span>Notes</span>
          </div>
          <button
            type="button"
            className="flex grow cursor-pointer justify-end"
            onClick={() => setAreNotesHidden(!areNotesHidden)}
          >
            <BackArrowIcon
              styles={`${areNotesHidden ? "-rotate-90" : "rotate-90"} size-5 stroke-text_color
                transition-transform duration-200`}
            />
          </button>
        </div>
        <div
          className={`${areNotesHidden ? "max-h-0 opacity-0" : `${event.extendedProps.customer_note ? "max-h-52" : "max-h-32"} opacity-100`}
            overflow-hidden transition-all duration-300`}
        >
          <div className="pt-3">
            <div className="flex flex-col gap-3 text-sm">
              {event.extendedProps.customer_note && (
                <div className="flex flex-col gap-1">
                  <p>Customer's note</p>
                  <p
                    className="text-text_color bg-bg_color h-fit max-h-24 w-full overflow-auto rounded-lg
                      border border-gray-400 p-2 dark:[color-scheme:dark]"
                  >
                    {event.extendedProps.customer_note}
                  </p>
                </div>
              )}
              <div className="flex flex-col gap-1">
                <p>Your notes</p>
                <textarea
                  id="merchant_note"
                  name="merchant note"
                  value={merchantNote}
                  onChange={(e) => updateMerchantNote(e.target.value)}
                  disabled={disabled}
                  placeholder="Add notes here..."
                  className="bg-bg_color text-text_color max-h-20 min-h-20 w-full rounded-lg border
                    border-gray-400 p-2 text-sm outline-hidden focus:border-gray-900
                    dark:focus:border-gray-200"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
