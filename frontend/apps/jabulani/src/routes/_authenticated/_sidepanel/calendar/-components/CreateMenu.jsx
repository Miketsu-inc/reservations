import { CalendarPlusIcon, PlusIcon } from "@reservations/assets";
import {
  Button,
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@reservations/components";

export default function CreateMenu({
  onCreateBlockedTime,
  onCreateBooking,
  isFloating = false,
}) {
  return (
    <Popover>
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
          <PlusIcon
            styles={isFloating ? "size-7 text-white" : "size-4 mr-2 text-white"}
          />
        </Button>
      </PopoverTrigger>

      <PopoverContent align="end">
        <div
          className="*:hover:bg-hvr_gray flex flex-col items-start *:w-full
            *:rounded-lg *:p-2"
        >
          <button
            onClick={onCreateBooking}
            className="flex cursor-pointer items-center gap-2 text-left"
          >
            <CalendarPlusIcon styles="size-6" /> Booking
          </button>
          <button
            onClick={onCreateBlockedTime}
            className="cursor-pointer text-left"
          >
            Blocked Time
          </button>
        </div>
      </PopoverContent>
    </Popover>
  );
}
