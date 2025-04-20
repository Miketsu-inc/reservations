import Loading from "@components/Loading";
import ServerError from "@components/ServerError";
import { SCREEN_SM } from "@lib/constants";
import { calculateStartEndTime, isDurationValid } from "@lib/datetime";
import { useToast } from "@lib/hooks";
import { getStoredPreferences, invalidateLocalSotrageAuth } from "@lib/lib";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { lazy, Suspense, useEffect, useState } from "react";

const Calendar = lazy(() => import("./-components/Calendar"));

async function fetchEvents(start, end) {
  start = new Date(start).toJSON();
  end = new Date(end).toJSON();

  const response = await fetch(
    `/api/v1/appointments/all?start=${start}&end=${end}`,
    {
      method: "GET",
    }
  );

  const result = await response.json();

  if (!response.ok) {
    invalidateLocalSotrageAuth(response.status);
    throw result.error;
  } else {
    if (result.data !== null) {
      return result.data;
    }
  }
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

    const preferences = getStoredPreferences();
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

    let start = search.start;
    let end = search.end;

    if (!start && !end && !isDurationValid(view, start, end)) {
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
  loader: async ({ deps: { start, end } }) => {
    const events = await fetchEvents(start, end);

    return {
      crumb: "Calendar",
      events: events,
    };
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function CalendarPage() {
  const search = Route.useSearch();
  const loaderData = Route.useLoaderData();
  const router = useRouter();
  const preferences = getStoredPreferences();
  const [businessHours, setBusinessHours] = useState();
  const { showToast } = useToast();

  useEffect(() => {
    async function fetchBusinessHours() {
      const response = await fetch(`/api/v1/merchants/business-hours`, {
        method: "GET",
        headers: {
          Accept: "application/json",
          "constent-type": "application/json",
        },
      });

      const result = await response.json();
      if (!response.ok) {
        showToast({
          message: "error fetching business hours",
          variant: "error",
        });
      } else {
        const transformedBusinessHours = Object.entries(result.data).map(
          ([day, times]) => ({
            daysOfWeek: [parseInt(day)],
            startTime: times.start_time.slice(0, 5),
            endTime: times.end_time.slice(0, 5),
          })
        );
        setBusinessHours(transformedBusinessHours);
      }
    }

    fetchBusinessHours();
  }, [showToast]);

  return (
    <div>
      <Suspense fallback={<Loading />}>
        <Calendar
          router={router}
          view={search.view}
          start={search.start}
          eventData={loaderData.events}
          preferences={preferences}
          businessHours={businessHours}
        />
      </Suspense>
    </div>
  );
}
