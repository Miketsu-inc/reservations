import { Avatar } from "@reservations/components";
import { UnassignedAvatar } from "./CalendarMemberFilter";

export function ResourceHeader({ resource }) {
  return (
    <div className="flex min-w-0 flex-col items-center gap-1">
      {resource.extendedProps.isUnassigned ? (
        <UnassignedAvatar styles="size-8" />
      ) : (
        <Avatar
          styles="size-8! text-[11px]!"
          img={resource.extendedProps.avatarUrl}
          initials={resource.extendedProps.initials}
          alt={resource.title}
        />
      )}
      <span className="text-fg max-w-full truncate">{resource.title}</span>
    </div>
  );
}

export function DayHeader({ info }) {
  return info.textParts.map((part, index) =>
    part.type === "day" ? (
      <span
        key={index}
        className={`mx-0.5 flex items-center justify-center text-sm
          font-semibold ${
            info.isToday
              ? "bg-primary text-fg-on-primary size-6 rounded-full"
              : "text-fg "
          }`}
      >
        {part.value}
      </span>
    ) : (
      <span key={index} className="text-fg-muted text-sm whitespace-pre">
        {part.value}
      </span>
    )
  );
}

export function ListDayHeader({ info }) {
  const month = info.textParts.find((part) => part.type === "month")?.value;

  return (
    <>
      <span
        className={`shrink-0 pr-1 text-xl font-semibold tabular-nums
          ${info.isToday ? "text-fg-primary" : ""}`}
      >
        {info.dayNumberText}
      </span>
      <span
        className={`font-semibold capitalize
          ${info.isToday ? "text-fg-primary" : ""}`}
      >
        {info.weekdayText}
      </span>
      <span className="text-fg-muted text-sm">{month}</span>
    </>
  );
}
