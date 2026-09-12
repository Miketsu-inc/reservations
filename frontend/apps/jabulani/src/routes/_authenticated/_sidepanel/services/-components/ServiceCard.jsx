import {
  ArrowLeft02Icon,
  Clock01Icon,
  Delete02Icon,
  Edit03Icon,
  MoreVerticalIcon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import {
  Icon,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
  Switch,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  formatDuration,
  invalidateLocalStorageAuth,
  useToast,
  useWindowSize,
} from "@reservations/lib";
import { Link } from "@tanstack/react-router";

export default function ServiceCard({
  service,
  serviceCount,
  onDelete,
  refresh,
  onMove,
}) {
  const { isWindowSmall } = useWindowSize();
  const { showToast } = useToast();
  const { merchantId } = useAuth();

  async function serviceStatusHandler(isActive) {
    const response = await fetch(
      `/api/v1/merchants/${merchantId}/services/${service.id}/${isActive ? "activate" : "deactivate"}`,
      {
        method: "PATCH",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
      }
    );

    if (!response.ok) {
      const result = await response.json();
      invalidateLocalStorageAuth(response.status);
      showToast({
        variant: "error",
        message: `Something went wrong while ${isActive ? "activating" : "deactivating"} the service ${result.error}`,
      });
    } else {
      refresh();
    }
  }

  return (
    <div
      className="relative flex h-fit max-w-full flex-row rounded-lg shadow-sm"
    >
      <div
        style={{
          backgroundColor: service.is_active ? service.color : undefined,
        }}
        className="w-2 shrink-0 rounded-l-lg bg-gray-400 dark:bg-gray-500"
      ></div>
      <div
        className={`${service.is_active ? "dark:opacity-10" : ""} absolute
          inset-0 z-0 opacity-0`}
        style={{
          background: `linear-gradient(90deg, ${service.color} 0%, ${service.color}30 30%, transparent 70%)`,
        }}
      />
      <div
        className="border-border_color bg-layer_bg w-full rounded-r-lg border"
      >
        <div className="relative z-5 flex flex-row items-center p-4">
          <Link
            to={`/services/edit/${service.id}`}
            className="flex flex-1 cursor-pointer lg:cursor-default"
            disabled={!isWindowSmall}
          >
            <div className="flex flex-row gap-4">
              <div className="flex flex-col justify-center gap-2">
                <p className="truncate font-semibold">{service.name}</p>
                <div
                  className="text-text_color/80 flex min-h-8 flex-row
                    items-center gap-2 text-sm"
                >
                  <Icon icon={Clock01Icon} styles="size-4" />
                  <p>{formatDuration(service.total_duration)}</p>
                  {service.booking_type !== "appointment" && (
                    <div
                      className="border-border_color bg-bg_color
                        text-text_color/60 ml-2 flex w-fit flex-row items-center
                        gap-1 rounded-lg border px-2 py-1 text-sm"
                    >
                      <Icon icon={UserGroupIcon} styles="size-4" />
                      <p>Group</p>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </Link>
          <div className="flex flex-row items-center gap-2 text-base">
            <Link
              className="hover:bg-hvr_gray hidden cursor-pointer rounded-lg p-2
                lg:block"
              to={`/services/edit/${service.id}`}
            >
              <Icon
                icon={Edit03Icon}
                styles="size-6 text-text_color/40 dark:text-text_color/50"
              />
            </Link>
            <ServiceOptions
              service={service}
              serviceCount={serviceCount}
              onActiveSwitch={serviceStatusHandler}
              onMoveDown={() => onMove(service.id, "backward")}
              onMoveUp={() => onMove(service.id, "forward")}
              onDelete={onDelete}
            />
          </div>
        </div>
      </div>
    </div>
  );
}

function ServiceOptions({
  service,
  serviceCount,
  onMoveDown,
  onMoveUp,
  onActiveSwitch,
  onDelete,
}) {
  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          className="hover:bg-hvr_gray hover:*:stroke-text_color h-fit
            cursor-pointer rounded-lg p-1"
        >
          <Icon
            icon={MoreVerticalIcon}
            styles="size-8 text-text_color/40 dark:text-text_color/50 rotate-90"
          />
        </button>
      </PopoverTrigger>
      <PopoverContent side="left">
        <div
          className="flex flex-col items-start *:flex *:w-full *:flex-row
            *:items-center *:gap-5 *:rounded-lg *:p-2"
        >
          <div className="flex flex-row items-center gap-3!">
            <Switch
              onSwitch={onActiveSwitch}
              defaultValue={service.is_active}
            />
            <p>Active</p>
          </div>
          <PopoverClose asChild>
            <button
              disabled={service.sequence === serviceCount}
              onClick={onMoveDown}
              className={`${
                service.sequence === serviceCount
                  ? "opacity-35"
                  : "hover:bg-hvr_gray cursor-pointer"
                }`}
            >
              <Icon icon={ArrowLeft02Icon} styles="size-6 -rotate-90 ml-1.5" />
              <p className="ml-0.5">Move down</p>
            </button>
          </PopoverClose>
          <PopoverClose asChild>
            <button
              disabled={service.sequence === 1}
              onClick={onMoveUp}
              className={`${service.sequence === 1 ? "opacity-35" : "hover:bg-hvr_gray cursor-pointer"}`}
            >
              <Icon icon={ArrowLeft02Icon} styles="size-6 rotate-90 ml-1.5" />
              <p className="ml-0.5">Move up</p>
            </button>
          </PopoverClose>
          <PopoverClose asChild>
            <button
              onClick={onDelete}
              className="hover:bg-hvr_gray cursor-pointer text-red-600
                dark:text-red-500"
            >
              <Icon icon={Delete02Icon} styles="size-6 ml-1.5 mb-0.5" />
              <p className="ml-0.5">Delete</p>
            </button>
          </PopoverClose>
        </div>
      </PopoverContent>
    </Popover>
  );
}
