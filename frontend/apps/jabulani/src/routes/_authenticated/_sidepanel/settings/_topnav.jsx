import {
  Briefcase04Icon,
  Calendar02Icon,
  CreditCardIcon,
  TimeScheduleIcon,
  User03Icon,
} from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";
import { TopNavBar, TopNavBarItem } from "@reservations/jabulani/components";
import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_topnav"
)({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="flex h-full min-h-0 min-w-0 flex-col pt-2">
      <TopNavBar>
        <TopNavBarItem from={Route.fullPath} to="/settings/profile">
          <Icon icon={User03Icon} styles="size-5" />
          <span>Profile</span>
        </TopNavBarItem>
        <TopNavBarItem from={Route.fullPath} to="/settings/merchant">
          <Icon icon={Briefcase04Icon} styles="size-5" />
          <span>Merchant</span>
        </TopNavBarItem>
        <TopNavBarItem from={Route.fullPath} to="/settings/calendar">
          <Icon icon={Calendar02Icon} styles="size-5" />
          <span>Calendar</span>
        </TopNavBarItem>
        <TopNavBarItem from={Route.fullPath} to="/settings/billing">
          <Icon icon={CreditCardIcon} styles="size-5" />
          <span>Billing</span>
        </TopNavBarItem>
        <TopNavBarItem from={Route.fullPath} to="/settings/scheduling">
          <Icon icon={TimeScheduleIcon} styles="size-5" />
          <span>Scheduling</span>
        </TopNavBarItem>
      </TopNavBar>
      <div className="flex h-full min-h-0 flex-col items-center px-4 pt-4">
        <div className="w-full max-w-4xl">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
