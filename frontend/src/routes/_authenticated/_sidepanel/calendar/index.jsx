import Loading from "@components/Loading";
import ServerError from "@components/ServerError";
import { calculateStartEndTime, isDurationValid } from "@lib/datetime";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { lazy, Suspense } from "react";

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

export const Route = createFileRoute("/_authenticated/_sidepanel/calendar/")({
  component: CalendarPage,
  validateSearch: (search) => {
    const view = [
      "dayGridMonth",
      "timeGridWeek",
      "timeGridDay",
      "listWeek",
    ].includes(search.view)
      ? search.view
      : "timeGridWeek";

    let start, end;

    if (
      search.start &&
      search.end &&
      isDurationValid(view, search.start, search.end)
    ) {
      start = search.start;
      end = search.end;
    } else {
      const calculated = calculateStartEndTime(view);
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

  loader: ({ deps: { start, end } }) => {
    return fetchEvents(start, end);
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function CalendarPage() {
  const search = Route.useSearch();
  const loaderData = Route.useLoaderData();
  const router = useRouter();

  return (
    <div>
      <Suspense fallback={<Loading />}>
        <Calendar router={router} view={search.view} eventData={loaderData} />
      </Suspense>
    </div>
  );
}
