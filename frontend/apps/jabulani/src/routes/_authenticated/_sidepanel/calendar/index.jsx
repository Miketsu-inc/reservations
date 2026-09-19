import { Loading, ServerError } from "@reservations/components";
import {
  businessHoursQueryOptions,
  calculateStartEndTime,
  dateStringToLocalDate,
  invalidateLocalStorageAuth,
  isDurationValid,
  preferencesQueryOptions,
  SCREEN_SM,
} from "@reservations/lib";
import { queryOptions } from "@tanstack/react-query";
import { createFileRoute, redirect, useRouter } from "@tanstack/react-router";
import { lazy, Suspense } from "react";

const Calendar = lazy(() => import("./-components/Calendar"));

function validateDateString(dateStr) {
  return dateStringToLocalDate(dateStr) ? dateStr : undefined;
}

async function fetchBookings(merchantId, start, end) {
  const startDate = dateStringToLocalDate(start);
  const endDate = dateStringToLocalDate(end);

  if (!startDate || !endDate) {
    throw new Error("Invalid calendar date range");
  }

  start = startDate.toISOString();
  end = endDate.toISOString();

  const response = await fetch(
    `/api/v1/merchants/${merchantId}/calendar/events?start=${start}&end=${end}`,
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

export function bookingsQueryOptions(merchantId, start, end) {
  return queryOptions({
    queryKey: [merchantId, "events", start, end],
    queryFn: () => fetchBookings(merchantId, start, end),
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
  loaderDeps: ({ search }) => search,
  loader: async ({
    deps: search,
    context: {
      queryClient,
      authContext: { merchantId, employeeId },
    },
  }) => {
    const preferences = await queryClient.ensureQueryData(
      preferencesQueryOptions(merchantId, employeeId)
    );

    let defaultView = preferences?.calendar_view
      ? mapCalendarView(
          preferences.calendar_view,
          preferences.calendar_view_mobile
        )
      : "timeGridWeek";

    const view = [
      "dayGridMonth",
      "timeGridWeek",
      "timeGridDay",
      "listWeek",
    ].includes(search.view)
      ? search.view
      : defaultView;

    let start = validateDateString(search.start);
    let end = validateDateString(search.end);

    if (!start || !end || !isDurationValid(view, start, end)) {
      const calculated = calculateStartEndTime(
        view,
        preferences?.first_day_of_week
      );

      start = calculated.start;
      end = calculated.end;
    }

    if (view !== search.view || start !== search.start || end !== search.end) {
      throw redirect({
        from: Route.fullPath,
        to: "/calendar",
        replace: true,
        search: {
          view: view,
          start: start,
          end: end,
        },
      });
    }

    await queryClient.ensureQueryData(
      bookingsQueryOptions(merchantId, start, end)
    );
    await queryClient.ensureQueryData(businessHoursQueryOptions(merchantId));
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
