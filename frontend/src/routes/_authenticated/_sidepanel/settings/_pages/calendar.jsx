import { createFileRoute } from "@tanstack/react-router";
import SectionHeader from "../-components/SectionHeader";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_pages/calendar"
)({
  component: CalendarPage,
});

function CalendarPage() {
  return (
    <div>
      <div className="flex flex-col gap-4">
        <SectionHeader title="Calendar" styles="" />
        <p className="text-text_color/70">
          Coustumize your own calendar the way you like.
        </p>
      </div>
    </div>
  );
}
