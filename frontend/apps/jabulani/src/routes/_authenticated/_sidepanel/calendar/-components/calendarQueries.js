import {
  dateStringToLocalDate,
  invalidateLocalStorageAuth,
} from "@reservations/lib";
import { queryOptions } from "@tanstack/react-query";

async function fetchBookings(merchantId, start, end) {
  const startTime = dateStringToLocalDate(start);
  const endTime = dateStringToLocalDate(end);

  if (!startTime || !endTime) {
    throw new Error("Invalid calendar date range");
  }

  const params = new URLSearchParams({
    start: startTime.toISOString(),
    end: endTime.toISOString(),
    start_date: start,
    end_date: end,
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
