import { Loading, ServerError } from "@reservations/components";
import {
  businessHoursQueryOptions,
  calculateStartEndTime,
  calendarTeamMembersQueryOptions,
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
import { calendarBookingsQueryOptions } from "../-components/calendarQueries";
import {
  canViewTeamCalendar,
  getCalendarRange,
  getCalendarRanges,
  resolveCalendarView,
} from "../-components/calendarViews";

const Calendar = lazy(() => import("../-components/Calendar"));

function validateDateString(dateStr) {
  return dateStringToLocalDate(dateStr) ? dateStr : undefined;
}

// Keeps the original array when nothing is dropped, so the loader can tell
// whether the search needs rewriting.
function validateEmployeeIds(employeeIds, teamMembers) {
  if (!Array.isArray(employeeIds)) return;

  const memberIds = new Set(teamMembers.map((member) => member.id));
  const validIds = employeeIds.filter((id) => memberIds.has(id));

  return validIds.length === employeeIds.length ? employeeIds : validIds;
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

    const canViewTeam = canViewTeamCalendar(role);
    const teamMembers = canViewTeam
      ? await queryClient.ensureQueryData(
          calendarTeamMembersQueryOptions(merchantId)
        )
      : [];
    const employees = canViewTeam
      ? validateEmployeeIds(search.employees, teamMembers)
      : undefined;

    const view = resolveCalendarView(
      getCalendarRanges(role, window.innerWidth < SCREEN_SM),
      getCalendarRange(search.view) ?? preferences?.calendar_view,
      (employees?.length ?? teamMembers.length) > 1
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

    if (
      view !== search.view ||
      start !== search.start ||
      end !== search.end ||
      employees !== search.employees
    ) {
      throw redirect({
        to: location.pathname,
        replace: true,
        search: { view, start, end, employees },
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
