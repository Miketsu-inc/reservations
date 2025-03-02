import Breadcrumbs from "@components/Breadcrumbs";
import CalendarIcon from "@icons/CalendarIcon";
import ChartIcon from "@icons/ChartIcon";
import CustomersIcon from "@icons/CustomersIcon";
import DashboardIcon from "@icons/DashboardIcon";
import HamburgerMenuIcon from "@icons/HamburgerMenuIcon";
import ServicesIcon from "@icons/ServicesIcon";
import SettingsIcon from "@icons/SettingsIcon";
import SidePanelToggleIcon from "@icons/SidePanelToggleIcon";
import SignOutIcon from "@icons/SignOutIcon";
import { useWindowSize } from "@lib/hooks";
import { createFileRoute, Link, Outlet } from "@tanstack/react-router";
import { useCallback, useEffect, useState } from "react";

export const Route = createFileRoute("/_authenticated/_sidepanel")({
  component: SidePanelLayout,
});

function SidePanelLayout() {
  const windowSize = useWindowSize();
  const isWindowSmall = windowSize === "sm" || windowSize === "md";
  const [isOpen, setIsOpen] = useState(isWindowSmall ? false : true);
  const [isCollapsed, setIsCollapsed] = useState(false);

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

  useEffect(() => {
    if (isWindowSmall) {
      setIsOpen(false);
      setIsCollapsed(false);
    } else {
      setIsOpen(true);
    }
  }, [isWindowSmall, setIsOpen, setIsCollapsed]);

  function closeSidePanelHandler() {
    if (isWindowSmall) {
      setIsOpen(false);
    }
  }

  const navigation = [
    {
      href: "/dashboard",
      label: "Dashboard",
      icon: <DashboardIcon styles="h-5 w-5" />,
    },
    {
      href: "/calendar",
      label: "Calendar",
      icon: <CalendarIcon styles="h-5 w-5" />,
    },
    {
      href: "/services",
      label: "Services",
      icon: <ServicesIcon styles="h-5 w-5" />,
    },
    {
      href: "/customers",
      label: "Customers",
      icon: <CustomersIcon styles="h-5 w-5" />,
    },
    {
      href: "#",
      label: "Statistics",
      icon: <ChartIcon styles="h-5 w-5" />,
      isPro: true,
    },
    {
      href: "/settings/profile",
      label: "Settings",
      icon: <SettingsIcon styles="h-5 w-5" />,
    },
    {
      href: "/",
      label: "Sign out",
      icon: <SignOutIcon styles="h-5 w-5" />,
      onClick: handleLogout,
    },
  ];

  return (
    <div className="h-screen overflow-y-auto">
      {/* sticky will have to be replaced with fixed when navlinks are removed */}
      {isWindowSmall && (
        <nav className="bg-layer_bg sticky top-0 z-40 w-full">
          <div className="flex flex-row items-center justify-between px-4 py-2">
            <button
              aria-controls="sidepanel"
              type="button"
              className="text-text_color hover:bg-hvr_gray rounded-lg text-sm focus:ring-2"
              onClick={() => setIsOpen(true)}
            >
              <span className="sr-only">Open sidepanel</span>
              <HamburgerMenuIcon styles="h-6 w-6" />
            </button>
            <img
              className="rounded-full"
              src="https://dummyimage.com/40x40/d156c3/000000.jpg"
            />
          </div>
        </nav>
      )}
      {isOpen && isWindowSmall && (
        <div
          onClick={closeSidePanelHandler}
          className={`fixed inset-0 z-40 bg-black transition-opacity duration-1000 ease-in-out
          ${isOpen ? "opacity-60" : "pointer-events-none opacity-0"}`}
        ></div>
      )}
      <aside
        id="sidepanel"
        className={`${isOpen ? "md:translate-x-0" : "-translate-x-full"}
          ${isCollapsed ? "w-16" : "w-60"} fixed top-0 left-0 z-50 h-dvh overflow-hidden
          transition-all duration-300`}
        aria-label="Sidepanel"
      >
        <div className="bg-layer_bg flex h-full flex-col px-3 py-4">
          <div className="flex flex-row items-center gap-3">
            <img
              className="rounded-full"
              src="https://dummyimage.com/40x40/d156c3/000000.jpg"
            />
            {!isCollapsed && (
              <span
                className={`whitespace-nowrap transition-opacity duration-300
                ${isCollapsed ? "w-0 opacity-0" : "w-auto opacity-100"}`}
              >
                Company name
              </span>
            )}
          </div>
          <ol className="flex flex-1 flex-col space-y-1 pt-5 font-medium">
            {navigation.map((item, index) => (
              <li
                className={`${index === navigation.length - 1 && "mt-auto"}`}
                key={index}
              >
                <Link
                  onClick={item?.onClick ? item.onClick : closeSidePanelHandler}
                  to={item.href}
                  className="group text-text_color hover:bg-hvr_gray flex items-center rounded-lg p-2"
                >
                  <span
                    className="group-hover:text-text_color shrink-0 text-gray-500 transition duration-75
                      dark:text-gray-400"
                  >
                    {item.icon}
                  </span>
                  {!isCollapsed && (
                    <>
                      <span
                        className={`${isCollapsed ? "w-0 opacity-0" : "w-auto opacity-100"} ms-3 flex-1
                        whitespace-nowrap transition-all duration-300`}
                      >
                        {item.label}
                      </span>
                      {item.isPro && (
                        <span
                          className="ms-3 inline-flex items-center justify-center rounded-full bg-gray-300 px-2
                            text-sm font-medium text-gray-800 dark:bg-gray-700 dark:text-gray-300"
                        >
                          Pro
                        </span>
                      )}
                    </>
                  )}
                </Link>
              </li>
            ))}
          </ol>
        </div>
      </aside>
      <div
        className={`${isCollapsed ? "md:ml-16" : "md:ml-60"} flex min-h-screen flex-col
          transition-all duration-300 md:px-4`}
      >
        {!isWindowSmall && (
          <div
            className="flex flex-row items-center gap-2 py-3 text-sm text-gray-800 transition-opacity
              dark:text-gray-300"
          >
            <button
              className="cursor-pointer"
              onClick={() => setIsCollapsed(!isCollapsed)}
            >
              <SidePanelToggleIcon styles="w-4 h-4 stroke-gray-800 dark:stroke-gray-300 hover:stroke-text_color" />
            </button>
            <div className="mx-2 h-3 w-[1px] border-none bg-gray-800 dark:bg-gray-300"></div>
            <Breadcrumbs />
          </div>
        )}
        <div className="bg-layer_bg rounded-lg px-4 py-2">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
