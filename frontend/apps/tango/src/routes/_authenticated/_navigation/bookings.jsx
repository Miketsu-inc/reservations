import {
  Loading,
  SearchInput,
  ServerError,
  Toggle,
  ToggleGroup,
} from "@reservations/components";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import BookingList from "./-components/BookingList";

async function fetchBookingCounts() {
  const response = await fetch("/api/v1/users/bookings/counts", {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  }

  return result.data;
}

function bookingCountsQueryOptions() {
  return queryOptions({
    queryKey: ["user-booking-counts"],
    queryFn: fetchBookingCounts,
  });
}

export const Route = createFileRoute("/_authenticated/_navigation/bookings")({
  validateSearch: (search) => {
    let status;

    if (
      search.status === "upcoming" ||
      search.status === "completed" ||
      search.status === "cancelled"
    ) {
      status = search.status;
    } else {
      status = "upcoming";
    }

    return {
      status,
    };
  },
  loader: async ({ context: { queryClient } }) => {
    await queryClient.ensureQueryData(bookingCountsQueryOptions());
  },
  pendingComponent: Loading,
  errorComponent: ({ error }) => <ServerError error={error.message} />,
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = Route.useNavigate();
  const { status } = Route.useSearch();
  const [searchText, setSearchText] = useState("");
  const { data: bookingCounts } = useQuery(bookingCountsQueryOptions());

  function statusChangeHandler(s) {
    navigate({
      to: "/bookings",
      search: () => ({ status: s }),
      replace: true,
    });
  }

  return (
    <div className="flex justify-center">
      <div className="flex w-full max-w-xl flex-col justify-center">
        <div className="flex flex-row items-center justify-between pt-4">
          <p className="text-2xl">Bookings</p>
          <SearchInput
            styles="w-40! md:w-full!"
            searchText={searchText}
            onChange={setSearchText}
          />
        </div>
        <div className="py-8 text-sm">
          <ToggleGroup
            multiple={false}
            value={status}
            onValueChange={statusChangeHandler}
          >
            <Toggle value="upcoming" badgeText={bookingCounts?.upcoming}>
              Upcoming
            </Toggle>
            <Toggle value="completed" badgeText={bookingCounts?.completed}>
              Completed
            </Toggle>
            <Toggle value="cancelled" badgeText={bookingCounts?.cancelled}>
              Cancelled
            </Toggle>
          </ToggleGroup>
        </div>
        <BookingList statusFilter={status} searchText={searchText} />
      </div>
    </div>
  );
}
