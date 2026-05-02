import {
  CalendarIcon,
  ChartIcon,
  CustomersIcon,
  DashboardIcon,
  HamburgerMenuIcon,
  LinkIcon,
  MoonIcon,
  ProductIcon,
  ServicesIcon,
  SettingsIcon,
  SidePanelToggleIcon,
  SignOutIcon,
  SunIcon,
} from "@reservations/assets";
import {
  Avatar,
  Loading,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
  ServerError,
  TooltipContent,
  TooltipTrigger,
  Tootlip,
} from "@reservations/components";
import { meQueryOptions, useTheme, useWindowSize } from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, Outlet } from "@tanstack/react-router";
import { useCallback, useState } from "react";

export const Route = createFileRoute("/_authenticated/_sidepanel")({
  component: SidePanelLayout,
  loader: async ({ context: { queryClient } }) => {
    await queryClient.ensureQueryData(meQueryOptions());
  },
  pendingComponent: Loading,
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function SidePanelLayout() {
  const navigate = Route.useNavigate();
  const windowSize = useWindowSize();
  const isWindowSmall = windowSize === "sm" || windowSize === "md";
  const [isOpenend, setIsOpened] = useState(false);
  const [isCollapsed, setIsCollapsed] = useState(
    localStorage.getItem("sidepanel_collapsed") === "true"
  );

  const { data: user, isLoading } = useQuery(meQueryOptions());

  const isOpen = isWindowSmall ? isOpenend : true;

  const { isDarkTheme, switchTheme } = useTheme();

  const handleLogout = useCallback(async () => {
    try {
      await fetch("api/v1/auth/logout", {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
      });

      navigate({
        from: Route.fullPath,
        to: "/login",
      });
    } catch (err) {
      console.log(err.message);
    }
  }, [navigate]);

  function closeSidePanelHandler() {
    if (isWindowSmall) {
      setIsOpened(false);
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
      href: "/team",
      label: "Team",
      icon: <CustomersIcon styles="size-5" />,
    },
    {
      href: "/products",
      label: "Products",
      icon: <ProductIcon styles="size-5" />,
    },
    {
      href: "/integrations",
      label: "Integrations",
      icon: <CalendarIcon styles="size-5" />,
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

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div className="bg-layer_bg flex h-dvh flex-col overflow-hidden">
      {isWindowSmall ? (
        <nav className="sticky top-0 z-20 w-full">
          <div className="flex flex-row items-center justify-between px-4 py-2">
            <button
              aria-controls="sidepanel"
              type="button"
              className="text-text_color hover:bg-primary/20 rounded-lg text-sm"
              onClick={() => setIsOpened(true)}
            >
              <span className="sr-only">Open sidepanel</span>
              <HamburgerMenuIcon styles="size-6" />
            </button>
            <div className="flex flex-row gap-4">
              <a
                className="hover:bg-primary/20 flex flex-row items-center gap-2
                  rounded-lg p-2"
                href="http://reservations.local:3000/m/bwnet"
              >
                <LinkIcon styles="size-5" />
                <span>Live booking page</span>
              </a>
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
      ) : (
        <header className={"py-2 pl-4 transition-[margin] duration-300"}>
          <div className="mr-4 flex flex-row items-center justify-between">
            <div className="flex flex-row items-center gap-4">
              <div
                className={`${!isWindowSmall && isCollapsed ? "w-8" : "w-40"}
                  flex h-10 flex-row items-center gap-3 transition-normal
                  duration-300 ease-in-out`}
              >
                <img
                  className="h-full rounded-lg object-cover"
                  src="https://dummyimage.com/160x40/d156c3/000000.jpg"
                />
              </div>
            </div>
            <div className="flex flex-row gap-4">
              <a
                className="hover:bg-primary/20 flex flex-row items-center gap-2
                  rounded-lg p-2"
                href="http://reservations.local:3000/m/bwnet"
              >
                <LinkIcon styles="size-5" />
                <span>Live booking page</span>
              </a>
              <button
                className="cursor-pointer transition-transform duration-300"
                onClick={switchTheme}
              >
                {isDarkTheme ? (
                  <SunIcon
                    styles="size-5 stroke-gray-800 dark:stroke-gray-300
                      hover:stroke-text_color"
                  />
                ) : (
                  <MoonIcon
                    styles="size-5 stroke-gray-800 dark:stroke-gray-300
                      hover:stroke-text_color"
                  />
                )}
              </button>
            </div>
          </div>
        </header>
      )}
      <div className="flex flex-1 overflow-hidden">
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
            ${!isWindowSmall && isCollapsed ? "w-16" : "w-60"}
            ${isWindowSmall ? "fixed top-0" : "relative"} bg-layer_bg left-0
            z-30 flex h-full flex-col overflow-hidden
            transition-[width,translate] duration-300`}
          aria-label="Sidepanel"
        >
          {isWindowSmall && (
            <div className="shrink-0 p-4">
              <div
                className={`${!isWindowSmall && isCollapsed ? "w-10" : "w-40"}
                flex h-10 flex-row items-center gap-3 transition-normal
                duration-300 ease-in-out`}
              >
                <img
                  className="h-full rounded-lg object-cover"
                  src="https://dummyimage.com/160x40/d156c3/000000.jpg"
                />
              </div>
            </div>
          )}
          <div
            className="flex min-h-0 flex-1 flex-col justify-between px-3 py-2"
          >
            <ol className="flex flex-col space-y-2">
              {navigation.map((item, index) => (
                <li key={index}>
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
                      className={`${isCollapsed && !isWindowSmall ? "h-10 w-10 justify-center" : "h-10 px-2"}
                      text-text_color hover:bg-primary/20 flex items-center
                      rounded-lg`}
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
                            ms-3 flex-1 whitespace-nowrap
                            transition-[opacity,width] duration-300`}
                          >
                            {item.label}
                          </span>
                          {item.isPro && (
                            <span
                              className="ms-3 inline-flex items-center
                                justify-center rounded-full bg-gray-300 px-2
                                text-sm font-medium text-gray-800
                                dark:bg-gray-700 dark:text-gray-300"
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
            <Popover>
              <PopoverTrigger>
                <div
                  className={`${!isWindowSmall && isCollapsed ? "px-0 py-2 hover:bg-transparent" : "gap-2 p-2"}
                    hover:bg-primary/20 flex flex-row items-center rounded-lg
                    transition-[padding,gap] duration-300 hover:cursor-pointer`}
                >
                  <Avatar
                    styles="size-10! text-sm! shrink-0"
                    initials={`${user?.first_name.charAt(0)}${user?.last_name.charAt(0)}`}
                  />
                  <div
                    className={`${!isWindowSmall && isCollapsed ? "w-0 opacity-0" : "w-auto opacity-100"}
                      flex flex-1 flex-col items-start gap-0.5
                      whitespace-nowrap`}
                  >
                    <span className="text-sm">{`${user?.first_name} ${user?.last_name}`}</span>
                    <span className="text-xs">{user?.memberships[0].role}</span>
                  </div>
                </div>
              </PopoverTrigger>
              <PopoverContent align="start">
                <div
                  className="*:hover:bg-primary/20 space-y-2 text-sm
                    *:cursor-pointer *:rounded-lg *:p-2"
                >
                  <div>
                    <PopoverClose asChild>
                      <button
                        className="flex cursor-pointer flex-row items-center
                          gap-2"
                        onClick={() => {
                          localStorage.setItem(
                            "sidepanel_collapsed",
                            !isCollapsed
                          );
                          setIsCollapsed(!isCollapsed);
                        }}
                      >
                        <SidePanelToggleIcon styles="size-4 stroke-text_color" />
                        {isCollapsed
                          ? "Expand sidepanel"
                          : "Collapse sidepanel"}
                      </button>
                    </PopoverClose>
                  </div>
                  <div>
                    <button
                      className="flex cursor-pointer flex-row items-center
                        gap-2"
                      onClick={handleLogout}
                    >
                      <SignOutIcon styles="size-5" />
                      Sign out
                    </button>
                  </div>
                </div>
              </PopoverContent>
            </Popover>
          </div>
        </aside>
        <div
          className={`flex min-h-0 flex-1 flex-col transition-[margin]
            duration-300 md:pt-2 md:pr-2`}
        >
          <div
            className={`${!isWindowSmall ? "md:px-4 md:pt-4" : ""} bg-bg_color
              flex-1 overflow-y-auto rounded-t-xl`}
          >
            <Outlet />
          </div>
        </div>
      </div>
    </div>
  );
}
