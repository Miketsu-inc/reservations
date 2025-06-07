import Button from "@components/Button";
import ClockIcon from "@icons/ClockIcon";
import EditIcon from "@icons/EditIcon";
import ThreeDotsIcon from "@icons/ThreeDotsIcon";
import TrashBinIcon from "@icons/TrashBinIcon";
import { formatDuration } from "@lib/datetime";

export default function ServiceCard({ service, onDelete, onEdit }) {
  return (
    <div className="relative flex h-fit max-w-full flex-row rounded-lg shadow-sm">
      <div
        style={{ backgroundColor: service.color }}
        className="w-2 rounded-l-lg"
      ></div>
      <div
        className="absolute inset-0 z-0 opacity-0 dark:opacity-10"
        style={{
          background: `linear-gradient(90deg, ${service.color} 0%, ${service.color}30 30%, transparent 70%)`,
        }}
      />
      <div className="border-border_color bg-layer_bg w-sm rounded-r-lg border">
        <div className="border-border_color relative z-[5] flex flex-row justify-between border-b p-4">
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
                <span style={{ fill: service.color }}>
                  <ClockIcon styles="size-4" />
                </span>
                <p className="text-sm">
                  {formatDuration(service.total_duration)}
                </p>
                {!service.is_active && (
                  <span
                    className="w-fit rounded-full bg-red-500/20 px-2 py-1 text-xs text-red-500
                      dark:bg-red-700/20 dark:text-red-700"
                  >
                    Inactive
                  </span>
                )}
              </div>
            </div>
          </div>
          <ThreeDotsIcon
            styles="size-8 stroke-4 stroke-gray-400 dark:stroke-gray-500 hover:bg-hvr_gray
              rounded-lg cursor-pointer hover:stroke-text_color"
          />
        </div>
        <div className="relative z-[5] flex flex-row items-center justify-between gap-4 p-4">
          <Button
            type="button"
            styles="py-2 px-4"
            variant="danger"
            buttonText="Delete"
            onClick={onDelete}
          >
            <TrashBinIcon styles="size-5 !stroke-white mr-1 mb-0.5" />
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
