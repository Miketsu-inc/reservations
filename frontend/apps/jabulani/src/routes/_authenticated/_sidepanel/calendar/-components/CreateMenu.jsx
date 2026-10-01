import { CalendarAdd01Icon, PlusSignIcon } from "@hugeicons/core-free-icons";
import {
  Button,
  Icon,
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@reservations/components";
import { useEffect, useState } from "react";

function CreateOptions({ onCreateBlockedTime, onCreateBooking }) {
  return (
    <div
      className="*:hover:bg-hvr_gray flex flex-col items-start *:w-full
        *:cursor-pointer *:rounded-lg *:p-2 *:outline-hidden"
    >
      <button onClick={onCreateBooking} className="flex items-center gap-2">
        <Icon icon={CalendarAdd01Icon} styles="size-6" /> Booking
      </button>
      <button className="text-left" onClick={onCreateBlockedTime}>
        Blocked Time
      </button>
    </div>
  );
}

export default function CreateMenu({
  onCreateBlockedTime,
  onCreateBooking,
  isFloating = false,
}) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="primary"
          styles={` ${
            isFloating
              ? "fixed bottom-10 right-5 z-10 p-3 shadow-xl"
              : "p-2 text-sm mr-5"
            } `}
          buttonText={isFloating ? "" : "Create"}
        >
          <Icon
            icon={PlusSignIcon}
            styles={isFloating ? "size-7 text-white" : "size-4 mr-2 text-white"}
          />
        </Button>
      </PopoverTrigger>

      <PopoverContent align="end">
        <CreateOptions
          onCreateBooking={() => {
            setIsOpen(false);
            onCreateBooking();
          }}
          onCreateBlockedTime={() => {
            setIsOpen(false);
            onCreateBlockedTime();
          }}
        />
      </PopoverContent>
    </Popover>
  );
}

export function CalendarCreateMenu({
  selection,
  onClose,
  onCreateBlockedTime,
  onCreateBooking,
}) {
  useEffect(() => {
    const listenerOptions = { capture: true, passive: true, once: true };

    window.addEventListener("scroll", onClose, listenerOptions);
    return () => window.removeEventListener("scroll", onClose, true);
  }, [onClose]);

  return (
    <Popover
      open
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <PopoverTrigger
        aria-hidden="true"
        className="pointer-events-none fixed size-px opacity-0"
        style={{ left: selection.x, top: selection.y }}
        tabIndex={-1}
      />
      <PopoverContent side="right" align="start">
        <CreateOptions
          onCreateBooking={() => onCreateBooking(selection.start)}
          onCreateBlockedTime={() => onCreateBlockedTime(selection.start)}
        />
      </PopoverContent>
    </Popover>
  );
}
