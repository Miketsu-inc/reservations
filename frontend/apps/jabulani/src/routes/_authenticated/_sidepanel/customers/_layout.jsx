import { BanIcon, PersonIcon, PlusIcon } from "@reservations/assets";
import { Button } from "@reservations/components";
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
  const windowSize = useWindowSize();
  const pathName = useRouterState({ select: (s) => s.location.pathname });

  return (
    <div className="flex flex-col justify-center gap-6 px-4 py-4 md:px-0">
      <div className="flex w-full flex-col gap-4">
        <h1 className="text-text_color text-2xl font-bold">Customers</h1>
        <div className="flex items-center justify-between">
          <nav
            className="dark:bg-layer_bg flex w-fit rounded-md bg-gray-200 p-1"
          >
            <Link
              activeProps={{
                className: " bg-bg_color text-primary! shadow-sm",
              }}
              activeOptions={{ exact: true }}
              to="/customers/"
              className="text-text_color/70 rounded-md px-4 py-2 text-sm
                font-medium"
            >
              <div className="flex items-center gap-2">
                <PersonIcon styles="size-4 fill-current" />
                <span className="font-semibold">Customers</span>
              </div>
            </Link>
            <Link
              activeProps={{
                className: " bg-bg_color text-red-600! shadow-sm",
              }}
              to="blacklist"
              className="text-text_color/70 rounded-md px-4 py-2 text-sm
                font-medium"
            >
              <div className="flex items-center gap-2">
                <BanIcon styles="size-4" />
                <span className="font-semibold">Blacklisted</span>
              </div>
            </Link>
          </nav>
          {!pathName.includes("blacklist") && (
            <Link from={Route.fullPath} to="new">
              <Button
                variant="primary"
                styles="sm:py-2 sm:px-4 w-fit p-2"
                buttonText={windowSize !== "sm" ? "New Customer" : ""}
              >
                <PlusIcon styles="size-6 sm:size-5 sm:mr-2 sm:mb-0.5 text-white" />
              </Button>
            </Link>
          )}
        </div>
      </div>
      <Outlet />
    </div>
  );
}
