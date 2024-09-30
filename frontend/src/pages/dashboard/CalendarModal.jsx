import XIcon from "../../assets/icons/XIcon";

export default function CalendarModal({ eventInfo, isOpen, close }) {
  return (
    isOpen && (
      <div className="fixed inset-0 z-10 flex items-center justify-center bg-black bg-opacity-60">
        <div
          className="h-auto w-auto rounded-lg bg-layer_bg py-2 text-text_color shadow-lg sm:w-auto
            md:w-1/4"
        >
          <div className="flex flex-col gap-2 rounded-lg p-6">
            <div className="mb-6 flex items-center justify-between gap-10 border-b-2 border-text_color pb-1">
              <h1 className="text-xl font-bold">
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
              </h1>
              <XIcon
                styles="hover:bg-hvr_gray w-10 h-10 rounded-lg"
                onClick={close}
              />
            </div>
            <p className="text-lg">
              Client: {eventInfo.extendedProps?.name || ""}
            </p>
            <p>Number: </p>
            <p>Email: </p>
            <p className="text-lg">Type: {eventInfo.title || ""}</p>
            <p className="text-lg">Duration:</p>
          </div>
        </div>
      </div>
    )
  );
}
