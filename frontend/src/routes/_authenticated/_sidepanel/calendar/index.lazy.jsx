import Loading from "@components/Loading";
import { createLazyFileRoute } from "@tanstack/react-router";
import { Suspense } from "react";
import Calendar from "./-components/Calendar";

export const Route = createLazyFileRoute(
  "/_authenticated/_sidepanel/calendar/"
)({
  component: CalendarPage,
});

function CalendarPage() {
  return (
    <span className="light">
      <Suspense fallback={<Loading />}>
        <Calendar />
      </Suspense>
    </span>
  );
}
