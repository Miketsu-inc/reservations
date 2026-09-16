import { Calendar02Icon } from "@hugeicons/core-free-icons";
import {
  Button,
  Card,
  Icon,
  Loading,
  ServerError,
  Toggle,
  ToggleGroup,
} from "@reservations/components";
import { invalidateLocalStorageAuth } from "@reservations/lib";
import { useInfiniteQuery } from "@tanstack/react-query";
import BookingItem from "./BookingItem";

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
    <Card styles="p-0! overflow-hidden">
      <div className="border-border_color border-b px-4 py-4 sm:px-6">
        <div
          className="flex flex-col gap-4 sm:flex-row sm:items-center
            sm:justify-between"
        >
          <div>
            <h2 className="text-text_color text-lg font-semibold">
              Booking history
            </h2>
            <p className="text-text_color/55 mt-0.5 text-sm">
              Browse this customer&apos;s visits by status
            </p>
          </div>
          <ToggleGroup
            styles="max-w-full text-sm"
            multiple={false}
            value={status}
            onValueChange={onStatusChange}
          >
            <Toggle value="upcoming" badgeText={counts?.times_upcoming ?? 0}>
              Upcoming
            </Toggle>
            <Toggle value="completed" badgeText={counts?.times_completed ?? 0}>
              Completed
            </Toggle>
            <Toggle
              value="cancelled"
              badgeText={counts?.times_cancelled_by_user ?? 0}
            >
              Cancelled
            </Toggle>
          </ToggleGroup>
        </div>
      </div>

      <div className="p-3 sm:p-5">
        {isError ? (
          <ServerError error={error?.message} />
        ) : isLoading ? (
          <Loading />
        ) : bookings.length === 0 ? (
          <EmptyState status={status} />
        ) : (
          <div className="space-y-3">
            {bookings.map((booking) => (
              <BookingItem key={booking.id} booking={booking} />
            ))}
          </div>
        )}

        {hasNextPage && (
          <Button
            styles="mt-4 w-full px-4 py-2.5"
            buttonText="Load more bookings"
            isLoading={isFetchingNextPage}
            onClick={() => fetchNextPage()}
            type="button"
          />
        )}
      </div>
    </Card>
  );
}

function EmptyState({ status }) {
  return (
    <div
      className="flex flex-col items-center justify-center px-4 py-10
        text-center"
    >
      <span className="bg-hvr_gray mb-3 rounded-full p-3">
        <Icon icon={Calendar02Icon} styles="size-7 text-text_color/45" />
      </span>
      <p className="text-text_color font-medium">No {status} bookings</p>
      <p className="text-text_color/55 mt-1 max-w-sm text-sm">
        {status === "upcoming"
          ? "New bookings for this customer will appear here."
          : `This customer does not have any ${status} bookings yet.`}
      </p>
    </div>
  );
}
