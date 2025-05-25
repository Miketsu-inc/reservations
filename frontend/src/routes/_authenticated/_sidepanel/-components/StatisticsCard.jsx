import Card from "@components/Card";
import InfoIcon from "@icons/InfoIcon";
import StatisticsChangeIndicator from "./StatisticsChangeIndicator";

export default function StatisticsCard({ title, text, percent }) {
  return (
    <Card>
      <div className="flex h-full flex-col gap-3">
        <div className="flex flex-row items-center gap-1">
          <span className="text-sm whitespace-nowrap">{title}</span>
          <InfoIcon styles="size-4 stroke-gray-500 dark:stroke-gray-400" />
        </div>
        <div className="flex flex-col">
          <span className="text-lg font-medium">{text}</span>
          <StatisticsChangeIndicator styles="text-xs" percent={percent} />
        </div>
      </div>
    </Card>
  );
}
