import { HierarchyIcon, MailAccount01Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";
import { TopNavBar, TopNavBarItem } from "@reservations/jabulani/components";
import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/_sidepanel/team/_topnav")(
  {
    component: RouteComponent,
  }
);

function RouteComponent() {
  return (
    <div className="flex h-full min-h-0 flex-col pt-2">
      <TopNavBar>
        <TopNavBarItem from={Route.fullPath} to="/team/members">
          <Icon icon={HierarchyIcon} styles="size-5" />
          <span>Members</span>
        </TopNavBarItem>
        <TopNavBarItem from={Route.fullPath} to="/team/invitations">
          <Icon icon={MailAccount01Icon} styles="size-5" />
          <span>Invitations</span>
        </TopNavBarItem>
      </TopNavBar>
      <div className="flex h-full min-h-0 flex-col px-4 pt-4">
        <Outlet />
        {/* <div>
        </div> */}
      </div>
    </div>
  );
}
