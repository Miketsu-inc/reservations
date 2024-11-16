import BackArrowIcon from "../../assets/icons/BackArrowIcon";
import SettingsIcon from "../../assets/icons/SettingsIcon";
import Selector from "../../components/Selector";
import SelectorItem from "../../components/SelectorItem";
import SidePanel from "../dashboard/SidePanel";

export default function SettingsPage() {
  return (
    <div className="bg-bg_color pt-6 text-text_color dark:bg-layer_bg">
      <SidePanel
        profileImage="https://dummyimage.com/40x40/000/fff.png&text=logo"
        profileText="Company name"
      />
      <div className="w-full sm:ml-64 sm:w-auto">
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
              dropdownStyles=""
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
              <SelectorItem styles="pl-8" key="5" value="">
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
        </div>
      </div>
    </div>
  );
}
