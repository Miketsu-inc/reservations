import {
  Calendar02Icon,
  Chart03Icon,
  CorporateIcon,
  DashboardSquare01Icon,
  LayoutLeftIcon,
  Link01Icon,
  LinkSquare02Icon,
  Logout05Icon,
  Menu01Icon,
  Moon02Icon,
  Note01Icon,
  Settings02Icon,
  ShoppingBag02Icon,
  Sun03Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import {
  Avatar,
  Icon,
  Loading,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
  ServerError,
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  meQueryOptions,
  TANGO_URL,
  useTheme,
  useWindowSize,
} from "@reservations/lib";
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
  const { isWindowSmall } = useWindowSize();
  const [isOpenend, setIsOpened] = useState(false);
  const [isCollapsed, setIsCollapsed] = useState(
    localStorage.getItem("sidepanel_collapsed") === "true"
  );

  const { data: user, isLoading } = useQuery(meQueryOptions());
  const { role, merchantUrlName } = useAuth();

  const isOpen = isWindowSmall ? isOpenend : true;

  const { isDarkTheme, switchTheme } = useTheme();

  const handleLogout = useCallback(async () => {
    try {
      await fetch("/api/v1/auth/logout", {
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
      icon: <Icon icon={DashboardSquare01Icon} styles="size-5" />,
    },
    {
      href: "/calendar",
      label: "Calendar",
      icon: <Icon icon={Calendar02Icon} styles="size-5" />,
    },
    {
      href: "/services",
      label: "Services",
      icon: <Icon icon={Note01Icon} styles="size-5" />,
    },
    {
      href: "/customers",
      label: "Customers",
      icon: <Icon icon={UserGroupIcon} styles="size-5" />,
    },
    {
      href: "/team",
      label: "Team",
      icon: <Icon icon={CorporateIcon} styles="size-5" />,
    },
    {
      href: "/products",
      label: "Products",
      icon: <Icon icon={ShoppingBag02Icon} styles="size-5" />,
    },
    {
      href: "/integrations",
      label: "Integrations",
      icon: <Icon icon={Link01Icon} styles="size-5" />,
    },
    {
      href: "#",
      label: "Statistics",
      icon: <Icon icon={Chart03Icon} styles="size-5" />,
      isPro: true,
    },
    {
      href: "/settings",
      label: "Settings",
      icon: <Icon icon={Settings02Icon} styles="size-5" />,
    },
  ];

  function withConditionalTooltip(condition, content, children) {
    if (!condition) return children;

    return (
      <Tooltip>
        <TooltipTrigger asChild>{children}</TooltipTrigger>
        <TooltipContent side="right" sideOffset={4}>
          <p>{content}</p>
        </TooltipContent>
      </Tooltip>
    );
  }

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div className="bg-canvas flex h-dvh flex-col overflow-hidden">
      {isWindowSmall ? (
        <nav className="sticky top-0 z-20 w-full">
          <div className="flex flex-row items-center justify-between px-4 py-2">
            <button
              aria-controls="sidepanel"
              type="button"
              className="text-fg-muted hover:bg-hover hover:text-fg rounded-lg
                p-1.5 text-sm"
              onClick={() => setIsOpened(true)}
            >
              <span className="sr-only">Open sidepanel</span>
              <Icon icon={Menu01Icon} styles="size-6" />
            </button>
            <div className="flex flex-row items-center gap-2">
              <a
                className="text-fg-muted hover:bg-hover hover:text-fg flex
                  flex-row items-center gap-2 rounded-lg px-2.5 py-1.5 text-sm
                  font-medium transition-colors"
                href={`${TANGO_URL}/m/${merchantUrlName}`}
              >
                <Icon icon={LinkSquare02Icon} styles="size-4" />
                <span>Live booking page</span>
              </a>
              <button
                aria-label="Switch theme"
                className="text-fg-muted hover:bg-hover hover:text-fg
                  cursor-pointer rounded-lg p-2 transition-colors"
                onClick={switchTheme}
              >
                <Icon
                  icon={Moon02Icon}
                  altIcon={Sun03Icon}
                  showAlt={isDarkTheme}
                  styles="size-4.5"
                />
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
            <div className="flex flex-row items-center gap-2">
              <a
                className="text-fg-muted hover:bg-hover hover:text-fg flex
                  flex-row items-center gap-2 rounded-lg px-2.5 py-1.5 text-sm
                  font-medium transition-colors"
                href={`${TANGO_URL}/m/${merchantUrlName}`}
              >
                <Icon icon={LinkSquare02Icon} styles="size-4" />
                <span>Live booking page</span>
              </a>
              <button
                aria-label="Switch theme"
                className="text-fg-muted hover:bg-hover hover:text-fg
                  cursor-pointer rounded-lg p-2 transition-colors"
                onClick={switchTheme}
              >
                <Icon
                  icon={Moon02Icon}
                  altIcon={Sun03Icon}
                  showAlt={isDarkTheme}
                  styles="size-4.5"
                />
              </button>
            </div>
          </div>
        </header>
      )}
      <div className="flex min-h-0 flex-1 overflow-hidden">
        {isOpen && isWindowSmall && (
          <div
            onClick={closeSidePanelHandler}
            className={`bg-bg-backdrop fixed inset-0 z-20 transition-opacity
            duration-300 ease-in-out
            ${isOpen ? "opacity-100" : "pointer-events-none opacity-0"}`}
          ></div>
        )}
        <aside
          id="sidepanel"
          className={`${isOpen ? "md:translate-x-0" : "-translate-x-full"}
            ${!isWindowSmall && isCollapsed ? "w-16" : "w-60"}
            ${isWindowSmall ? "border-border shadow-modal fixed top-0 border-r" : "relative"}
            bg-canvas left-0 z-30 flex h-full flex-col overflow-hidden
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
            <ol className="flex flex-col space-y-1.5">
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
                        className:
                          "bg-bg-raised! text-fg! shadow-card ring-border ring-1 [&>span:first-child]:text-fg-primary!",
                      }}
                      className={`${isCollapsed && !isWindowSmall ? "h-9 w-10 justify-center" : "h-9 px-2.5"}
                      text-fg-muted hover:bg-hover hover:text-fg flex
                      items-center rounded-lg text-sm font-medium`}
                    >
                      <span className="text-fg-muted shrink-0">
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
                              className="bg-primary-subtle text-fg-primary ms-3
                                inline-flex items-center justify-center
                                rounded-full px-2 py-0.5 text-[11px]
                                font-semibold tracking-wide uppercase"
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
                    hover:bg-hover flex flex-row items-center rounded-lg
                    transition-[padding,gap] duration-300 hover:cursor-pointer`}
                >
                  <Avatar
                    styles="size-9! text-xs! shrink-0"
                    initials={`${user?.first_name.charAt(0)}${user?.last_name.charAt(0)}`}
                  />
                  <div
                    className={`${!isWindowSmall && isCollapsed ? "w-0 opacity-0" : "w-auto opacity-100"}
                      flex flex-1 flex-col items-start whitespace-nowrap`}
                  >
                    <span className="text-sm">{`${user?.first_name} ${user?.last_name}`}</span>
                    <span className="text-fg-muted text-xs capitalize">
                      {role}
                    </span>
                  </div>
                </div>
              </PopoverTrigger>
              <PopoverContent align="start">
                <div
                  className="*:hover:bg-hover space-y-1 text-sm *:cursor-pointer
                    *:rounded-lg *:p-2"
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
                        <Icon icon={LayoutLeftIcon} styles="size-5" />
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
                      <Icon icon={Logout05Icon} styles="size-5" />
                      Sign out
                    </button>
                  </div>
                </div>
              </PopoverContent>
            </Popover>
          </div>
        </aside>
        <div
          className={`flex min-h-0 min-w-0 flex-1 flex-col transition-[margin]
            duration-300 lg:pt-2 lg:pr-2`}
        >
          <div
            className="bg-bg lg:border-border lg:shadow-card flex min-h-0
              min-w-0 flex-1 flex-col overflow-y-auto pb-2 lg:rounded-t-2xl
              lg:border lg:border-b-0"
          >
            <Outlet />
          </div>
        </div>
      </div>
    </div>
  );
}
