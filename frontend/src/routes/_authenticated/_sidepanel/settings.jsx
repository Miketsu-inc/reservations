import Button from "@components/Button";
import Selector from "@components/Selector";
import SelectorItem from "@components/SelectorItem";
import BackArrowIcon from "@icons/BackArrowIcon";
import SettingsIcon from "@icons/SettingsIcon";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { createFileRoute } from "@tanstack/react-router";
import { useCallback } from "react";

export const Route = createFileRoute("/_authenticated/_sidepanel/settings")({
  component: SettingsPage,
});

function SettingsPage() {
  const navigate = Route.useNavigate();

  const logOutOnAllDevices = useCallback(async () => {
    const response = await fetch("api/v1/auth/user/logout/all", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    });

    let result = await response;
    if (!response.ok) {
      result = result.json();
      console.log(result.error.message);
    }

    invalidateLocalSotrageAuth(401);
    navigate({
      from: Route.fullPath,
      to: "/",
    });
  }, [navigate]);

  return (
    <div className="bg-bg_color pt-6 text-text_color dark:bg-layer_bg">
      <h1 className="mt-3 flex justify-between px-4 pb-2 text-left text-2xl font-bold">
        <span>Settings</span>
        <SettingsIcon styles="h-8 w-8 sm:h-10 sm:w-10 dark:text-gray-400 text-gray-700" />
      </h1>
      <div
        className="items-left flex w-full flex-col justify-center bg-gray-200 px-6 py-4
          text-text_color dark:bg-bg_color"
      >
        <h2 className="mt-4 text-left text-gray-600 dark:text-gray-300">
          General
        </h2>
        <div
          className="bg-layer-bg flex flex-col items-center justify-center rounded-lg
            dark:bg-bg_color"
        >
          <button
            className="mt-4 flex w-full flex-col gap-2 rounded-md bg-bg_color p-2 text-left
              dark:bg-layer_bg"
          >
            <span>Give us a Feedback!</span>
            <span className="text-sm">Rate our website</span>
          </button>
          <div className="mt-6 w-full rounded-t-md bg-bg_color px-4 py-3 dark:bg-layer_bg">
            Switch themes
          </div>
          <Selector
            defaultValue="Organization data"
            styles="p-2 px-4 bg-bg_color dark:bg-layer_bg mt-1"
          >
            <SelectorItem styles="pl-8" key="3" value="">
              Email
            </SelectorItem>
            <SelectorItem styles="pl-8" key="4" value="">
              Description
            </SelectorItem>
            <SelectorItem styles="pl-8" key="5" value="">
              Change password
            </SelectorItem>
            <SelectorItem styles="pl-8" key="6" value="">
              Add / remove services
            </SelectorItem>
          </Selector>
          <div className="mt-1 w-full rounded-b-md bg-bg_color px-4 py-3 dark:bg-layer_bg">
            FAQ
          </div>
        </div>
        <h2 className="mt-8 text-left text-gray-600 dark:text-gray-300">
          Other
        </h2>
        <button
          className="mt-4 flex w-full justify-between gap-2 rounded-t-md bg-bg_color p-2 py-3
            text-left dark:bg-layer_bg"
        >
          <span>Terms and privacy policy</span>
          <BackArrowIcon styles="rotate-180" />
        </button>
        <button
          className="mb-6 mt-1 flex w-full justify-between gap-2 rounded-b-md bg-bg_color p-2 py-3
            text-left dark:bg-layer_bg"
        >
          <span>Notifications</span>
          <BackArrowIcon styles="rotate-180" />
        </button>
        <h2 className="mt-8 pb-4 text-left text-gray-600 dark:text-gray-300">
          Danger zone
        </h2>
        <Button
          styles="bg-red-700 py-2"
          buttonText="Log out on all devices"
          onClick={logOutOnAllDevices}
        />
      </div>
    </div>
  );
}
