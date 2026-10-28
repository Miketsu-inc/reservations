import { Loading, ServerError } from "@reservations/components";
import {
  businessHoursQueryOptions,
  calculateStartEndTime,
  calendarTeamMembersQueryOptions,
  dateStringToLocalDate,
  isDurationValid,
  preferencesQueryOptions,
  SCREEN_LG,
} from "@reservations/lib";
import {
  createFileRoute,
  Outlet,
  redirect,
  useRouter,
} from "@tanstack/react-router";
import { lazy, Suspense } from "react";
import { calendarBookingsQueryOptions } from "../-components/calendarQueries";
import {
  canViewTeamCalendar,
  getCalendarRanges,
  resolveCalendarRange,
} from "../-components/calendarViews";

const Calendar = lazy(() => import("../-components/Calendar"));

function validateDateString(dateStr) {
  return dateStringToLocalDate(dateStr) ? dateStr : undefined;
}

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/calendar/_topnav/_calendar"
)({
  component: CalendarLayout,
  loaderDeps: ({ search }) => search,
  loader: async ({
    deps: search,
    location,
    context: {
      queryClient,
      authContext: { merchantId, employeeId, role },
    },
  }) => {
    const preferences = await queryClient.ensureQueryData(
      preferencesQueryOptions(merchantId, employeeId)
    );

    // Load the team before the first render, so the team columns don't pop in.
    if (canViewTeamCalendar(role)) {
      await queryClient.ensureQueryData(
        calendarTeamMembersQueryOptions(merchantId)
      );
    }

    const view = resolveCalendarRange(
      getCalendarRanges(role, window.innerWidth < SCREEN_LG),
      search.view,
      preferences?.calendar_view
    );

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
