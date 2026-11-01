import {
  ArrowReloadHorizontalIcon,
  CalendarBlock02Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import { AvatarStack, Icon } from "@reservations/components";
import {
  formatDuration,
  getBookingStatusStyles,
  timeStringFromDate,
  toAvatar,
} from "@reservations/lib";

function GroupParticipants({ extendedProps }) {
  return (
    <span className="flex min-w-0 items-center gap-1">
      <Icon icon={UserGroupIcon} styles="size-3.5 shrink-0" />
      <span className="truncate">
        {extendedProps.participants.length}/{extendedProps.max_participants}
      </span>
    </span>
  );
}

function getCustomerName({ participants }) {
  if (participants.length === 0) return "Walk-in";

  return toAvatar(participants[0]).name;
}

function formatTimeRange(event, timeFormat) {
  if (event.allDay) return "All day";

  return `${timeStringFromDate(event.start, timeFormat)} - ${timeStringFromDate(event.end, timeFormat)}`;
}

export default function AgendaEventContent({
  eventInfo,
  membersById,
  timeFormat,
}) {
  const { event } = eventInfo;
  const { extendedProps } = event;
  const isBlockedTime = extendedProps.type === "blocked";
  const isGroup =
    !isBlockedTime && extendedProps.booking_type !== "appointment";
  const employeeIds = isBlockedTime
    ? extendedProps.employee_ids
    : [extendedProps.employee_id];
  const assignees = membersById
    ? employeeIds.map((id) => membersById.get(id)).filter(Boolean)
    : [];

  return (
    <div className="flex min-w-0 flex-col gap-0.5">
      <div className="flex min-w-0 items-center justify-between gap-2">
        <span className="truncate font-semibold">
          {isBlockedTime ? event.title : extendedProps.service_name}
        </span>
        {!isBlockedTime && (
          <span
            className={`${getBookingStatusStyles(extendedProps.booking_status)}
            shrink-0 rounded-md px-1.5 py-0.5 text-[11px] font-medium
            capitalize`}
          >
            {extendedProps.booking_status}
          </span>
        )}
      </div>
      <span className="text-fg-muted flex items-center gap-1.5 text-sm">
        {formatTimeRange(event, timeFormat)}
        {!isBlockedTime && ` · ${formatDuration(extendedProps.duration)}`}
        {extendedProps.is_recurring && (
          <Icon
            icon={ArrowReloadHorizontalIcon}
            styles="size-3.5 shrink-0"
            aria-label="Repeating booking"
          />
        )}
      </span>
      <div
        className="text-fg-muted flex min-w-0 items-center justify-between gap-2
          text-sm"
      >
        {isBlockedTime ? (
          <span className="flex min-w-0 shrink-0 items-center gap-1">
            <Icon icon={CalendarBlock02Icon} styles="size-3.5 shrink-0" />
            <span className="truncate">Blocked time</span>
          </span>
        ) : isGroup ? (
          <GroupParticipants extendedProps={extendedProps} />
        ) : (
          <span className="truncate">{getCustomerName(extendedProps)}</span>
        )}
        <AvatarStack
          avatars={assignees.map((member) => toAvatar(member))}
          size="sm"
        />
      </div>
    </div>
  );
}
