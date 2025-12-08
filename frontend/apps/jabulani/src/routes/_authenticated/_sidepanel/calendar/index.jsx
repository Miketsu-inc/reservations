import { Loading, ServerError } from "@reservations/components";
import {
  businessHoursQueryOptions,
  calculateStartEndTime,
  invalidateLocalStorageAuth,
  isDurationValid,
  SCREEN_SM,
} from "@reservations/lib";
import { queryOptions } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { lazy, Suspense } from "react";
import { globalQueryClient } from "../../../../main";

const Calendar = lazy(() => import("./-components/Calendar"));

function normalizeDateString(dateStr) {
  if (!dateStr) return undefined;

  try {
    // Parse the date string (handles various formats)
    const date = new Date(dateStr);

    // Check if valid date
    if (isNaN(date.getTime())) {
      return undefined;
    }

    // Return in YYYY-MM-DD format
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");

    return `${year}-${month}-${day}`;
  } catch (e) {
    return undefined;
  }
}

async function fetchBookings(start, end) {
  start = new Date(start).toJSON();
  end = new Date(end).toJSON();

  const response = await fetch(
    `/api/v1/bookings/calendar/events?start=${start}&end=${end}`,
    {
      method: "GET",
    }
  );

  const result = await response.json();

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    if (result.data !== null) {
      return result.data;
    }
  }
}

export function bookingsQueryOptions(start, end) {
  return queryOptions({
    queryKey: ["events", start, end],
    queryFn: () => fetchBookings(start, end),
  });
}

function mapCalendarView(view, mobile_view) {
  const viewMapping = {
    month: "dayGridMonth",
    week: "timeGridWeek",
    day: "timeGridDay",
    list: "listWeek",
  };

  if (window.innerWidth < SCREEN_SM) {
    return viewMapping[mobile_view];
  }
  return viewMapping[view];
}

export const Route = createFileRoute("/_authenticated/_sidepanel/calendar/")({
  component: CalendarPage,
  validateSearch: (search) => {
    let defaultView = "timeGridWeek";

    const preferences = globalQueryClient.getQueryData(["preferences"]);
    if (preferences?.calendar_view) {
      defaultView = mapCalendarView(
        preferences.calendar_view,
        preferences.calendar_view_mobile
      );
    }

    const view = [
      "dayGridMonth",
      "timeGridWeek",
      "timeGridDay",
      "listWeek",
    ].includes(search.view)
      ? search.view
      : defaultView;

    let start = normalizeDateString(search.start);
    let end = normalizeDateString(search.end);

    if (!start || !end || !isDurationValid(view, start, end)) {
      const calculated = calculateStartEndTime(
        view,
        preferences?.first_day_of_week
      );
      start = calculated.start;
      end = calculated.end;
    }

    return {
      view,
      start,
      end,
    };
  },
  loaderDeps: ({ search: { view, start, end } }) => ({
    view,
    start,
    end,
  }),
  loader: async ({ deps: { start, end }, context: { queryClient } }) => {
    await queryClient.ensureQueryData(bookingsQueryOptions(start, end));
    await queryClient.ensureQueryData(businessHoursQueryOptions());
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function CalendarPage() {
  const search = Route.useSearch();
  const router = useRouter();

  return (
    <Suspense fallback={<Loading />}>
      <Calendar router={router} route={Route} search={search} />
    </Suspense>
  );
}
