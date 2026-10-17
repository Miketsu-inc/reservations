import {
  Calendar02Icon,
  Clock01Icon,
  Location01Icon,
  Note01Icon,
  Tick02Icon,
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
  onAccept,
  route,
  showCustomer = true,
  emptyTitle = "No bookings yet",
  emptyMessage = "When customers schedule bookings, they will appear here.",
}) {
  const { isWindowSmall } = useWindowSize();
  const { merchantId, employeeId } = useAuth();
  const { data: preferences } = useQuery(
    preferencesQueryOptions(merchantId, employeeId)
  );

  return (
    <div>
      {bookings.length > 0 ? (
        <div className="space-y-3">
          {bookings.map((booking, index) => (
            <BookingCard
              key={`${booking.id}-${index}`}
              booking={booking}
              isWindowSmall={isWindowSmall}
              onAccept={onAccept}
              route={route}
              showCustomer={showCustomer}
              timeFormat={preferences?.time_format}
            />
          ))}
        </div>
      ) : (
        <div
          className="border-border flex flex-col items-center justify-center
            rounded-xl border border-dashed p-6 text-center"
        >
          <div className="bg-bg-sunken mb-3 rounded-full p-2.5">
            <Icon icon={Calendar02Icon} styles="size-6 text-fg-muted" />
          </div>
          <p className="mb-1 text-sm font-medium">{emptyTitle}</p>
          <p className="text-fg-muted text-sm">{emptyMessage}</p>
        </div>
      )}
    </div>
  );
}

function monthNameFromDate(date) {
  return date.toLocaleDateString([], { month: "short" });
}

function BookingCard({
  booking,
  isWindowSmall,
  route,
  onAccept,
  showCustomer,
  timeFormat,
}) {
  const fromDate = new Date(booking.from_date);
  const toDate = new Date(booking.to_date);

  const isGroupBooking = booking.booking_type !== "appointment";
  const status = isGroupBooking
    ? (booking.participant_status ?? booking.status)
    : (booking.booking_status ?? booking.status);
  const isNotConfirmed = status === "booked";

  const customerName = `${booking.customer_first_name} ${booking.customer_last_name}`;
  const isWalkIn = !customerName;
  const employeeName = `${booking.employee_first_name} ${booking.employee_last_name}`;

  return (
    <div
      className="border-border bg-bg-raised shadow-card
        hover:border-border-strong rounded-xl border"
    >
      <div
        className="border-border-subtle flex flex-row justify-between border-b
          px-4 py-3"
      >
        <Link
          className="flex min-w-0 flex-1 cursor-pointer flex-row gap-3 sm:gap-4
            lg:cursor-default"
          from={route.fullPath}
          to="/calendar/bookings/$bookingId"
          params={{ bookingId: String(booking.id) }}
          disabled={!isWindowSmall}
        >
          <div
            className="flex shrink-0 flex-col items-center justify-center
              rounded-lg pt-1"
          >
            <p className="leading-tight font-semibold">{fromDate.getDate()}</p>
            <p className="text-fg-muted text-sm">
              {monthNameFromDate(fromDate)}
            </p>
          </div>
          <div className="flex min-w-0 flex-col items-start justify-center">
            <div className="flex min-w-0 flex-row items-center gap-2 pl-0.5">
              <div
                className="size-2 shrink-0 rounded-full"
                style={{
                  backgroundColor:
                    booking.service_color ?? DEFAULT_SERVICE_COLOR,
                }}
              />
              <p className="truncate text-sm font-medium sm:text-base">
                {booking.service_name}
              </p>
              <div
                className={`${getBookingStatusStyles(status)} w-fit shrink-0
                  rounded-full px-2 py-0.5 text-xs font-medium capitalize`}
              >
                <p>{status}</p>
              </div>
            </div>
            <div
              className="text-fg-muted flex flex-row items-center gap-2 text-sm"
            >
              <Icon icon={Clock01Icon} styles="size-3.5" />
              <p>
                {`${timeStringFromDate(fromDate, timeFormat)} - ${timeStringFromDate(toDate, timeFormat)}`}
              </p>
            </div>
          </div>
        </Link>
        <div className="flex flex-row items-center">
          {isNotConfirmed && onAccept && (
            <button
              className="lg:hover:bg-hover text-fg-muted hover:text-fg h-full
                cursor-pointer rounded-lg px-2 lg:h-fit lg:py-2"
              onClick={() => onAccept(booking)}
              type="button"
            >
              <Icon icon={Tick02Icon} styles="size-5" />
            </button>
          )}
          <Link
            className="text-fg-muted lg:hover:bg-hover lg:hover:text-fg hidden
              h-full items-center rounded-lg px-2 lg:flex lg:h-fit lg:py-2"
            from={route.fullPath}
            to="/calendar/bookings/$bookingId"
            params={{ bookingId: String(booking.id) }}
          >
            <Icon icon={Calendar02Icon} styles="size-5" />
          </Link>
        </div>
      </div>
      <div className="flex flex-row items-center justify-between px-3 py-2">
        <div className="flex min-w-0 flex-row items-center gap-2">
          {showCustomer ? (
            <>
              {!isWalkIn && (
                <Avatar
                  styles="size-6! text-[10px]!"
                  initials={`${booking.customer_first_name?.[0]}${booking.customer_last_name?.[0]}`}
                />
              )}
              <p className="text-fg-muted truncate text-sm">
                {isWalkIn ? "Walk-in" : customerName}
              </p>
            </>
          ) : employeeName ? (
            <>
              <Avatar
                styles="size-6! text-[10px]!"
                initials={`${booking.employee_first_name?.[0]}${booking.employee_last_name?.[0]}`}
              />
              <p className="text-fg-muted truncate text-sm">{employeeName}</p>
            </>
          ) : (
            <>
              <Icon icon={Location01Icon} styles="size-4 shrink-0" />
              <p className="text-fg-muted truncate text-sm">
                {booking.formatted_location}
              </p>
            </>
          )}
        </div>
        <div className="flex flex-row items-center gap-2">
          {booking?.customer_note && (
            <Pill>
              <Icon icon={Note01Icon} styles="size-3.5" />
              <p>Note</p>
            </Pill>
          )}
          {isGroupBooking && (
            <Pill>
              <Icon icon={UserGroupIcon} styles="size-3.5" />
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
      className="bg-bg-sunken text-fg-muted flex w-fit flex-row items-center
        gap-1 rounded-md px-1.5 py-0.5 text-xs"
    >
      {children}
    </div>
  );
}
