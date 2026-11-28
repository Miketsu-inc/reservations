import { HierarchyIcon, MailAccount01Icon } from "@hugeicons/core-free-icons";
import { Icon, ServerError } from "@reservations/components";
import { TopNavBar, TopNavBarItem } from "@reservations/jabulani/components";
import { ACTIONS, createPermissionChecker } from "@reservations/jabulani/lib";
import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/_sidepanel/team/_topnav")(
  {
    beforeLoad: ({ context: { authContext } }) => {
      const can = createPermissionChecker(
        authContext.permissions,
        authContext.employeeId
      );

      if (!can.do(ACTIONS.MANAGE_TEAM)) {
        throw new Error("You don't have access to this page");
      }
    },
    component: RouteComponent,
    errorComponent: ({ error }) => <ServerError error={error.message} />,
  }
);

function RouteComponent() {
  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden pt-2">
      <TopNavBar>
        <TopNavBarItem from={Route.fullPath} to="/team/members">
          <Icon icon={HierarchyIcon} styles="size-4.5" />
          <span>Members</span>
        </TopNavBarItem>
        <TopNavBarItem from={Route.fullPath} to="/team/invitations">
          <Icon icon={MailAccount01Icon} styles="size-4.5" />
          <span>Invitations</span>
        </TopNavBarItem>
      </TopNavBar>
      <div className="flex min-h-0 flex-1 flex-col overflow-y-auto px-4 pt-4">
        <Outlet />
      </div>
    </div>
  );
}
