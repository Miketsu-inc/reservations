import CalendarIcon from "../../assets/icons/CalendarIcon";
import ChartIcon from "../../assets/icons/ChartIcon";
import DashboardIcon from "../../assets/icons/DashboardIcon";
import SettingsIcon from "../../assets/icons/SettingsIcon";
import SignOutIcon from "../../assets/icons/SignOutIcon";
import Calendar from "./Calendar";
import SidePanel from "./SidePanel";
import SidePanelIcon from "./SidePanelIcon";
import SidePanelItem from "./SidePanelItem";

export default function DashboardPage() {
  return (
    <>
      <SidePanel
        profileImage="https://dummyimage.com/40x40/000/fff.png&text=logo"
        profileText="Company name"
      >
        <SidePanelItem link="#" text="Dashboard">
          <SidePanelIcon>
            <DashboardIcon styles="h-5 w-5" />
          </SidePanelIcon>
        </SidePanelItem>
        <SidePanelItem link="#" text="Calendar">
          <SidePanelIcon>
            <CalendarIcon styles="h-5 w-5" />
          </SidePanelIcon>
        </SidePanelItem>
        <SidePanelItem link="#" text="Statistics" isPro={true}>
          <SidePanelIcon>
            <ChartIcon styles="h-5 w-5" />
          </SidePanelIcon>
        </SidePanelItem>
        <SidePanelItem link="#" text="Settings">
          <SidePanelIcon>
            <SettingsIcon styles="h-5 w-5" />
          </SidePanelIcon>
        </SidePanelItem>
        <span className="flex-1"></span>
        <SidePanelItem link="#" text="Sign out">
          <SidePanelIcon>
            <SignOutIcon styles="h-5 w-5" />
          </SidePanelIcon>
        </SidePanelItem>
      </SidePanel>
      <div className="light p-4 sm:ml-64">
        <div className="rounded-lg bg-bg_color p-4">
          <Calendar />
        </div>
      </div>
    </>
  );
}
