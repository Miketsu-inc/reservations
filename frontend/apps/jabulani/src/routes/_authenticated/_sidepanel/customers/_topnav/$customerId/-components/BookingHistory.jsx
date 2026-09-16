import {
  Button,
  Loading,
  ServerError,
  Toggle,
  ToggleGroup,
} from "@reservations/components";
import { invalidateLocalStorageAuth } from "@reservations/lib";
import { useInfiniteQuery } from "@tanstack/react-query";
import BookingsList from "../../../../-components/BookingsList";

const PAGE_SIZE = 8;

async function fetchCustomerBookings(merchantId, customerId, status, cursor) {
  const params = new URLSearchParams({
    status,
    limit: PAGE_SIZE.toString(),
    cursor,
  });
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/customers/${customerId}/bookings?${params}`,
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

export default function BookingHistory({
  merchantId,
  customerId,
  status,
  counts,
  onStatusChange,
  route,
}) {
  const {
    data,
    error,
    fetchNextPage,
    hasNextPage,
    isError,
    isFetchingNextPage,
    isLoading,
  } = useInfiniteQuery({
    queryKey: [merchantId, "customer-bookings", customerId, status],
    queryFn: ({ pageParam }) =>
      fetchCustomerBookings(merchantId, customerId, status, pageParam),
    initialPageParam: "",
    getNextPageParam: (lastPage) =>
      lastPage.has_next_page ? lastPage.next_cursor : undefined,
  });

  const bookings = data?.pages.flatMap((page) => page.bookings) ?? [];

  return (
    <div>
      <p className="text-2xl">Booking history</p>
      <div className="py-8 text-sm">
        <ToggleGroup
          multiple={false}
          value={status}
          onValueChange={onStatusChange}
        >
          <Toggle value="upcoming" badgeText={counts?.times_upcoming}>
            Upcoming
          </Toggle>
          <Toggle value="completed" badgeText={counts?.times_completed}>
            Completed
          </Toggle>
          <Toggle value="cancelled" badgeText={counts?.times_cancelled_by_user}>
            Cancelled
          </Toggle>
        </ToggleGroup>
      </div>

      {isError ? (
        <ServerError error={error?.message} />
      ) : isLoading ? (
        <Loading />
      ) : (
        <BookingsList
          bookings={bookings}
          route={route}
          showCustomer={false}
          emptyTitle={`No ${status} bookings`}
          emptyMessage={
            status === "upcoming"
              ? "New bookings for this customer will appear here."
              : `This customer does not have any ${status} bookings yet.`
          }
        />
      )}

      {hasNextPage && (
        <Button
          styles="mt-4 w-full px-4 py-2"
          buttonText="Load more"
          isLoading={isFetchingNextPage}
          onClick={() => fetchNextPage()}
          type="button"
        />
      )}
    </div>
  );
}
