import { ServerError } from "@reservations/components";
import {
  calculateStartEndTime,
  dateStringToLocalDate,
  preferencesQueryOptions,
} from "@reservations/lib";
import { createFileRoute, redirect } from "@tanstack/react-router";
import { calendarBookingQueryOptions } from "./-components/calendarQueries";

function isDateInRange(date, start, end) {
  const startDate = dateStringToLocalDate(start);
  const endDate = dateStringToLocalDate(end);

  return startDate && endDate && date >= startDate && date < endDate;
}

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/calendar/bookings/$bookingId"
)({
  loaderDeps: ({ search }) => search,
  loader: async ({
    deps: search,
    params: { bookingId },
    context: {
      queryClient,
      authContext: { merchantId, employeeId },
    },
  }) => {
    const booking = await queryClient.ensureQueryData(
      calendarBookingQueryOptions(merchantId, bookingId)
    );
    const bookingDate = new Date(booking.from_date);

    if (!isDateInRange(bookingDate, search.start, search.end)) {
      const preferences = await queryClient.ensureQueryData(
        preferencesQueryOptions(merchantId, employeeId)
      );
      const range = calculateStartEndTime(
        search.view,
        preferences?.first_day_of_week,
        bookingDate
      );

      if (range) {
        throw redirect({
          to: "/calendar/bookings/$bookingId",
          params: { bookingId },
          replace: true,
          search: { ...search, ...range },
        });
      }
    }

    return booking;
  },
  // The persistent calendar layout renders the panel from this route's param.
  // This leaf route owns loading and error handling without replacing the calendar.
  component: () => null,
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});
