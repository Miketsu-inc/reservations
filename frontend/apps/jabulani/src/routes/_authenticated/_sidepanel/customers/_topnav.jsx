import { UnavailableIcon, User03Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";
import { TopNavBar, TopNavBarItem } from "@reservations/jabulani/components";
import {
  createFileRoute,
  Outlet,
  useRouterState,
} from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/customers/_topnav"
)({
  component: CustomersLayout,
});

function CustomersLayout() {
  const pathName = useRouterState({ select: (s) => s.location.pathname });
  const isCustomersActive =
    pathName === "/customers" ||
    (pathName.startsWith("/customers/") &&
      !pathName.startsWith("/customers/blacklist"));

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden pt-2">
      <TopNavBar>
        <TopNavBarItem
          styles={isCustomersActive ? "after:opacity-100" : ""}
          from={Route.fullPath}
          to="/customers"
          activeProps={{}}
        >
          <Icon icon={User03Icon} styles="size-5" />
          <span>Customers</span>
        </TopNavBarItem>
        <TopNavBarItem
          from={Route.fullPath}
          to="/customers/blacklist"
          activeOptions={{ exact: true }}
        >
          <Icon icon={UnavailableIcon} styles="size-5" />
          <span>Blacklisted</span>
        </TopNavBarItem>
      </TopNavBar>
      <div className="flex min-h-0 flex-1 flex-col overflow-y-auto px-4 pt-4">
        <Outlet />
      </div>
    </div>
  );
}
