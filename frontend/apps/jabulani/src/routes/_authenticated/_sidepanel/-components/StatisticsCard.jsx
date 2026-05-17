import { InformationCircleIcon } from "@hugeicons/core-free-icons";
import {
  Card,
  Icon,
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
                <Icon
                  icon={InformationCircleIcon}
                  styles="size-4 text-gray-500 dark:text-gray-400"
                />
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
