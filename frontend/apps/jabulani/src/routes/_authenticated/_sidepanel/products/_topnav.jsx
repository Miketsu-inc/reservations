import { ShoppingBag02Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";
import { TopNavBar, TopNavBarItem } from "@reservations/jabulani/components";
import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/products/_topnav"
)({
  component: ProductsLayout,
});

function ProductsLayout() {
  return (
    <div className="flex h-full min-h-0 min-w-0 flex-col overflow-hidden pt-2">
      <TopNavBar>
        <TopNavBarItem from={Route.fullPath} to="/products">
          <Icon icon={ShoppingBag02Icon} styles="size-5" />
          <span>Products</span>
        </TopNavBarItem>
      </TopNavBar>
      <div className="flex min-h-0 flex-1 flex-col overflow-y-auto">
        <Outlet />
      </div>
    </div>
  );
}
