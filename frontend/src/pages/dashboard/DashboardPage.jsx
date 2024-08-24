import XIcon from "../../assets/XIcon";
import Calendar from "./Calendar";
import SidePanel from "./SidePanel";
import SidePanelItem from "./SidePanelItem";

export default function DashboardPage() {
  return (
    <>
      <SidePanel
        profileImage="https://dummyimage.com/40x40/000/fff.png&text=logo"
        profileText="Company name"
      >
        <SidePanelItem link="#" text="Dashboard">
          {/* <XIcon
            styles="h-5 w-5 flex-shrink-0 text-gray-500 transition duration-75
              group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
          /> */}
        </SidePanelItem>
        <SidePanelItem link="#" text="Calendar" />
        <SidePanelItem link="#" text="Pro feature" isPro={true} />
        <SidePanelItem link="#" text="Settings" />
        <SidePanelItem link="#" text="Sign out" />
      </SidePanel>
      <div className="p-4 sm:ml-64">
        <div className="rounded-lg bg-white p-4">
          <Calendar />
        </div>
      </div>
    </>
  );
}
