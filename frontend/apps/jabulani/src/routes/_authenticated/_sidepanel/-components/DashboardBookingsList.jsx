import { Loading, ServerError } from "@reservations/components";
import { invalidateLocalStorageAuth } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import BookingsList from "./BookingsList";

async function fetchDashboardBookings(merchantId, view) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/dashboard/bookings/${view}`,
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
  } else {
    return result.data;
  }
}

function dashboardBookingsQueryOptions(merchantId, view) {
  return queryOptions({
    queryKey: [merchantId, "dashboard-bookings", view],
    queryFn: () => fetchDashboardBookings(merchantId, view),
    staleTime: 15_000,
    gcTime: 5 * 60 * 1000,
  });
}

export default function DashboardBookingsList({
  merchantId,
  view,
  onAccept,
  route,
}) {
  const { data, isLoading, isError, error } = useQuery(
    dashboardBookingsQueryOptions(merchantId, view)
  );

  if (isError) {
    return <ServerError error={error.message} />;
  }

  if (isLoading) {
    return <Loading />;
  }

  return (
    <BookingsList bookings={data.bookings} onAccept={onAccept} route={route} />
  );
}
