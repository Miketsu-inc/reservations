export const BOOKING_STATUS_OPTIONS = [
  { value: "booked", label: "Booked" },
  { value: "confirmed", label: "Confirmed" },
  { value: "completed", label: "Completed" },
  { value: "cancelled", label: "Cancelled" },
  { value: "no-show", label: "No-show" },
];

const BOOKING_STATUS_STYLES = {
  booked: {
    badge:
      "bg-amber-600/20 text-amber-600 dark:bg-amber-600/15 dark:text-amber-400",
    color: "text-amber-600 dark:text-amber-400",
  },
  confirmed: {
    badge:
      "bg-blue-600/20 text-blue-600 dark:bg-blue-500/15 dark:text-blue-400",
    color: "text-blue-600 dark:text-blue-400",
  },
  completed: {
    badge:
      "bg-green-600/20 text-green-600 dark:bg-green-500/15 dark:text-green-400",
    color: "text-green-600 dark:text-green-400",
  },
  cancelled: {
    badge: "bg-red-600/20 text-red-600 dark:bg-red-500/15 dark:text-red-400",
    color: "text-red-600 dark:text-red-400",
  },
  "no-show": {
    badge:
      "bg-gray-600/20 text-gray-600 dark:bg-gray-500/15 dark:text-gray-400",
    color: "text-gray-600 dark:text-gray-400",
  },
};

function getStatusStyles(status) {
  return BOOKING_STATUS_STYLES[status] ?? BOOKING_STATUS_STYLES["no-show"];
}

export function getBookingStatusStyles(status) {
  return getStatusStyles(status).badge;
}

export function getBookingStatusColorStyles(status) {
  return getStatusStyles(status).color;
}
