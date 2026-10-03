const BOOKING_STATUS_STYLES = {
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

export function getBookingStatusStyles(status) {
  return BOOKING_STATUS_STYLES[status] ?? BOOKING_STATUS_STYLES["no-show"];
}
