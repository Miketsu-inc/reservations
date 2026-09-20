import {
  dateStringToLocalDate,
  invalidateLocalStorageAuth,
} from "@reservations/lib";
import { queryOptions } from "@tanstack/react-query";

async function fetchBookings(merchantId, start, end) {
  const startDate = dateStringToLocalDate(start);
  const endDate = dateStringToLocalDate(end);

  if (!startDate || !endDate) {
    throw new Error("Invalid calendar date range");
  }

  const response = await fetch(
    `/api/v1/merchants/${merchantId}/calendar/events?start=${startDate.toISOString()}&end=${endDate.toISOString()}`,
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
  return queryOptions({
    queryKey: [merchantId, "events", start, end],
    queryFn: () => fetchBookings(merchantId, start, end),
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
