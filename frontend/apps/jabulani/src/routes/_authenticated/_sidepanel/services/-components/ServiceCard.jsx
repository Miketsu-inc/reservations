import { PopoverClose } from "@radix-ui/react-popover";
import {
  ArrowIcon,
  ClockIcon,
  EditIcon,
  ThreeDotsIcon,
  TrashBinIcon,
} from "@reservations/assets";
import {
  Button,
  Popover,
  PopoverContent,
  PopoverTrigger,
  Switch,
} from "@reservations/components";
import {
  formatDuration,
  invalidateLocalStorageAuth,
  useToast,
} from "@reservations/lib";

export default function ServiceCard({
  isWindowSmall,
  service,
  serviceCount,
  onDelete,
  onEdit,
  refresh,
  onMoveBack,
  onMoveForth,
}) {
  const { showToast } = useToast();

  async function serviceStatusHandler(isActive) {
    const response = await fetch(
      `/api/v1/merchants/services/${service.id}/${isActive ? "activate" : "deactivate"}`,
      {
        method: "POST",
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
      <div className="border-border_color bg-layer_bg w-sm rounded-r-lg border">
        <div
          className="border-border_color relative z-5 flex flex-row
            justify-between border-b p-4"
        >
          <div className="flex flex-row gap-4">
            <div
              style={{ backgroundColor: service.color }}
              className="flex size-[70px] shrink-0 overflow-hidden rounded-lg"
            >
              <img
                className="size-full object-cover"
                src="https://dummyimage.com/70x70/d156c3/000000.jpg"
                alt="service photo"
              />
            </div>
            <div className="flex flex-col justify-center gap-2">
              <p className="flex-wrap font-semibold">{service.name}</p>
              <div className="flex flex-row items-center gap-2">
                <span
                  className="fill-gray-400 dark:fill-gray-500"
                  style={{
                    fill: service.is_active ? service.color : undefined,
                  }}
                >
                  <ClockIcon styles="size-4" />
                </span>
                <p className="text-sm">
                  {formatDuration(service.total_duration)}
                </p>
              </div>
            </div>
          </div>
          <Popover>
            <PopoverTrigger asChild>
              <button
                className="hover:bg-hvr_gray hover:*:stroke-text_color h-fit
                  cursor-pointer rounded-lg p-1"
              >
                <ThreeDotsIcon
                  styles="size-6 stroke-4 stroke-gray-400 dark:stroke-gray-500"
                />
              </button>
            </PopoverTrigger>
            <PopoverContent side="left">
              <div
                className="flex flex-col items-start *:flex *:w-full *:flex-row
                  *:items-center *:rounded-lg *:p-2"
              >
                <div className="flex flex-row items-center gap-3">
                  <Switch
                    onSwitch={serviceStatusHandler}
                    defaultValue={service.is_active}
                  />
                  <p>Active</p>
                </div>
                <PopoverClose asChild>
                  <button
                    disabled={service.sequence === serviceCount}
                    onClick={() => onMoveBack(service.id)}
                    className={`${
                      service.sequence === serviceCount
                        ? "opacity-35"
                        : "hover:bg-hvr_gray cursor-pointer"
                      } gap-5`}
                  >
                    <ArrowIcon styles="size-6 ml-2 -rotate-90 stroke-current" />
                    <p>Move back</p>
                  </button>
                </PopoverClose>
                <PopoverClose asChild>
                  <button
                    disabled={service.sequence === 1}
                    onClick={() => onMoveForth(service.id)}
                    className={`${service.sequence === 1 ? "opacity-35" : "hover:bg-hvr_gray cursor-pointer"}
                      gap-5`}
                  >
                    <ArrowIcon styles="size-6 rotate-90 stroke-current ml-2" />
                    <p>Move forth</p>
                  </button>
                </PopoverClose>
              </div>
            </PopoverContent>
          </Popover>
        </div>
        <div
          className="relative z-5 flex flex-row items-center justify-between
            gap-4 p-4"
        >
          <Button
            type="button"
            styles="py-2 px-4"
            variant="danger"
            buttonText={`${!isWindowSmall ? "Delete" : ""}`}
            onClick={onDelete}
          >
            <TrashBinIcon styles="size-5 stroke-white! mr-1 mb-0.5" />
          </Button>
          <Button
            type="button"
            styles="py-2 px-4 flex-1"
            variant="primary"
            buttonText="Edit"
            onClick={onEdit}
          >
            <EditIcon styles="size-4 mr-2" />
          </Button>
        </div>
      </div>
    </div>
  );
}
