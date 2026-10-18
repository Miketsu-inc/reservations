import { InformationCircleIcon } from "@hugeicons/core-free-icons";
import {
  Card,
  Icon,
  Tooltip,
  TooltipContent,
  TooltipTrigger,
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
          <span className="text-fg-muted text-sm whitespace-nowrap">
            {title}
          </span>
          {tooltip && (
            <Tooltip>
              <TooltipTrigger>
                <Icon
                  icon={InformationCircleIcon}
                  styles="size-3.5 text-fg-muted"
                />
              </TooltipTrigger>
              <TooltipContent>
                <p>{tooltipText}</p>
              </TooltipContent>
            </Tooltip>
          )}
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-xl">{text}</span>
          <StatisticsChangeIndicator styles="text-xs" percent={percent} />
        </div>
      </div>
    </Card>
  );
}
