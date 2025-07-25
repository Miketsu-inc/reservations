import Breadcrumbs from "@components/Breadcrumbs";
import { TooltipContent, TooltipTrigger, Tootlip } from "@components/Tooltip";
import CalendarIcon from "@icons/CalendarIcon";
import ChartIcon from "@icons/ChartIcon";
import CustomersIcon from "@icons/CustomersIcon";
import DashboardIcon from "@icons/DashboardIcon";
import HamburgerMenuIcon from "@icons/HamburgerMenuIcon";
import LinkIcon from "@icons/LinkIcon";
import MoonIcon from "@icons/MoonIcon";
import ProductIcon from "@icons/ProductIcon";
import ServicesIcon from "@icons/ServicesIcon";
import SettingsIcon from "@icons/SettingsIcon";
import SidePanelToggleIcon from "@icons/SidePanelToggleIcon";
import SignOutIcon from "@icons/SignOutIcon";
import SunIcon from "@icons/SunIcon";
import { useTheme, useWindowSize } from "@lib/hooks";
import { createFileRoute, Link, Outlet } from "@tanstack/react-router";
import { useCallback, useEffect, useState } from "react";

export const Route = createFileRoute("/_authenticated/_sidepanel")({
  component: SidePanelLayout,
});

function SidePanelLayout() {
  const windowSize = useWindowSize();
  const isWindowSmall = windowSize === "sm" || windowSize === "md";
  const [isOpen, setIsOpen] = useState(isWindowSmall ? false : true);
  const [isCollapsed, setIsCollapsed] = useState(
    localStorage.getItem("sidepanel_collapsed") === "true"
  );

  const { isDarkTheme, switchTheme } = useTheme();

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
    if (windowSize === "sm" || windowSize === "md") {
      setIsOpen(false);
    } else {
      setIsOpen(true);
    }
  }, [windowSize]);

  function closeSidePanelHandler() {
    if (isWindowSmall) {
      setIsOpen(false);
    }
  }

  const navigation = [
    {
      href: "/dashboard",
      label: "Dashboard",
      icon: <DashboardIcon styles="size-5" />,
    },
    {
      href: "/calendar",
      label: "Calendar",
      icon: <CalendarIcon styles="size-5" />,
    },
    {
      href: "/services",
      label: "Services",
      icon: <ServicesIcon styles="size-5" />,
    },
    {
      href: "/customers",
      label: "Customers",
      icon: <CustomersIcon styles="size-5" />,
    },
    {
      href: "/products",
      label: "Products",
      icon: <ProductIcon styles="size-5" />,
    },
    {
      href: "#",
      label: "Statistics",
      icon: <ChartIcon styles="size-5" />,
      isPro: true,
    },
    {
      href: "/settings/profile",
      label: "Settings",
      icon: <SettingsIcon styles="size-5" />,
    },
    {
      href: "/",
      label: "Sign out",
      icon: <SignOutIcon styles="size-5" />,
      onClick: handleLogout,
    },
  ];

  function withConditionalTooltip(condition, content, children) {
    if (!condition) return children;

    return (
      <Tootlip>
        <TooltipTrigger asChild>{children}</TooltipTrigger>
        <TooltipContent side="right" sideOffset={4}>
          <p>{content}</p>
        </TooltipContent>
      </Tootlip>
    );
  }

  return (
    <div className="h-screen overflow-y-auto">
      {isWindowSmall && (
        <nav className="bg-layer_bg sticky top-0 z-20 w-full">
          <div className="flex flex-row items-center justify-between px-4 py-2">
            <button
              aria-controls="sidepanel"
              type="button"
              className="text-text_color hover:bg-primary/20 rounded-lg text-sm"
              onClick={() => setIsOpen(true)}
            >
              <span className="sr-only">Open sidepanel</span>
              <HamburgerMenuIcon styles="h-6 w-6" />
            </button>
            <div className="flex flex-row gap-4">
              <Link
                className="hover:bg-primary/20 flex flex-row items-center gap-2
                  rounded-lg p-2"
                from={Route.fullPath}
                to="/m/bwnet"
              >
                <LinkIcon styles="size-5" />
                <span>Live booking page</span>
              </Link>
              <button
                className="cursor-pointer transition-transform duration-300"
                onClick={switchTheme}
              >
                {isDarkTheme ? (
                  <SunIcon styles="size-5" />
                ) : (
                  <MoonIcon styles="size-5" />
                )}
              </button>
            </div>
          </div>
        </nav>
      )}
      {isOpen && isWindowSmall && (
        <div
          onClick={closeSidePanelHandler}
          className={`fixed inset-0 z-20 bg-black transition-opacity
          duration-1000 ease-in-out
          ${isOpen ? "opacity-60" : "pointer-events-none opacity-0"}`}
        ></div>
      )}
      <aside
        id="sidepanel"
        className={`${isOpen ? "md:translate-x-0" : "-translate-x-full"}
          ${!isWindowSmall && isCollapsed ? "w-16" : "w-60"} fixed top-0 left-0
          z-30 h-dvh overflow-hidden transition-all duration-300`}
        aria-label="Sidepanel"
      >
        <div
          className="bg-layer_bg border-border_color flex h-full flex-col
            border-r px-3 py-4"
        >
          <div
            className={`${!isWindowSmall && isCollapsed ? "w-10" : "w-40"} flex
              h-10 flex-row items-center gap-3 transition-normal duration-300
              ease-in-out`}
          >
            <img
              className="h-full rounded-lg object-cover"
              src="https://dummyimage.com/160x40/d156c3/000000.jpg"
            />
          </div>
          <ol className="flex flex-1 flex-col space-y-2 pt-8 font-medium">
            {navigation.map((item, index) => (
              <li
                className={`${index === navigation.length - 1 ? "mt-auto" : ""}`}
                key={index}
              >
                {withConditionalTooltip(
                  !isWindowSmall && isCollapsed,
                  item.label,
                  <Link
                    onClick={
                      item?.onClick ? item.onClick : closeSidePanelHandler
                    }
                    to={item.href}
                    activeProps={{
                      className: "bg-primary/20 *:text-primary! *:duration-0",
                    }}
                    className="text-text_color hover:bg-primary/20 flex
                      items-center rounded-lg p-2"
                  >
                    <span
                      className="shrink-0 text-gray-500 transition duration-75
                        dark:text-gray-400"
                    >
                      {item.icon}
                    </span>
                    {(!isCollapsed || isWindowSmall) && (
                      <>
                        <span
                          className={`${!isWindowSmall && isCollapsed ? "w-0 opacity-0" : "w-auto opacity-100"}
                          ms-3 flex-1 whitespace-nowrap transition-opacity
                          duration-300`}
                        >
                          {item.label}
                        </span>
                        {item.isPro && (
                          <span
                            className="ms-3 inline-flex items-center
                              justify-center rounded-full bg-gray-300 px-2
                              text-sm font-medium text-gray-800 dark:bg-gray-700
                              dark:text-gray-300"
                          >
                            Pro
                          </span>
                        )}
                      </>
                    )}
                  </Link>
                )}
              </li>
            ))}
          </ol>
        </div>
      </aside>
      {!isWindowSmall && (
        <div
          className={`${isCollapsed ? "md:ml-16" : "md:ml-60"} bg-layer_bg
          border-b-border_color border-b py-2 pl-4 transition-[margin]
          duration-300`}
        >
          <div className="mr-4 flex flex-row items-center justify-between">
            <div
              className="flex flex-row items-center gap-2 py-3 text-sm
                text-gray-800 dark:text-gray-300"
            >
              <button
                className="cursor-pointer"
                onClick={() => {
                  localStorage.setItem("sidepanel_collapsed", !isCollapsed);
                  setIsCollapsed(!isCollapsed);
                }}
              >
                <SidePanelToggleIcon
                  styles="w-4 h-4 stroke-gray-800 dark:stroke-gray-300
                    hover:stroke-text_color"
                />
              </button>
              <div
                className="mx-2 h-3 w-px border-none bg-gray-800
                  dark:bg-gray-300"
              ></div>
              <Breadcrumbs />
            </div>
            <div className="flex flex-row gap-4">
              <Link
                className="hover:bg-primary/20 flex flex-row items-center gap-2
                  rounded-lg p-2"
                from={Route.fullPath}
                to="/m/bwnet"
              >
                <LinkIcon styles="size-5" />
                <span>Live booking page</span>
              </Link>
              <button
                className="cursor-pointer transition-transform duration-300"
                onClick={switchTheme}
              >
                {isDarkTheme ? (
                  <SunIcon styles="size-5" />
                ) : (
                  <MoonIcon styles="size-5" />
                )}
              </button>
            </div>
          </div>
        </div>
      )}
      <div
        className={`${!isWindowSmall && isCollapsed ? "md:ml-16" : "md:ml-60"}
          flex min-h-screen flex-col transition-[margin] duration-300 md:px-4
          md:pt-4`}
      >
        <Outlet />
      </div>
    </div>
  );
}
