export const BOOKING_STATUS_OPTIONS = [
  { value: "booked", label: "Booked" },
  { value: "confirmed", label: "Confirmed" },
  { value: "completed", label: "Completed" },
  { value: "cancelled", label: "Cancelled" },
  { value: "no-show", label: "No-show" },
];

const BOOKING_STATUS_STYLES = {
  booked: {
    badge: "bg-warning-subtle text-fg-warning",
    color: "text-fg-warning",
  },
  confirmed: {
    badge: "bg-info-subtle text-fg-info",
    color: "text-fg-info",
  },
  completed: {
    badge: "bg-success-subtle text-fg-success",
    color: "text-fg-success",
  },
  cancelled: {
    badge: "bg-danger-subtle text-fg-danger",
    color: "text-fg-danger",
  },
  "no-show": {
    badge: "bg-hover text-fg-muted",
    color: "text-fg-muted",
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
