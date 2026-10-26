import {
  ArrowLeft02Icon,
  ArrowReloadHorizontalIcon,
  ArrowUpRight01Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import {
  Avatar,
  Button,
  Card,
  Icon,
  Loading,
  ServerError,
} from "@reservations/components";
import {
  dayNameFromDate,
  formatDuration,
  getBookingStatusStyles,
  getDisplayPrice,
  invalidateLocalStorageAuth,
  timeStringFromDate,
} from "@reservations/lib";
import { queryOptions, useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import MapboxMap from "../../../m/$merchantName/-components/MapboxMap";
import CancelBooking from "./-components/CancelBooking";

// has to match the backend's small variant size
const THUMBNAIL_MAX_SIZE = 640;

function thumbnailWidth(image) {
  const scale = Math.min(
    1,
    THUMBNAIL_MAX_SIZE / Math.max(image.width, image.height)
  );
  return Math.round(image.width * scale);
}

async function fetchUserBookingDetails(bookingId) {
  const response = await fetch(`/api/v1/users/bookings/${bookingId}`, {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  }
  return result.data;
}

function userBookingDetailsQueryOptions(bookingId) {
  return queryOptions({
    queryKey: ["user-booking-details", bookingId],
    queryFn: () => fetchUserBookingDetails(bookingId),
  });
}

export const Route = createFileRoute(
  "/_authenticated/_navigation/bookings/$bookingId"
)({
  component: BookingDetailsPage,
  loader: async ({ params: { bookingId }, context: { queryClient } }) => {
    await queryClient.ensureQueryData(
      userBookingDetailsQueryOptions(bookingId)
    );
  },
  pendingComponent: Loading,
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});

function BookingDetailsPage() {
  const { bookingId } = Route.useParams();
  const { data: booking } = useSuspenseQuery(
    userBookingDetailsQueryOptions(bookingId)
  );

  const coverImage = booking.images?.[0];

  const fromDate = new Date(booking.from_date);
  const toDate = new Date(booking.to_date);

  const durationString = formatDuration(
    (toDate.getTime() - fromDate.getTime()) / 60000
  );

  const now = new Date();
  const isCancelled =
    booking.status === "cancelled" ||
    booking.participant_status === "cancelled";
  const isNoShow = booking.participant_status === "no-show";
  const isPast = now > toDate;

  const isCompleted =
    !isCancelled &&
    !isNoShow &&
    (booking.participant_status === "completed" ||
      booking.status === "completed" ||
      isPast);
  const isConfirmed = booking.participant_status === "confirmed";
  const isBooked = booking.participant_status === "booked";
  const isUpcoming = isBooked || isConfirmed;

  const displayStatus = isCancelled
    ? "cancelled"
    : isNoShow
      ? "no-show"
      : isCompleted
        ? "completed"
        : isConfirmed
          ? "confirmed"
          : "booked";

  return (
    <div className="flex justify-center">
      <div className="flex w-full max-w-xl flex-col gap-5">
        <div className="hidden pt-4 pb-1 lg:block">
          <Link
            to="/bookings"
            className="text-fg-muted hover:text-fg flex h-fit items-center
              gap-1.5"
          >
            <Icon icon={ArrowLeft02Icon} styles="size-4.5" />
            <span>Bookings</span>
          </Link>
        </div>

        <div
          className="relative -mx-4 -mt-4 h-52 w-[calc(100%+2rem)]
            overflow-hidden sm:mx-0 sm:mt-0 sm:h-44 sm:w-full sm:rounded-2xl"
        >
          {coverImage ? (
            <img
              className="size-full object-cover"
              src={coverImage.url}
              srcSet={`${coverImage.thumbnail_url} ${thumbnailWidth(coverImage)}w, ${coverImage.url} ${coverImage.width}w`}
              sizes="(min-width: 640px) 36rem, 100vw"
              alt={coverImage.alt_text}
            />
          ) : (
            <div
              className="bg-bg-sunken flex size-full items-center
                justify-center"
            >
              <span className="text-fg-muted text-4xl font-semibold sm:text-5xl">
                {booking.merchant_name?.slice(0, 2).toUpperCase()}
              </span>
            </div>
          )}
          <div
            className="pointer-events-none absolute inset-0 bg-linear-to-t
              from-black/25 via-black/10 to-transparent dark:from-black/55"
          />
          <span
            className="absolute bottom-3 left-4 text-2xl font-medium text-white
              drop-shadow-sm dark:text-white/80"
          >
            <span>{booking.merchant_name}</span>
          </span>
        </div>

        <div className="flex flex-col gap-5 pt-1">
          <div className="flex items-start justify-between gap-4">
            <div className="flex flex-col gap-1.5">
              <span className="text-fg-muted text-sm">{`${dayNameFromDate(fromDate)}, ${fromDate.getDate()} ${fromDate.toLocaleDateString([], { month: "long" })}`}</span>
              <div className="flex items-baseline gap-2.5">
                <span className="text-fg text-3xl font-medium">
                  {timeStringFromDate(fromDate)} - {timeStringFromDate(toDate)}
                </span>

                <span className="text-fg-muted text-sm">{durationString}</span>
              </div>
            </div>

            <div
              className={`${getBookingStatusStyles(displayStatus)} shrink-0
                rounded-lg px-3 py-1.5 text-xs font-medium capitalize`}
            >
              <span>{displayStatus}</span>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <a
              href={`https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(
                booking.formatted_location
              )}`}
              target="_blank"
              rel="noopener noreferrer"
              className="bg-bg-raised border-border hover:bg-hover text-fg flex
                items-center gap-2 rounded-lg border px-3 py-2 text-sm"
            >
              <span className="text-nowrap">Get Directions</span>
              <Icon icon={ArrowUpRight01Icon} styles="size-4 mt-0.5" />
            </a>

            {(isCompleted || isCancelled || isNoShow) && (
              <Link
                to="/m/$merchantName"
                params={{ merchantName: booking.merchant_url }}
                className="w-full"
              >
                <Button variant="primary" styles="w-fit text-sm px-3 py-2">
                  Book again
                </Button>
              </Link>
            )}
          </div>
        </div>

        <div className="flex flex-col gap-2 pt-2">
          <h3 className="text-fg text-sm sm:text-base">Details</h3>
          <Card styles="p-0! divide-border divide-y overflow-hidden">
            <div className="flex flex-col gap-1 p-4">
              <div className="flex items-start justify-between gap-4">
                <span className="text-fg line-clamp-1">
                  {booking.service_name}
                </span>
                <span className="text-fg shrink-0">
                  {getDisplayPrice(booking.price, booking.price_type)}
                </span>
              </div>

              <span className="text-fg-muted text-xs">{durationString}</span>
            </div>

            <div className="flex items-center justify-between gap-3 px-4 py-3">
              <div className="flex items-center gap-3">
                <Avatar
                  styles="size-7! text-[10px]!"
                  img={booking.employee_avatar_url}
                  initials={`${booking.employee_first_name?.[0] ?? ""}${booking.employee_last_name?.[0] ?? ""}`}
                />

                <span className="text-fg text-sm">{`${booking.employee_first_name ?? ""} ${booking.employee_last_name ?? ""}`}</span>
              </div>

              <div className="flex shrink-0 flex-wrap items-center gap-1.5">
                <Badge
                  show={booking.booking_type !== "appointment"}
                  icon={UserGroupIcon}
                >
                  Group
                </Badge>
                <Badge
                  show={booking.is_recurring}
                  icon={ArrowReloadHorizontalIcon}
                >
                  Repeating
                </Badge>
              </div>
            </div>
          </Card>
        </div>

        {!isCancelled && !isCompleted && (
          <CancelBooking booking={booking} isUpcoming={isUpcoming} />
        )}

        <div className="flex flex-col gap-2 pt-2">
          <h3 className="text-fg text-sm sm:text-base">Location</h3>
          {booking.formatted_location && (
            <p className="text-fg-muted">{booking.formatted_location}</p>
          )}

          <MapboxMap
            coordinates={[booking.longitude, booking.latitude]}
            minHeight={220}
            zoom={15}
            styles="w-full h-56"
          />
        </div>
      </div>
    </div>
  );
}

function Badge({ show = true, icon, children }) {
  if (!show) return null;

  return (
    <div
      className="bg-bg-sunken text-fg-muted flex items-center gap-1.5 rounded-lg
        px-3 py-1.5 text-xs font-medium"
    >
      {icon && <Icon icon={icon} styles="size-3.5" />}
      <span>{children}</span>
    </div>
  );
}
