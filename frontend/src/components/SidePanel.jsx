import CalendarIcon from "@icons/CalendarIcon";
import ChartIcon from "@icons/ChartIcon";
import CustomersIcon from "@icons/CustomersIcon";
import DashboardIcon from "@icons/DashboardIcon";
import HamburgerMenuIcon from "@icons/HamburgerMenuIcon";
import ServicesIcon from "@icons/ServicesIcon";
import SettingsIcon from "@icons/SettingsIcon";
import SignOutIcon from "@icons/SignOutIcon";
import { useClickOutside, useWindowSize } from "@lib/hooks";
import { useCallback, useEffect, useRef, useState } from "react";
import SidePanelItem from "./SidePanelItem";
import SidePanelProfile from "./SidePanelProfile";

export default function SidePanel({ profileImage, profileText }) {
  const windowSize = useWindowSize();
  const [isOpen, setIsOpen] = useState(
    windowSize !== "sm" && windowSize !== "md" ? true : false
  );
  const sidePanelRef = useRef();
  useClickOutside(sidePanelRef, closeSidePanelHandler);

  useEffect(() => {
    if (windowSize === "sm" || windowSize === "md") {
      setIsOpen(false);
    } else {
      setIsOpen(true);
    }
  }, [windowSize, setIsOpen]);

  function sidePanelClickHandler() {
    setIsOpen(true);
  }

  function closeSidePanelHandler() {
    if (windowSize === "sm" || windowSize === "md") {
      setIsOpen(false);
    }
  }

  const handleLogout = useCallback(async () => {
    try {
      await fetch("api/v1/auth/user/logout", {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
      });
    } catch (err) {
      console.log(err.message);
    }
  }, []);

  return (
    <>
      <button
        aria-controls="sidepanel"
        type="button"
        className="ms-3 mt-2 inline-flex items-center rounded-lg p-2 text-sm text-text_color
          hover:bg-hvr_gray focus:outline-hidden focus:ring-2 focus:ring-gray-200 md:hidden
          dark:focus:ring-gray-600"
        onClick={sidePanelClickHandler}
      >
        <span className="sr-only">Open sidepanel</span>
        <HamburgerMenuIcon styles="h-6 w-6" />
      </button>
      <aside
        ref={sidePanelRef}
        id="sidepanel"
        className={`${isOpen ? "md:translate-x-0" : "-translate-x-full"} fixed left-0 top-0 z-40
          h-screen w-64 overflow-y-auto transition-transform`}
        aria-label="Sidepanel"
      >
        <div className="flex h-full flex-col bg-layer_bg px-3 py-4">
          <SidePanelProfile
            image={profileImage}
            text={profileText}
            closeSidePanel={closeSidePanelHandler}
            windowSize={windowSize}
          />
          <hr className="my-4"></hr>
          <div className="flex flex-1 flex-col space-y-2 font-medium">
            <SidePanelItem link="/dashboard" text="Dashboard">
              <DashboardIcon styles="h-5 w-5" />
            </SidePanelItem>
            <SidePanelItem link="/calendar" text="Calendar">
              <CalendarIcon styles="h-5 w-5" />
            </SidePanelItem>
            <SidePanelItem link="/services" text="Services">
              <ServicesIcon styles="h-5 w-5" />
            </SidePanelItem>
            <SidePanelItem link="/customers" text="Customers">
              <CustomersIcon styles="h-5 w-5 fill-transparent" />
            </SidePanelItem>
            <SidePanelItem link="#" text="Statistics" isPro={true}>
              <ChartIcon styles="h-5 w-5" />
            </SidePanelItem>
            <SidePanelItem link="/settings/profile" text="Settings">
              <SettingsIcon styles="h-5 w-5" />
            </SidePanelItem>
            <span className="flex-1"></span>
            <SidePanelItem link="/" text="Sign out" action={handleLogout}>
              <SignOutIcon styles="h-5 w-5" />
            </SidePanelItem>
          </div>
        </div>
      </aside>
    </>
  );
}
