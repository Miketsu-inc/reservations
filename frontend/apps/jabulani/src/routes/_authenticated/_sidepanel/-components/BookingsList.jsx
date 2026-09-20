import {
  Calendar02Icon,
  Clock01Icon,
  Note01Icon,
  Tick02Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import { Avatar, Icon } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  DEFAULT_SERVICE_COLOR,
  preferencesQueryOptions,
  timeStringFromDate,
  useWindowSize,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";

export default function BookingsList({ bookings, onAccept, route }) {
  return (
    <div className="h-full">
      {bookings.length > 0 ? (
        <div className="space-y-4">
          {bookings.map((booking, index) => (
            <BookingCard
              key={`${booking.id}-${index}`}
              booking={booking}
              onAccept={onAccept}
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
            <Icon
              icon={Calendar02Icon}
              styles="size-8 text-gray-500 dark:text-gray-400"
            />
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

function monthNameFromDate(date) {
  return date.toLocaleDateString([], { month: "short" });
}

const STATUS_STYLES = {
  booked:
    "bg-amber-600/20 text-amber-600 dark:bg-amber-600/15 dark:text-amber-400",
  confirmed:
    "bg-blue-600/20 text-blue-600 dark:bg-blue-500/15 dark:text-blue-400",
  completed:
    "bg-green-600/20 text-green-600 dark:bg-green-500/15 dark:text-green-400",
  cancelled: "bg-red-600/20 text-red-600 dark:bg-red-500/15 dark:text-red-400",
  "no-show":
    "bg-gray-600/20 text-gray-600 dark:bg-gray-500/15 dark:text-gray-400",
};

function BookingCard({ booking, route, onAccept }) {
  const { isWindowSmall } = useWindowSize();

  const fromDate = new Date(booking.from_date);
  const toDate = new Date(booking.to_date);

  const isGroupBooking = booking.booking_type !== "appointment";
  const status = isGroupBooking
    ? booking.participant_status
    : booking.booking_status;
  const isNotConfirmed = status === "booked";

  const isWalkIn =
    booking.customer_first_name === null && booking.customer_last_name === null;

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
          className="flex flex-1 cursor-pointer flex-row gap-4
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
          <div className="flex flex-col items-start justify-center">
            <div className="flex flex-row items-center gap-2">
              <div
                className="rounded-lg p-1"
                style={{
                  backgroundColor:
                    booking.service_color ?? DEFAULT_SERVICE_COLOR,
                }}
              />
              <p className="text-lg">{booking.service_name}</p>
              <div
                className={`${STATUS_STYLES[status]} w-fit rounded-full px-2
                  py-1 text-sm`}
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
          {!isWalkIn && (
            <Avatar
              styles="size-8! text-xs!"
              initials={`${booking.customer_first_name[0]}${booking.customer_last_name[0]}`}
            />
          )}
          <p className="truncate text-sm">
            {isWalkIn
              ? "Walk-in"
              : `${booking.customer_first_name} ${booking.customer_last_name}`}
          </p>
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
                Group {booking.current_participants}/{booking.max_participants}
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
