import {
  Calendar02Icon,
  Clock01Icon,
  Location01Icon,
  Note01Icon,
  Tick02Icon,
  User03Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import { Avatar, Icon } from "@reservations/components";
import { getBookingStatusStyles, useAuth } from "@reservations/jabulani/lib";
import {
  DEFAULT_SERVICE_COLOR,
  preferencesQueryOptions,
  timeStringFromDate,
  useWindowSize,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";

export default function BookingsList({
  bookings,
  visibleCount = bookings.length,
  onAccept,
  route,
  showCustomer = true,
  emptyTitle = "No bookings yet",
  emptyMessage = "When customers schedule bookings, they will appear here.",
}) {
  const visibleBookings = bookings.slice(0, visibleCount);

  return (
    <div className="h-full">
      {visibleBookings.length > 0 ? (
        <div className="space-y-4">
          {visibleBookings.map((booking, index) => (
            <BookingCard
              key={`${booking.id}-${index}`}
              booking={booking}
              onAccept={onAccept}
              route={route}
              showCustomer={showCustomer}
            />
          ))}
        </div>
      ) : (
        <div
          className="bg-layer_bg flex flex-col items-center justify-center
            rounded-lg p-4 text-center shadow-sm"
        >
          <div className="mb-3 rounded-full bg-gray-300 p-3 dark:bg-gray-700">
            <Icon
              icon={Calendar02Icon}
              styles="size-8 text-gray-500 dark:text-gray-400"
            />
          </div>
          <p className="mb-1">{emptyTitle}</p>
          <p className="mb-3 text-sm text-gray-500 dark:text-gray-400">
            {emptyMessage}
          </p>
        </div>
      )}
    </div>
  );
}

function monthNameFromDate(date) {
  return date.toLocaleDateString([], { month: "short" });
}

function BookingCard({ booking, route, onAccept, showCustomer }) {
  const { isWindowSmall } = useWindowSize();

  const fromDate = new Date(booking.from_date);
  const toDate = new Date(booking.to_date);

  const isGroupBooking = booking.booking_type !== "appointment";
  const status = isGroupBooking
    ? (booking.participant_status ?? booking.status)
    : (booking.booking_status ?? booking.status);
  const isNotConfirmed = status === "booked";

  const isWalkIn =
    booking.customer_first_name === null && booking.customer_last_name === null;
  const employeeName = [booking.employee_first_name, booking.employee_last_name]
    .filter(Boolean)
    .join(" ");

  const { merchantId, employeeId } = useAuth();
  const { data: preferences } = useQuery(
    preferencesQueryOptions(merchantId, employeeId)
  );

  return (
    <div
      className="border-border_color bg-layer_bg dark:hover:bg-layer_bg/80
        rounded-lg border shadow-sm hover:border-gray-400 hover:bg-gray-100
        dark:hover:border-zinc-700"
    >
      <div
        className="border-border_color flex flex-row justify-between border-b
          px-4 py-4"
      >
        <Link
          className="flex min-w-0 flex-1 cursor-pointer flex-row gap-3 sm:gap-4
            lg:cursor-default"
          from={route.fullPath}
          to="/calendar/bookings/$bookingId"
          params={{ bookingId: String(booking.id) }}
          disabled={!isWindowSmall}
        >
          <div className="flex flex-col items-center justify-center">
            <p className="text-lg font-semibold">{fromDate.getDate()}</p>
            <p className="text-text_color/60 text-sm">
              {monthNameFromDate(fromDate)}
            </p>
          </div>
          <div className="border-border_color border-r" />
          <div className="flex min-w-0 flex-col items-start justify-center">
            <div className="flex min-w-0 flex-row items-center gap-2">
              <div
                className="rounded-lg p-1"
                style={{
                  backgroundColor:
                    booking.service_color ?? DEFAULT_SERVICE_COLOR,
                }}
              />
              <p className="truncate text-base sm:text-lg">
                {booking.service_name}
              </p>
              <div
                className={`${getBookingStatusStyles(status)} w-fit shrink-0
                  rounded-full px-2 py-1 text-xs sm:text-sm`}
              >
                <p>{status}</p>
              </div>
            </div>
            <div
              className="text-text_color/60 flex flex-row items-center gap-2
                text-sm"
            >
              <Icon icon={Clock01Icon} styles="size-3.5" />
              <p>
                {`${timeStringFromDate(fromDate, preferences?.time_format)} - ${timeStringFromDate(toDate, preferences?.time_format)}`}
              </p>
            </div>
          </div>
        </Link>
        <div className="flex flex-row items-center">
          {isNotConfirmed && onAccept && (
            <button
              className="lg:hover:bg-hvr_gray h-full cursor-pointer rounded-lg
                px-2 lg:h-fit lg:py-2"
              onClick={() => onAccept(booking)}
            >
              <Icon icon={Tick02Icon} styles="size-6 text-text_color" />
            </button>
          )}
          <Link
            className="lg:hover:bg-hvr_gray hidden h-full items-center
              rounded-lg px-2.5 lg:flex lg:h-fit lg:py-2.5"
            from={route.fullPath}
            to="/calendar/bookings/$bookingId"
            params={{ bookingId: String(booking.id) }}
          >
            <Icon icon={Calendar02Icon} styles="size-5 text-text_color" />
          </Link>
        </div>
      </div>
      <div className="flex flex-row items-center justify-between px-3 py-2">
        <div className="flex min-w-0 flex-row items-center gap-2">
          {showCustomer ? (
            <>
              {!isWalkIn && (
                <Avatar
                  styles="size-8! text-xs!"
                  initials={`${booking.customer_first_name?.[0] ?? ""}${booking.customer_last_name?.[0] ?? ""}`}
                />
              )}
              <p className="truncate text-sm">
                {isWalkIn
                  ? "Walk-in"
                  : `${booking.customer_first_name} ${booking.customer_last_name}`}
              </p>
            </>
          ) : employeeName ? (
            <>
              <Icon icon={User03Icon} styles="size-4 shrink-0" />
              <p className="truncate text-sm">{employeeName}</p>
            </>
          ) : (
            <>
              <Icon icon={Location01Icon} styles="size-4 shrink-0" />
              <p className="truncate text-sm">{booking.formatted_location}</p>
            </>
          )}
        </div>
        <div className="flex flex-row items-center gap-2">
          {booking?.customer_note && (
            <Pill>
              <Icon icon={Note01Icon} styles="size-4" />
              <p>Note</p>
            </Pill>
          )}
          {isGroupBooking && (
            <Pill>
              <Icon icon={UserGroupIcon} styles="size-4" />
              <p>
                Group
                {booking.current_participants !== undefined &&
                  booking.max_participants !== undefined &&
                  ` ${booking.current_participants}/${booking.max_participants}`}
              </p>
            </Pill>
          )}
        </div>
      </div>
    </div>
  );
}

function Pill({ children }) {
  return (
    <div
      className="border-border_color bg-bg_color text-text_color/60 flex w-fit
        flex-row items-center gap-1 rounded-lg border px-2 py-1 text-sm"
    >
      {children}
    </div>
  );
}
