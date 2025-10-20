import { HourGlassIcon, ServicesIcon } from "@reservations/assets";
import { formatDuration } from "@reservations/lib";

export default function PhaseItem({ phase, isLast }) {
  return (
    <div className="relative">
      <div className="flex w-full items-center gap-4">
        <div
          className={`flex size-10 items-center justify-center rounded-full
            border-2 ${
              phase.phase_type === "active"
                ? "border-secondary"
                : "border-yellow-500"
            }`}
        >
          {phase.phase_type === "wait" ? (
            <HourGlassIcon styles="size-5 stroke-yellow-500" />
          ) : (
            <ServicesIcon styles="size-5 text-secondary" />
          )}
        </div>

        <div className="flex w-full flex-1 items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="text-text_color font-medium">
              {phase.name || `Phase ${phase.sequence}`}
            </span>
            <span
              className={`rounded-full px-2 py-1 text-xs ${
                phase.phase_type === "active"
                  ? "text-secondary bg-secondary/20"
                  : "bg-yellow-500/20 text-yellow-500"
                }`}
            >
              {phase.phase_type}
            </span>
          </div>
          <span className="text-text_color">
            {formatDuration(phase.duration)}
          </span>
        </div>
      </div>

      {!isLast && (
        <div
          className="absolute -bottom-[29px] left-[19px] h-5 w-[2px] bg-gray-400
            dark:bg-gray-300"
        />
      )}
    </div>
  );
}
