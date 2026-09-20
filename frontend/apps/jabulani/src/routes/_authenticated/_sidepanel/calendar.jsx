import { Loading, ServerError } from "@reservations/components";
import {
  businessHoursQueryOptions,
  calculateStartEndTime,
  dateStringToLocalDate,
  isDurationValid,
  preferencesQueryOptions,
  SCREEN_SM,
} from "@reservations/lib";
import {
  createFileRoute,
  Outlet,
  redirect,
  useRouter,
} from "@tanstack/react-router";
import { lazy, Suspense } from "react";
import { calendarBookingsQueryOptions } from "./calendar/-components/calendarQueries";

const Calendar = lazy(() => import("./calendar/-components/Calendar"));

function validateDateString(dateStr) {
  return dateStringToLocalDate(dateStr) ? dateStr : undefined;
}

function mapCalendarView(view, mobileView) {
  const viewMapping = {
    month: "dayGridMonth",
    week: "timeGridWeek",
    day: "timeGridDay",
    list: "listWeek",
  };

  const preferredView = window.innerWidth < SCREEN_SM ? mobileView : view;

  return viewMapping[preferredView] ?? "timeGridWeek";
}

export const Route = createFileRoute("/_authenticated/_sidepanel/calendar")({
  component: CalendarLayout,
  loaderDeps: ({ search }) => search,
  loader: async ({
    deps: search,
    location,
    context: {
      queryClient,
      authContext: { merchantId, employeeId },
    },
  }) => {
    const preferences = await queryClient.ensureQueryData(
      preferencesQueryOptions(merchantId, employeeId)
    );

    const defaultView = preferences?.calendar_view
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
        to: location.pathname,
        replace: true,
        search: { view, start, end },
      });
    }

    await queryClient.ensureQueryData(
      calendarBookingsQueryOptions(merchantId, start, end)
    );
    await queryClient.ensureQueryData(businessHoursQueryOptions(merchantId));
  },
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});

function CalendarLayout() {
  const search = Route.useSearch();
  const router = useRouter();

  return (
    <Suspense fallback={<Loading />}>
      <Calendar router={router} route={Route} search={search} />
      <Outlet />
    </Suspense>
  );
}
