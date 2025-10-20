import { InfoIcon } from "@reservations/assets";
import {
  Card,
  TooltipContent,
  TooltipTrigger,
  Tootlip,
} from "@reservations/components";
import StatisticsChangeIndicator from "./StatisticsChangeIndicator";

export default function StatisticsCard({
  title,
  text,
  percent,
  tooltip,
  tooltipText,
}) {
  return (
    <Card>
      <div className="flex h-full flex-col gap-3">
        <div className="flex flex-row items-center gap-1">
          <span className="text-sm whitespace-nowrap">{title}</span>
          {tooltip && (
            <Tootlip>
              <TooltipTrigger>
                <InfoIcon styles="size-4 stroke-gray-500 dark:stroke-gray-400" />
              </TooltipTrigger>
              <TooltipContent>
                <p>{tooltipText}</p>
              </TooltipContent>
            </Tootlip>
          )}
        </div>
        <div className="flex flex-col">
          <span className="text-lg font-medium">{text}</span>
          <StatisticsChangeIndicator styles="text-xs" percent={percent} />
        </div>
      </div>
    </Card>
  );
}
