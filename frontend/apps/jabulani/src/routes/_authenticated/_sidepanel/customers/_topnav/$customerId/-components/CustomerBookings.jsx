import {
  Button,
  DatePicker,
  Loading,
  SearchInput,
  ServerError,
  Toggle,
  ToggleGroup,
} from "@reservations/components";
import { BOOKING_STATUS_OPTIONS, useAuth } from "@reservations/jabulani/lib";
import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import BookingsList from "../../../../-components/BookingsList";
import {
  ALL_BOOKING_STATUSES,
  customerBookingsQueryOptions,
  customerStatsQueryOptions,
} from "./queries";

const STATUS_COUNT_KEYS = {
  booked: "times_booked",
  confirmed: "times_confirmed",
  completed: "times_completed",
  cancelled: "times_cancelled",
  "no-show": "times_no_show",
};

export default function CustomerBookings({ customerId, route }) {
  const { merchantId } = useAuth();
  const [statuses, setStatuses] = useState(ALL_BOOKING_STATUSES);
  const [searchText, setSearchText] = useState("");
  const [beforeDate, setBeforeDate] = useState(null);
  const { data: counts } = useQuery(
    customerStatsQueryOptions(merchantId, customerId)
  );
  const {
    data,
    error,
    fetchNextPage,
    hasNextPage,
    isError,
    isFetchingNextPage,
    isLoading,
  } = useInfiniteQuery(
    customerBookingsQueryOptions(merchantId, customerId, statuses, beforeDate)
  );

  const allBookings = useMemo(
    () => data?.pages.flatMap((page) => page.bookings) ?? [],
    [data]
  );
  const bookings = useMemo(
    () => filterBookings(allBookings, searchText),
    [allBookings, searchText]
  );
  const statusFilters = BOOKING_STATUS_OPTIONS.map((status) => ({
    ...status,
    count: counts?.[STATUS_COUNT_KEYS[status.value]],
  })).filter(({ count }) => count == null || count > 0);
  const hasActiveFilters = Boolean(
    searchText.trim() ||
    beforeDate ||
    statuses.length !== ALL_BOOKING_STATUSES.length
  );

  return (
    <section>
      <div
        className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center
          sm:justify-between"
      >
        <div>
          <p className="text-xl">Bookings</p>
          <p className="text-text_color/60 mt-1 text-sm">
            Search and filter this customer's bookings
          </p>
        </div>
        <div className="flex w-full flex-col gap-2 sm:w-auto sm:flex-row">
          <SearchInput
            styles="w-full! sm:w-48!"
            searchText={searchText}
            onChange={setSearchText}
          />
          <div className="w-full sm:w-40">
            <DatePicker
              styles="w-full!"
              value={beforeDate}
              placeholderText="All dates"
              required={false}
              clearable
              closeOnSelect
              onSelect={(date) => setBeforeDate(date ?? null)}
            />
          </div>
        </div>
      </div>

      {statusFilters.length > 0 && (
        <div className="mb-6 text-sm">
          <ToggleGroup
            styles="-mx-1 gap-2 px-1 pb-1"
            multiple
            disableDeselect={false}
            value={statuses}
            onValueChange={(nextStatuses) => {
              if (nextStatuses.length > 0) {
                setStatuses(
                  ALL_BOOKING_STATUSES.filter((status) =>
                    nextStatuses.includes(status)
                  )
                );
              }
            }}
          >
            {statusFilters.map(({ value, label, count }) => (
              <Toggle key={value} value={value} badgeText={count}>
                {label}
              </Toggle>
            ))}
          </ToggleGroup>
        </div>
      )}

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
            hasActiveFilters ? "No matching bookings" : "No bookings yet"
          }
          emptyMessage={
            hasActiveFilters
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
    </section>
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
