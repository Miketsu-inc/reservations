import Card from "@components/Card";
import { createFileRoute, Outlet } from "@tanstack/react-router";
import SettingsHeader from "./settings/-components/SettingsHeader";
import SettingsNavigation from "./settings/-components/SettingsNavigation";

export const Route = createFileRoute("/_authenticated/_sidepanel/settings")({
  component: SettingsLayout,
  //   loader: () => ({ redirect: "/_authenticated/_sidepanel/settings/profile" }),
  loader: () => ({
    crumb: "Settings",
  }),
});

function SettingsLayout() {
  return (
    <Card>
      <div className="flex h-full w-full flex-col items-center justify-center p-4 lg:px-14">
        <SettingsHeader />
        <div className="flex h-full w-full flex-col gap-8 md:flex-row md:gap-14">
          <div className="w-full md:w-1/4">
            <SettingsNavigation />
          </div>
          <div className="flex-1 md:mt-0">
            <Outlet />
          </div>
        </div>
      </div>
    </Card>
  );
}
