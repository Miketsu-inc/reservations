import {
  Briefcase04Icon,
  Calendar02Icon,
  Clock01Icon,
  CreditCardIcon,
  Location01Icon,
  TimeScheduleIcon,
  UnavailableIcon,
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
    <div className="flex h-full min-h-0 min-w-0 flex-col overflow-hidden pt-2">
      <TopNavBar>
        <TopNavBarItem from={Route.fullPath} to="/settings/merchant">
          <Icon icon={Briefcase04Icon} styles="size-5" />
          <span>Profile</span>
        </TopNavBarItem>
        <TopNavBarItem from={Route.fullPath} to="/settings/business-hours">
          <Icon icon={Clock01Icon} styles="size-5" />
          <span>Business hours</span>
        </TopNavBarItem>
        <TopNavBarItem from={Route.fullPath} to="/settings/scheduling">
          <Icon icon={TimeScheduleIcon} styles="size-5" />
          <span>Scheduling</span>
        </TopNavBarItem>
        <TopNavBarItem from={Route.fullPath} to="/settings/blocked-time-types">
          <Icon icon={UnavailableIcon} styles="size-5" />
          <span>Blocked time types</span>
        </TopNavBarItem>
        <TopNavBarItem from={Route.fullPath} to="/settings/calendar">
          <Icon icon={Calendar02Icon} styles="size-5" />
          <span>Calendar</span>
        </TopNavBarItem>
        <TopNavBarItem from={Route.fullPath} to="/settings/location">
          <Icon icon={Location01Icon} styles="size-5" />
          <span>Location</span>
        </TopNavBarItem>
        <TopNavBarItem from={Route.fullPath} to="/settings/billing">
          <Icon icon={CreditCardIcon} styles="size-5" />
          <span>Billing</span>
        </TopNavBarItem>
      </TopNavBar>
      <div
        className="flex min-h-0 flex-1 flex-col items-center overflow-y-auto
          px-4"
      >
        <div className="w-full max-w-5xl">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
