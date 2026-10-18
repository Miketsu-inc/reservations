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
      className={`shadow-card relative flex h-fit max-w-full flex-row
        overflow-hidden rounded-xl ${service.is_active ? "" : "opacity-60"}`}
    >
      <div
        style={{
          backgroundColor: service.is_active ? service.color : undefined,
        }}
        className="bg-bg-inactive w-1.5 shrink-0 rounded-l-xl"
      ></div>
      <div
        className="bg-bg-raised hover:border-border-strong border-border w-full
          min-w-0 rounded-r-xl border"
      >
        <div className="flex flex-row items-center py-3 pr-2 pl-4">
          <Link
            to={`/services/edit/${service.id}`}
            className="flex flex-1 cursor-pointer lg:cursor-default"
            disabled={!isWindowSmall}
          >
            <div className="flex flex-row gap-4">
              <div className="flex flex-col justify-center gap-1">
                <p className="truncate text-sm font-semibold">{service.name}</p>
                <div
                  className="text-fg-muted flex min-h-6 flex-row items-center
                    gap-1.5 text-sm"
                >
                  <Icon icon={Clock01Icon} styles="size-3.5 text-fg-muted" />
                  <p>{formatDuration(service.total_duration)}</p>
                  {service.booking_type !== "appointment" && (
                    <div
                      className="bg-hover text-fg-muted ml-2 flex w-fit flex-row
                        items-center gap-1 rounded-md px-1.5 py-0.5 text-xs
                        font-medium"
                    >
                      <Icon icon={UserGroupIcon} styles="size-3.5" />
                      <p>Group</p>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </Link>
          <div className="flex flex-row items-center gap-0.5 text-sm">
            <Link
              className="text-fg-muted hover:bg-hover hover:text-fg hidden
                cursor-pointer rounded-lg p-2 lg:block"
              to={`/services/edit/${service.id}`}
            >
              <Icon icon={Edit03Icon} styles="size-4.5" />
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
          className="text-fg-muted hover:bg-hover hover:text-fg h-fit
            cursor-pointer rounded-lg p-2"
        >
          <Icon icon={MoreVerticalIcon} styles="size-4.5 rotate-90" />
        </button>
      </PopoverTrigger>
      <PopoverContent side="left">
        <div
          className="flex flex-col items-start *:flex *:w-full *:flex-row
            *:items-center *:gap-2.5 *:rounded-lg *:p-2"
        >
          <div className="flex flex-row items-center">
            <Switch
              size="small"
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
                  : "hover:bg-hover cursor-pointer"
                }`}
            >
              <Icon
                icon={ArrowLeft02Icon}
                styles="size-4.5 -rotate-90 text-fg-muted"
              />
              <p>Move down</p>
            </button>
          </PopoverClose>
          <PopoverClose asChild>
            <button
              disabled={service.sequence === 1}
              onClick={onMoveUp}
              className={`${service.sequence === 1 ? "opacity-35" : "hover:bg-hover cursor-pointer"}`}
            >
              <Icon
                icon={ArrowLeft02Icon}
                styles="size-4.5 rotate-90 text-fg-muted"
              />
              <p>Move up</p>
            </button>
          </PopoverClose>
          <PopoverClose asChild>
            <button
              onClick={onDelete}
              className="hover:bg-hover text-fg-danger cursor-pointer"
            >
              <Icon icon={Delete02Icon} styles="size-4.5" />
              <p>Delete</p>
            </button>
          </PopoverClose>
        </div>
      </PopoverContent>
    </Popover>
  );
}
