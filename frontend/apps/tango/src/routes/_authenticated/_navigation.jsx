import {
  BackArrowIcon,
  CalendarIcon,
  CustomersIcon,
  DashboardIcon,
  MoonIcon,
  SearchIcon,
  SettingsIcon,
  SignOutIcon,
  SunIcon,
} from "@reservations/assets";
import {
  Avatar,
  Loading,
  Popover,
  PopoverContent,
  PopoverTrigger,
  ServerError,
} from "@reservations/components";
import { meQueryOptions, useTheme, useWindowSize } from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, Outlet } from "@tanstack/react-router";
import { useCallback } from "react";

export const Route = createFileRoute("/_authenticated/_navigation")({
  component: NavLayout,
  loader: async ({ context: { queryClient } }) => {
    await queryClient.ensureQueryData(meQueryOptions());
  },
  pendingComponent: Loading,
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function NavLayout() {
  const navigate = Route.useNavigate();
  const windowSize = useWindowSize();
  const isWindowSmall = windowSize === "sm" || windowSize === "md";

  const { data: user, isLoading } = useQuery(meQueryOptions());

  const { isDarkTheme, switchTheme } = useTheme();

  const handleLogout = useCallback(async () => {
    const response = await fetch("/api/v1/auth/logout", {
      method: "POST",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    });

    if (response.ok) {
      navigate({
        from: Route.fullPath,
        to: "/",
      });
    }
  }, [navigate]);

  const navigation = [
    {
      href: "/home",
      label: "Home",
      icon: <DashboardIcon styles="size-5" />,
    },
    {
      href: "/profile",
      label: "Profile",
      icon: <CalendarIcon styles="size-5" />,
    },
    {
      href: "/bookings",
      label: "Bookings",
      icon: <CalendarIcon styles="size-5" />,
    },
    {
      href: "/favorites",
      label: "Favorites",
      icon: <CustomersIcon styles="size-5" />,
    },
    {
      href: "/settings",
      label: "Settings",
      icon: <SettingsIcon styles="size-5" />,
    },
  ];

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div className="bg-layer_bg flex h-dvh flex-col overflow-hidden">
      {isWindowSmall ? (
        <div className="sticky top-0 z-20 w-full">
          <div className="flex flex-row items-center px-4 py-4">
            <Link to=".." className="cursor-pointer">
              <BackArrowIcon styles="size-6 stroke-text_color" />
            </Link>
          </div>
        </div>
      ) : (
        <header className={"py-2 pl-4 transition-[margin] duration-300"}>
          <div className="mr-4 flex flex-row items-center justify-between">
            <div className="flex flex-row items-center gap-4">
              <div
                className={`flex h-10 w-40 flex-row items-center gap-3
                  transition-normal duration-300 ease-in-out`}
              >
                <img
                  className="h-full rounded-lg object-cover"
                  src="https://dummyimage.com/160x40/d156c3/000000.jpg"
                />
              </div>
            </div>
            <div className="flex flex-row gap-4">
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
        {!isWindowSmall && (
          <aside
            id="sidepanel"
            className={`bg-layer_bg relative left-0 z-30 flex h-full w-60
            flex-col overflow-hidden transition-[width,translate] duration-300
            md:translate-x-0`}
            aria-label="Sidepanel"
          >
            {isWindowSmall && (
              <div className="shrink-0 p-4">
                <div
                  className={`flex h-10 w-40 flex-row items-center gap-3
                  transition-normal duration-300 ease-in-out`}
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
                    <Link
                      to={item.href}
                      activeProps={{
                        className: "bg-primary/20 *:text-primary! *:duration-0",
                      }}
                      className={`text-text_color hover:bg-primary/20 flex h-10
                      items-center rounded-lg px-2`}
                    >
                      <span
                        className="shrink-0 text-gray-500 transition duration-75
                          dark:text-gray-400"
                      >
                        {item.icon}
                      </span>
                      {!isWindowSmall && (
                        <>
                          <span
                            className={`ms-3 flex-1 whitespace-nowrap
                            transition-[opacity,width] duration-300`}
                          >
                            {item.label}
                          </span>
                        </>
                      )}
                    </Link>
                  </li>
                ))}
              </ol>
              <Popover>
                <PopoverTrigger>
                  <div
                    className={`hover:bg-primary/20 flex flex-row items-center
                    gap-2 rounded-lg p-2 transition-[padding,gap] duration-300
                    hover:cursor-pointer`}
                  >
                    <Avatar
                      styles="size-10! text-sm! shrink-0"
                      initials={`${user?.first_name.charAt(0)}${user?.last_name.charAt(0)}`}
                    />
                    <div
                      className={`flex flex-1 flex-col items-start gap-0.5
                      whitespace-nowrap`}
                    >
                      <span className="text-sm">{`${user?.first_name} ${user?.last_name}`}</span>
                    </div>
                  </div>
                </PopoverTrigger>
                <PopoverContent align="start">
                  <div
                    className="*:hover:bg-primary/20 space-y-2 text-sm
                      *:cursor-pointer *:rounded-lg *:p-2"
                  >
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
        )}
        <div
          className="flex min-h-0 flex-1 flex-col transition-[margin]
            duration-300 md:pt-2 md:pr-2"
        >
          <div
            id="page-scroll-container"
            className="bg-bg_color flex-1 overflow-y-auto rounded-t-xl px-4 pt-4
              pb-32 md:pb-12"
          >
            <Outlet />
          </div>
        </div>
      </div>
      {isWindowSmall && (
        <nav
          className="border-border_color bg-layer_bg fixed right-0 bottom-0
            left-0 z-50 border-t"
          style={{
            paddingBottom: "env(safe-area-inset-bottom)",
          }}
        >
          <div
            className="relative mx-auto grid max-w-md
              grid-cols-[1fr_1fr_auto_1fr_1fr] items-center p-2"
          >
            <MobileNavLink from={Route.fullPath} to="/home">
              <DashboardIcon styles="size-8" />
              <p className="text-sm">Home</p>
            </MobileNavLink>
            <MobileNavLink from={Route.fullPath} to="/bookings">
              <CalendarIcon styles="size-8" />
              <p className="text-sm">Bookings</p>
            </MobileNavLink>
            <div className="relative flex w-9 justify-center">
              <button
                className="bg-primary border-bg_color absolute -top-16 flex
                  size-18 items-center justify-center rounded-full border-4"
              >
                <SearchIcon styles="size-8 text-white" />
              </button>
            </div>
            <MobileNavLink from={Route.fullPath} to="/favorites">
              <DashboardIcon styles="size-8" />
              <p className="text-sm">Favorites</p>
            </MobileNavLink>
            <MobileNavLink from={Route.fullPath} to="/profile">
              <Avatar
                styles="size-8! text-xs"
                initials={`${user.first_name[0]}${user.last_name[0]}`}
              />
              <p className="text-sm">Profile</p>
            </MobileNavLink>
          </div>
        </nav>
      )}
    </div>
  );
}

function MobileNavLink({ children, from, to }) {
  return (
    <Link
      from={from}
      to={to}
      activeProps={{
        className: "bg-primary/20 *:text-primary!",
      }}
      className="flex justify-center rounded-lg"
    >
      <span className="flex flex-col items-center gap-1 py-2">{children}</span>
    </Link>
  );
}
