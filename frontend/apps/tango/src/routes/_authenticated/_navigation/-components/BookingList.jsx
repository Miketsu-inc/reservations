import { ClockIcon, CustomersIcon, RefreshIcon } from "@reservations/assets";
import { Avatar, Button, Loading, ServerError } from "@reservations/components";
import { timeStringFromDate } from "@reservations/lib";
import { useInfiniteQuery } from "@tanstack/react-query";

async function fetchBookings(status, limit, cursor) {
  const response = await fetch(
    `/api/v1/users/bookings?status=${status}&limit=${limit}&cursor=${cursor}`,
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
    throw result.error;
  } else {
    return result.data;
  }
}

export default function BookingList({ statusFilter, searchText }) {
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetching,
    isLoading,
    isError,
    error,
  } = useInfiniteQuery({
    queryKey: ["user-bookings", statusFilter],
    queryFn: ({ pageParam }) => fetchBookings(statusFilter, 10, pageParam),
    initialPageParam: undefined,
    getNextPageParam: (lastPage) =>
      lastPage.has_next_page ? lastPage.next_cursor : undefined,
  });

  const filteredBookings =
    data?.pages
      .flatMap((group) => group.bookings)
      .filter((booking) => {
        if (!searchText?.trim()) return true;

        const search = searchText.toLowerCase();
        return (
          booking.service_name?.toLowerCase().includes(search) ||
          booking.merchant_name?.toLowerCase().includes(search)
        );
      }) ?? [];

  if (isError) {
    return <ServerError error={error.message} />;
  }

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div className="space-y-4">
      {filteredBookings.length === 0 ? (
        <div className="flex justify-center pt-8">
          <p>
            {searchText?.trim()
              ? "No bookings match your search"
              : `You do not have ${statusFilter} bookings`}
          </p>
        </div>
      ) : (
        <>
          {filteredBookings.map((booking) => (
            <BookingCard key={booking.id} booking={booking} />
          ))}
        </>
      )}
      {hasNextPage && (
        <Button
          styles="px-4 py-2 w-full mt-2"
          buttonText="Load more"
          onClick={fetchNextPage}
          disabled={isFetching}
        />
      )}
    </div>
  );
}

function monthNameFromDate(date) {
  return date.toLocaleDateString([], { month: "short" });
}

function BookingCard({ booking }) {
  const fromDate = new Date(booking.from_date);
  const toDate = new Date(booking.to_date);

  const isGroupBooking = booking.booking_type !== "appointment";
  const isRecurring = booking.is_recurring;
  const isPending = booking.booking_status === "booked";

  return (
    <div
      className="border-border_color bg-layer_bg dark:hover:bg-layer_bg/80
        cursor-pointer rounded-lg border shadow-sm hover:border-gray-400
        hover:bg-gray-100 dark:hover:border-zinc-700"
    >
      <div
        className="border-border_color flex flex-row justify-between border-b
          px-4 py-4"
      >
        <div className="flex flex-row gap-4">
          <div className="flex flex-col items-center justify-center">
            <p className="text-lg font-semibold">{fromDate.getDate()}</p>
            <p className="text-text_color/60 text-sm">
              {monthNameFromDate(fromDate)}
            </p>
          </div>
          <div className="border-border_color border-r" />
          <div className="flex flex-col items-start justify-center">
            <p className="text-lg">{booking.service_name}</p>
            <div className="flex flex-row items-center gap-2">
              <ClockIcon styles="size-3 stroke-text_color/60" />
              <p className="text-text_color/60 text-sm">
                {timeStringFromDate(fromDate)} - {timeStringFromDate(toDate)}
              </p>
            </div>
          </div>
        </div>
        <div className="flex flex-col items-end gap-1">
          <p>{booking.price}</p>
          {isPending && (
            <p
              className="rounded-lg bg-yellow-300 px-2 py-1 text-xs
                text-yellow-800 dark:bg-yellow-900 dark:text-yellow-400"
            >
              pending
            </p>
          )}
        </div>
      </div>
      <div className="flex flex-row items-center justify-between px-4 py-2">
        <div className="flex flex-row items-center gap-2">
          <Avatar
            styles="size-8! text-xs!"
            initials={booking.merchant_name.slice(0, 2)}
          />
          <p className="text-sm">{booking.merchant_name}</p>
        </div>
        <div className="flex flex-row items-center gap-2">
          {isGroupBooking && (
            <Pill>
              <CustomersIcon styles="size-4" />
              <p>Group</p>
            </Pill>
          )}
          {isRecurring && (
            <Pill>
              <RefreshIcon styles="size-4" />
              <p>Repeating</p>
            </Pill>
          )}
        </div>
      </div>
    </div>
  );
}

function Pill({ children }) {
  return (
    <div
      className="border-border_color bg-bg_color text-text_color/60 flex w-fit
        flex-row items-center gap-1 rounded-lg border px-2 py-1 text-sm"
    >
      {children}
    </div>
  );
}
