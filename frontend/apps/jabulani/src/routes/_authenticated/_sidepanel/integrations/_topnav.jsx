import { Calendar02Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";
import { TopNavBar, TopNavBarItem } from "@reservations/jabulani/components";
import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/integrations/_topnav"
)({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="flex h-full min-h-0 min-w-0 flex-col overflow-hidden pt-2">
      <TopNavBar>
        <TopNavBarItem from={Route.fullPath} to="/integrations/calendar">
          <Icon icon={Calendar02Icon} styles="size-5" />
          <span>Calendar</span>
        </TopNavBarItem>
      </TopNavBar>
      <div
        className="flex min-h-0 flex-1 flex-col items-center overflow-y-auto
          px-4 pt-4"
      >
        <div className="w-full max-w-4xl">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
