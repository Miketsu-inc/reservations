import {
  ArrowReloadHorizontalIcon,
  Clock01Icon,
  Location01Icon,
  User03Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  DEFAULT_SERVICE_COLOR,
  preferencesQueryOptions,
  timeStringFromDate,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";

const STATUS_STYLES = {
  booked:
    "bg-amber-600/15 text-amber-700 dark:bg-amber-500/15 dark:text-amber-400",
  confirmed:
    "bg-blue-600/15 text-blue-700 dark:bg-blue-500/15 dark:text-blue-400",
  completed:
    "bg-green-600/15 text-green-700 dark:bg-green-500/15 dark:text-green-400",
  cancelled: "bg-red-600/15 text-red-700 dark:bg-red-500/15 dark:text-red-400",
  "no-show":
    "bg-zinc-600/15 text-zinc-700 dark:bg-zinc-500/15 dark:text-zinc-400",
};

export default function BookingItem({ booking }) {
  const { merchantId, employeeId } = useAuth();
  const { data: preferences } = useQuery(
    preferencesQueryOptions(merchantId, employeeId)
  );
  const fromDate = new Date(booking.from_date);
  const toDate = new Date(booking.to_date);
  const employeeName = [booking.employee_first_name, booking.employee_last_name]
    .filter(Boolean)
    .join(" ");
  const statusLabel = booking.status.replace("-", " ");
  const displayPrice =
    booking.price_type === "free"
      ? "Free"
      : booking.price_type === "from"
        ? `From ${booking.price}`
        : booking.price;

  return (
    <article
      className="border-border_color bg-layer_bg overflow-hidden rounded-lg
        border shadow-sm transition-colors hover:border-gray-400
        dark:hover:border-zinc-700"
    >
      <div
        className="border-border_color flex items-start justify-between gap-3
          border-b p-3 sm:items-center sm:p-4"
      >
        <div className="flex min-w-0 items-center gap-3 sm:gap-4">
          <div
            className="bg-bg_color flex w-12 shrink-0 flex-col items-center
              rounded-lg px-2 py-1.5"
          >
            <span className="text-text_color text-lg leading-5 font-semibold">
              {fromDate.getDate()}
            </span>
            <span className="text-text_color/55 text-xs uppercase">
              {fromDate.toLocaleDateString([], { month: "short" })}
            </span>
            <span className="text-text_color/45 text-[10px]">
              {fromDate.getFullYear()}
            </span>
          </div>
          <div
            className="border-border_color hidden self-stretch border-r
              sm:block"
          />
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <span
                className="size-2.5 shrink-0 rounded-full"
                style={{
                  backgroundColor:
                    booking.service_color ?? DEFAULT_SERVICE_COLOR,
                }}
              />
              <h3 className="text-text_color truncate font-semibold sm:text-lg">
                {booking.service_name}
              </h3>
            </div>
            <div
              className="text-text_color/60 mt-1 flex items-center gap-1.5
                text-xs sm:text-sm"
            >
              <Icon icon={Clock01Icon} styles="size-3.5 shrink-0" />
              <span>
                {timeStringFromDate(fromDate, preferences?.time_format)} –{" "}
                {timeStringFromDate(toDate, preferences?.time_format)}
              </span>
            </div>
          </div>
        </div>
        <div className="flex shrink-0 flex-col items-end gap-2">
          <span
            className={`${STATUS_STYLES[booking.status]} rounded-full px-2.5
              py-1 text-xs font-medium capitalize`}
          >
            {statusLabel}
          </span>
          <span className="text-text_color hidden text-sm font-medium sm:block">
            {displayPrice}
          </span>
        </div>
      </div>

      <div
        className="text-text_color/60 flex flex-wrap items-center
          justify-between gap-2 px-3 py-2.5 text-xs sm:px-4 sm:text-sm"
      >
        <div className="flex min-w-0 flex-wrap items-center gap-x-4 gap-y-2">
          {booking.formatted_location && (
            <span className="flex min-w-0 items-center gap-1.5">
              <Icon icon={Location01Icon} styles="size-4 shrink-0" />
              <span className="truncate">{booking.formatted_location}</span>
            </span>
          )}
          {employeeName && (
            <span className="flex items-center gap-1.5">
              <Icon icon={User03Icon} styles="size-4 shrink-0" />
              {employeeName}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <span className="text-text_color font-medium sm:hidden">
            {displayPrice}
          </span>
          {booking.booking_type !== "appointment" && (
            <Pill icon={UserGroupIcon}>Group</Pill>
          )}
          {booking.is_recurring && (
            <Pill icon={ArrowReloadHorizontalIcon}>Repeating</Pill>
          )}
        </div>
      </div>
    </article>
  );
}

function Pill({ children, icon }) {
  return (
    <span
      className="border-border_color bg-bg_color flex items-center gap-1
        rounded-lg border px-2 py-1"
    >
      <Icon icon={icon} styles="size-3.5" />
      <span className="hidden sm:inline">{children}</span>
    </span>
  );
}
