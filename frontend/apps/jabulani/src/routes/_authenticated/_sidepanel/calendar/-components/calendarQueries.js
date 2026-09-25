import { invalidateLocalStorageAuth } from "@reservations/lib";
import { queryOptions } from "@tanstack/react-query";

async function fetchBookings(merchantId, start, end, timeZone) {
  const params = new URLSearchParams({
    start,
    end,
    time_zone: timeZone,
  });

  const response = await fetch(
    `/api/v1/merchants/${merchantId}/calendar/events?${params}`,
    { method: "GET" }
  );
  const result = await response.json();

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  }

  if (result.data !== null) {
    return result.data;
  }
}

export function calendarBookingsQueryOptions(merchantId, start, end) {
  const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  return queryOptions({
    queryKey: [merchantId, "events", start, end, timeZone],
    queryFn: () => fetchBookings(merchantId, start, end, timeZone),
  });
}

async function fetchBooking(merchantId, bookingId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/calendar/bookings/${bookingId}`,
    { method: "GET" }
  );
  const result = await response.json();

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  }

  return result.data;
}

export function calendarBookingQueryOptions(merchantId, bookingId) {
  return queryOptions({
    queryKey: [merchantId, "calendar-booking", bookingId],
    queryFn: () => fetchBooking(merchantId, bookingId),
  });
}
