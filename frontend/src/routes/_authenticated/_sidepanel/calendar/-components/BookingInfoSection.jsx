import CalendarIcon from "@icons/CalendarIcon";
import ClockIcon from "@icons/ClockIcon";
import PersonIcon from "@icons/PersonIcon";
import PhoneIcon from "@icons/PhoneIcon";
import ServicesIcon from "@icons/ServicesIcon";
import { formatToDateString, timeStringFromDate } from "@lib/datetime";
import { preferencesQueryOptions } from "@lib/queries";
import { useQuery } from "@tanstack/react-query";

export default function BookingInfoSection({ booking }) {
  const { data: preferences } = useQuery(preferencesQueryOptions());

  return (
    <div
      className="bg-primary dark:bg-primary/80 mb-2 flex flex-col gap-3
        rounded-lg p-3 text-sm font-semibold text-white"
    >
      <div className="flex flex-row justify-between">
        <div className="flex flex-row items-center gap-3">
          <PersonIcon styles="fill-white size-4" />
          <p>{booking.title}</p>
        </div>
        <div className="flex items-center gap-2">
          <PhoneIcon styles="fill-white size-4" />
          <p>{booking.extendedProps.phone_number}</p>
        </div>
      </div>
      <div className="flex justify-between">
        <div className="flex items-center gap-3">
          <ServicesIcon styles="size-4" />
          <p className="text-center">{booking.extendedProps.service_name}</p>
        </div>
        {booking.extendedProps.price && <p>{booking.extendedProps.price}</p>}
      </div>
      <div className="flex flex-row justify-between">
        <div className="flex flex-row items-center gap-3">
          <CalendarIcon styles="size-4" />
          <p>{formatToDateString(booking.start)}</p>
        </div>
        <div className="flex items-center gap-3">
          <ClockIcon styles="fill-white size-4" />
          <p className="text-center">
            {`${timeStringFromDate(booking.start, preferences?.time_format)} - ${timeStringFromDate(booking.end, preferences?.time_format)}`}
          </p>
        </div>
      </div>
    </div>
  );
}
