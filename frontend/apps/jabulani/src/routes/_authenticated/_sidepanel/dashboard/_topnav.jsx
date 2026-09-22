import { DashboardSquare01Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";
import { TopNavBar, TopNavBarItem } from "@reservations/jabulani/components";
import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/dashboard/_topnav"
)({
  component: DashboardLayout,
});

function DashboardLayout() {
  return (
    <div className="flex h-full min-h-0 min-w-0 flex-col overflow-hidden pt-2">
      <TopNavBar>
        <TopNavBarItem from={Route.fullPath} to="/dashboard">
          <Icon icon={DashboardSquare01Icon} styles="size-5" />
          <span>Dashboard</span>
        </TopNavBarItem>
      </TopNavBar>
      <div className="flex min-h-0 flex-1 flex-col overflow-y-auto">
        <Outlet />
      </div>
    </div>
  );
}
