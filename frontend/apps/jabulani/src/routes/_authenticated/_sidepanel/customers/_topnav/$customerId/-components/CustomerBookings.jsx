import {
  Button,
  DatePicker,
  Loading,
  SearchInput,
  ServerError,
  Toggle,
  ToggleGroup,
} from "@reservations/components";
import { invalidateLocalStorageAuth } from "@reservations/lib";
import { keepPreviousData, useInfiniteQuery } from "@tanstack/react-query";
import { useState } from "react";
import BookingsList from "../../../../-components/BookingsList";

const PAGE_SIZE = 8;

const ALL_STATUSES = [
  "booked",
  "confirmed",
  "completed",
  "cancelled",
  "no-show",
];

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

export default function CustomerBookings({
  merchantId,
  customerId,
  counts,
  route,
}) {
  const [statuses, setStatuses] = useState(ALL_STATUSES);
  const [searchText, setSearchText] = useState("");
  const [beforeDate, setBeforeDate] = useState(null);
  const {
    data,
    error,
    fetchNextPage,
    hasNextPage,
    isError,
    isFetchingNextPage,
    isLoading,
  } = useInfiniteQuery({
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

  const allBookings = data?.pages.flatMap((page) => page.bookings) ?? [];
  const bookings = filterBookings(allBookings, searchText);
  const cancelledCount = counts
    ? Math.max(0, counts.times_cancelled_by_user - counts.times_no_show)
    : null;

  return (
    <div>
      <div
        className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center
          sm:justify-between"
      >
        <div>
          <p className="text-2xl">Bookings</p>
          <p className="text-text_color/60 mt-1 text-sm">
            Search and filter this customer&apos;s bookings
          </p>
        </div>
        <div className="flex w-full flex-col gap-2 sm:w-auto sm:flex-row">
          <SearchInput
            styles="w-full! sm:w-48!"
            searchText={searchText}
            onChange={setSearchText}
          />
          <div className="flex items-center gap-2">
            <DatePicker
              styles="w-full! sm:w-40!"
              value={beforeDate}
              placeholderText="Jump to date"
              required={false}
              closeOnSelect
              onSelect={setBeforeDate}
            />
            {beforeDate && (
              <button
                className="text-text_color/60 hover:text-text_color px-1
                  text-sm"
                type="button"
                onClick={() => setBeforeDate(null)}
              >
                Clear
              </button>
            )}
          </div>
        </div>
      </div>

      <div className="mb-6 text-sm">
        <ToggleGroup
          styles="-mx-1 px-1 pb-1"
          multiple
          disableDeselect={false}
          value={statuses}
          onValueChange={(nextStatuses) => {
            if (nextStatuses.length > 0) setStatuses(nextStatuses);
          }}
        >
          <Toggle value="booked" badgeText={counts?.times_booked_status}>
            Booked
          </Toggle>
          <Toggle value="confirmed" badgeText={counts?.times_confirmed}>
            Confirmed
          </Toggle>
          <Toggle value="completed" badgeText={counts?.times_completed}>
            Completed
          </Toggle>
          <Toggle value="cancelled" badgeText={cancelledCount}>
            Cancelled
          </Toggle>
          <Toggle value="no-show" badgeText={counts?.times_no_show}>
            No-show
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
          emptyTitle={
            searchText || beforeDate || statuses.length !== ALL_STATUSES.length
              ? "No matching bookings"
              : "No bookings yet"
          }
          emptyMessage={
            searchText || beforeDate || statuses.length !== ALL_STATUSES.length
              ? "Try changing your search or filter."
              : "New bookings for this customer will appear here."
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

function filterBookings(bookings, searchText) {
  const search = searchText.trim().toLocaleLowerCase();
  if (!search) return bookings;

  return bookings.filter((booking) =>
    [
      booking.service_name,
      booking.formatted_location,
      booking.employee_first_name,
      booking.employee_last_name,
      [booking.employee_first_name, booking.employee_last_name]
        .filter(Boolean)
        .join(" "),
      booking.booking_status,
      booking.participant_status,
    ].some((value) => value?.toLocaleLowerCase().includes(search))
  );
}

function startOfNextDay(date) {
  const result = new Date(date);
  result.setHours(0, 0, 0, 0);
  result.setDate(result.getDate() + 1);
  return result;
}
