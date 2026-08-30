import {
  PlusSignIcon,
  UnavailableIcon,
  User03Icon,
} from "@hugeicons/core-free-icons";
import { Button, Icon } from "@reservations/components";
import { useWindowSize } from "@reservations/lib";
import {
  createFileRoute,
  Link,
  Outlet,
  useRouterState,
} from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/customers/_layout"
)({
  component: CustomersLayout,
});

function CustomersLayout() {
  const { windowSize } = useWindowSize();
  const pathName = useRouterState({ select: (s) => s.location.pathname });

  return (
    <div className="flex h-full min-h-0 flex-col gap-6 px-4 pt-4">
      <div className="flex w-full shrink-0 flex-col gap-4">
        <h1 className="text-text_color text-xl">Customers</h1>
        <div className="flex items-center justify-between">
          <nav
            className="dark:bg-layer_bg flex w-fit rounded-md bg-gray-200 p-1"
          >
            <Link
              activeProps={{
                className: "bg-bg_color text-primary! shadow-sm",
              }}
              activeOptions={{ exact: true }}
              to="/customers/"
              className="text-text_color/70 rounded-md px-4 py-2 text-sm
                font-medium"
            >
              <div className="flex items-center gap-2">
                <Icon icon={User03Icon} styles="size-4" />
                <span className="font-semibold">Customers</span>
              </div>
            </Link>
            <Link
              activeProps={{
                className: "bg-bg_color text-red-600! shadow-sm",
              }}
              to="/customers/blacklist"
              className="text-text_color/70 rounded-md px-4 py-2 text-sm
                font-medium"
            >
              <div className="flex items-center gap-2">
                <Icon icon={UnavailableIcon} styles="size-4" />
                <span className="font-semibold">Blacklisted</span>
              </div>
            </Link>
          </nav>
          {!pathName.includes("blacklist") && (
            <Link from={Route.fullPath} to="new">
              <Button
                variant="primary"
                styles="p-2 md:px-4 w-fit"
                buttonText={windowSize !== "sm" ? "New Customer" : ""}
              >
                <Icon
                  icon={PlusSignIcon}
                  styles="size-6 md:size-5 md:mr-2 text-white"
                />
              </Button>
            </Link>
          )}
        </div>
      </div>
      <div className="flex min-h-0 flex-1 flex-col">
        <Outlet />
      </div>
    </div>
  );
}
