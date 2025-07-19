import CalendarIcon from "@icons/CalendarIcon";
import ClockIcon from "@icons/ClockIcon";
import PersonIcon from "@icons/PersonIcon";
import PhoneIcon from "@icons/PhoneIcon";
import ServicesIcon from "@icons/ServicesIcon";
import { formatToDateString, timeStringFromDate } from "@lib/datetime";

export default function AppointmentInfoSection({ event }) {
  return (
    <div className="bg-primary dark:bg-primary/80 mb-2 flex flex-col gap-3 rounded-lg p-3 text-sm font-semibold text-white">
      <div className="flex flex-row justify-between">
        <div className="flex flex-row items-center gap-3">
          <PersonIcon styles="fill-white size-4" />
          <p>{event.title}</p>
        </div>
        <div className="flex items-center gap-2">
          <PhoneIcon styles="fill-white size-4" />
          <p>{event.extendedProps.phone_number}</p>
        </div>
      </div>
      <div className="flex justify-between">
        <div className="flex items-center gap-3">
          <ServicesIcon styles="size-4" />
          <p className="text-center">{event.extendedProps.service_name}</p>
        </div>
        {event.extendedProps.price && <p>{event.extendedProps.price}</p>}
      </div>
      <div className="flex flex-row justify-between">
        <div className="flex flex-row items-center gap-3">
          <CalendarIcon styles="size-4" />
          <p>{formatToDateString(event.start)}</p>
        </div>
        <div className="flex items-center gap-3">
          <ClockIcon styles="fill-white size-4" />
          <p className="text-center">
            {`${timeStringFromDate(event.start)} - ${timeStringFromDate(event.end)}`}
          </p>
        </div>
      </div>
    </div>
  );
}
