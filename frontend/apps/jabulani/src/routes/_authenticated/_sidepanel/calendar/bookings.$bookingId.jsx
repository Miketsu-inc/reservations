import { ServerError } from "@reservations/components";
import { createFileRoute } from "@tanstack/react-router";
import { bookingQueryOptions } from "./-queries";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/calendar/bookings/$bookingId"
)({
  loader: ({
    params: { bookingId },
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => queryClient.ensureQueryData(bookingQueryOptions(merchantId, bookingId)),
  // The persistent calendar layout renders the panel from this route's param.
  // This leaf route owns loading and error handling without replacing the calendar.
  component: () => null,
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});
