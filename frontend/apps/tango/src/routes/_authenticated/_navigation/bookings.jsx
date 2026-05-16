import { SearchInput, Toggle, ToggleGroup } from "@reservations/components";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import BookingList from "./-components/BookingList";

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
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = Route.useNavigate();
  const { status } = Route.useSearch();
  const [searchText, setSearchText] = useState("");

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
            <Toggle value="upcoming">Upcoming</Toggle>
            <Toggle value="completed">Completed</Toggle>
            <Toggle value="cancelled">Cancelled</Toggle>
          </ToggleGroup>
        </div>
        <BookingList statusFilter={status} searchText={searchText} />
      </div>
    </div>
  );
}
