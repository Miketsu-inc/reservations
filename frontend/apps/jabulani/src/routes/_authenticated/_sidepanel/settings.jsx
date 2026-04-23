import { createFileRoute, Outlet } from "@tanstack/react-router";
import SettingsNavigation from "./settings/-components/SettingsNavigation";

export const Route = createFileRoute("/_authenticated/_sidepanel/settings")({
  component: SettingsLayout,
});

function SettingsLayout() {
  return (
    <div className="flex h-full w-full flex-col p-4">
      <div className="mb-6 w-full shrink-0">
        <div className="flex w-full flex-row">
          <div className="w-16">
            <img
              className="h-auto w-full rounded-3xl object-cover"
              src="https://dummyimage.com/200x200/d156c3/000000.jpg"
            />
          </div>
          <div className="flex flex-col justify-center pl-5 lg:gap-2">
            <h1 className="text-xl font-bold lg:text-4xl">
              {/* {merchantInfo.merchant_name} */}
              Bwnet
            </h1>
            <p className="text-sm">Your personal account</p>
          </div>
        </div>
      </div>
      <div className="flex w-full flex-col gap-8 md:flex-row md:gap-14">
        <div className="w-full md:w-1/4">
          <SettingsNavigation />
        </div>
        <div className="flex-1 md:mt-0">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
