import Card from "@components/Card";
import { Popover, PopoverContent, PopoverTrigger } from "@components/Popover";
import BackArrowIcon from "@icons/BackArrowIcon";
import CalendarIcon from "@icons/CalendarIcon";
import ClockIcon from "@icons/ClockIcon";
import TickIcon from "@icons/TickIcon";
import TrashBinIcon from "@icons/TrashBinIcon";
import { formatToDateString, timeStringFromDate } from "@lib/datetime";
import { preferencesQueryOptions } from "@lib/queries";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { useState } from "react";
import DeleteBookingPopoverContent from "../calendar/-components/DeleteBookingPopoverContent";

export default function BookingsList({
  bookings,
  visibleCount,
  onCancel,
  onAccept,
  route,
}) {
  const visibleBookings = bookings.slice(0, visibleCount);

  return (
    <div className="h-full">
      {visibleBookings.length > 0 ? (
        <div className="space-y-4">
          {visibleBookings.map((booking) => (
            <BookingCard
              key={booking.id}
              booking={booking}
              onCancel={(book) => onCancel(book)}
              onAccept={(book) => onAccept(book)}
              route={route}
            />
          ))}
        </div>
      ) : (
        <div
          className="bg-layer_bg flex flex-col items-center justify-center
            rounded-lg p-4 text-center shadow-sm"
        >
          <div className="mb-3 rounded-full bg-gray-300 p-3 dark:bg-gray-700">
            <CalendarIcon styles="size-8 stroke-gray-500 dark:stroke-gray-400" />
          </div>
          <p className="mb-1">No bookings yet</p>
          <p className="mb-3 text-sm text-gray-500 dark:text-gray-400">
            When customers schedule bookings, they will appear here.
          </p>
        </div>
      )}
    </div>
  );
}

function monthDateFormat(date) {
  return date.toLocaleDateString([], {
    weekday: "short",
    month: "short",
    day: "numeric",
  });
}

function BookingCard({ booking, route, onCancel, onAccept }) {
  const [showNote, setShowNote] = useState(false);
  const { data: preferences } = useQuery(preferencesQueryOptions());

  return (
    <Card styles="py-2">
      <div className="flex h-fit flex-row items-center">
        <div
          className="flex w-full flex-col lg:flex-row lg:items-center
            lg:justify-between lg:pr-3 xl:pr-6"
        >
          <div className="flex flex-col gap-2 py-1">
            <span className="dark:font-semibold">{`${booking.last_name} ${booking.first_name}`}</span>
            <div className="flex flex-row items-center gap-3">
              <span className="text-sm">{`${monthDateFormat(new Date(booking.from_date))}`}</span>
              <div className="flex flex-row items-center gap-2">
                <ClockIcon styles="size-3 fill-gray-500 dark:fill-gray-400" />
                <span className="text-sm">{`${timeStringFromDate(new Date(booking.from_date), preferences?.time_format)} - ${timeStringFromDate(new Date(booking.to_date), preferences?.time_format)}`}</span>
              </div>
            </div>
          </div>
          <span
            className="w-fit rounded-full px-2 py-1 text-xs"
            style={{
              backgroundColor: `${booking.service_color}20`,
              color: booking.service_color,
            }}
          >
            {booking.service_name}
          </span>
        </div>
        <div className="flex flex-row items-center gap-1">
          <Link
            from={route.fullPath}
            to="/calendar"
            params={{
              start: formatToDateString(new Date(booking.from_date)),
            }}
          >
            <CalendarIcon styles="size-5 stroke-text_color" />
          </Link>
          <Popover>
            <PopoverTrigger asChild>
              <button className="cursor-pointer ps-1">
                <TrashBinIcon styles="size-5" />
              </button>
            </PopoverTrigger>
            <PopoverContent side="bottom" align="end" styles="w-fit">
              <DeleteBookingPopoverContent
                booking={booking}
                onDeleted={onCancel}
              />
            </PopoverContent>
          </Popover>
          <button className="cursor-pointer" onClick={() => onAccept(booking)}>
            <TickIcon styles="size-6 stroke-text_color" />
          </button>
        </div>
      </div>
      {booking.customer_note && (
        <div className="pt-3 md:pt-2">
          <button
            className="flex cursor-pointer flex-row items-center gap-2
              text-gray-500 dark:text-gray-400"
            onClick={() => setShowNote(!showNote)}
          >
            <BackArrowIcon
              styles={`${showNote ? "rotate-90" : "-rotate-90"}
              transition-transform duration-300 size-3 stroke-gray-500
              dark:stroke-gray-400`}
            />
            <span className="text-xs">{showNote ? "Hide" : "View"} note</span>
          </button>
          <div
            className={`${showNote ? "mt-2 max-h-8 opacity-100 md:max-h-4" : "max-h-0 opacity-0"}
            overflow-hidden transition-all duration-300`}
          >
            <p className="text-xs text-gray-500 dark:text-gray-400">
              {booking.customer_note}
            </p>
          </div>
        </div>
      )}
    </Card>
  );
}
