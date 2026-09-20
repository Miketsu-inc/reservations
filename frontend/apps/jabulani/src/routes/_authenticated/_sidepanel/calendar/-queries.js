import {
  dateStringToLocalDate,
  invalidateLocalStorageAuth,
} from "@reservations/lib";
import { queryOptions } from "@tanstack/react-query";

function responseError(response, result, fallbackMessage) {
  const error = new Error(
    result?.error?.message ?? result?.error ?? fallbackMessage
  );
  error.status = response.status;

  return error;
}

function retryTransientError(failureCount, error) {
  return (error.status === undefined || error.status >= 500) && failureCount < 3;
}

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
    throw responseError(response, result, "Could not load calendar events");
  }

  if (result.data !== null) {
    return result.data;
  }
}

async function fetchBooking(merchantId, bookingId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/calendar/bookings/${bookingId}`,
    { method: "GET" }
  );
  const result = await response.json();

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw responseError(response, result, "Could not load booking");
  }

  return result.data;
}

export function bookingsQueryOptions(merchantId, start, end) {
  return queryOptions({
    queryKey: [merchantId, "events", start, end],
    queryFn: () => fetchBookings(merchantId, start, end),
    retry: retryTransientError,
  });
}

export function bookingQueryOptions(merchantId, bookingId) {
  return queryOptions({
    queryKey: [merchantId, "calendar-booking", bookingId],
    queryFn: () => fetchBooking(merchantId, bookingId),
    retry: retryTransientError,
  });
}
