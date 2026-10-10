import { BOOKING_STATUS_OPTIONS } from "@reservations/jabulani/lib";
import { invalidateLocalStorageAuth } from "@reservations/lib";
import {
  infiniteQueryOptions,
  keepPreviousData,
  queryOptions,
} from "@tanstack/react-query";

const PAGE_SIZE = 8;

export const ALL_BOOKING_STATUSES = BOOKING_STATUS_OPTIONS.map(
  ({ value }) => value
);

async function fetchCustomerResource(merchantId, customerId, resource = "") {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/customers/${customerId}${resource}`,
    {
      method: "GET",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    }
  );

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  }

  return result.data;
}

export function customerProfileQueryOptions(merchantId, customerId) {
  return queryOptions({
    queryKey: [merchantId, "customer-profile", customerId],
    queryFn: () => fetchCustomerResource(merchantId, customerId),
  });
}

export function customerStatsQueryOptions(merchantId, customerId) {
  return queryOptions({
    queryKey: [merchantId, "customer-stats", customerId],
    queryFn: () => fetchCustomerResource(merchantId, customerId, "/stats"),
  });
}

async function fetchCustomerBookings(
  merchantId,
  customerId,
  statuses,
  beforeDate,
  cursor
) {
  const params = new URLSearchParams({
    customer_id: customerId,
    status: statuses.join(","),
    limit: PAGE_SIZE.toString(),
    cursor,
  });
  if (beforeDate) {
    params.set("before", startOfNextDay(beforeDate).toISOString());
  }

  const response = await fetch(
    `/api/v1/merchants/${merchantId}/bookings?${params}`,
    {
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    }
  );

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  }

  return result.data;
}

export function customerBookingsQueryOptions(
  merchantId,
  customerId,
  statuses = ALL_BOOKING_STATUSES,
  beforeDate = null
) {
  return infiniteQueryOptions({
    queryKey: [
      merchantId,
      "customer-bookings",
      customerId,
      statuses,
      beforeDate?.toISOString(),
    ],
    queryFn: ({ pageParam }) =>
      fetchCustomerBookings(
        merchantId,
        customerId,
        statuses,
        beforeDate,
        pageParam
      ),
    initialPageParam: "",
    getNextPageParam: (lastPage) =>
      lastPage.has_next_page ? lastPage.next_cursor : undefined,
    placeholderData: keepPreviousData,
  });
}

function startOfNextDay(date) {
  const result = new Date(date);
  result.setHours(0, 0, 0, 0);
  result.setDate(result.getDate() + 1);
  return result;
}
