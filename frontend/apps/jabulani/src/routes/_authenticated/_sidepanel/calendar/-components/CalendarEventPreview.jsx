import {
  ArrowReloadHorizontalIcon,
  Calendar02Icon,
  Clock01Icon,
  User03Icon,
} from "@hugeicons/core-free-icons";
import {
  Avatar,
  Icon,
  PreviewCard,
  PreviewCardContent,
  PreviewCardTrigger,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  calendarTeamMembersQueryOptions,
  formatDuration,
  getDisplayPrice,
  timeStringFromDate,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useRef, useState, useSyncExternalStore } from "react";

const STATUS_STYLES = {
  booked:
    "bg-amber-600/20 text-amber-600 dark:bg-amber-600/15 dark:text-amber-400",
  confirmed:
    "bg-blue-600/20 text-blue-600 dark:bg-blue-500/15 dark:text-blue-400",
  completed:
    "bg-green-600/20 text-green-600 dark:bg-green-500/15 dark:text-green-400",
  cancelled: "bg-red-600/20 text-red-600 dark:bg-red-500/15 dark:text-red-400",
  "no-show":
    "bg-gray-600/20 text-gray-600 dark:bg-gray-500/15 dark:text-gray-400",
};

const DEFAULT_PREVIEW_ALIGN_OFFSET = -8;
const PREVIEW_HEADER_GAP = 8;
const activePreviewListeners = new Set();
let activePreviewId = null;
let removeScrollListener = null;

function subscribeToActivePreview(listener) {
  activePreviewListeners.add(listener);

  return () => {
    activePreviewListeners.delete(listener);
  };
}

function getActivePreviewId() {
  return activePreviewId;
}

function getServerActivePreviewId() {
  return null;
}

function setActivePreviewId(value) {
  const nextPreviewId =
    typeof value === "function" ? value(activePreviewId) : value;

  if (nextPreviewId === activePreviewId) return;

  activePreviewId = nextPreviewId;
  removeScrollListener?.();
  removeScrollListener = null;

  if (activePreviewId && typeof window !== "undefined") {
    function closePreviewOnScroll() {
      setActivePreviewId(null);
    }

    window.addEventListener("scroll", closePreviewOnScroll, {
      capture: true,
      once: true,
    });
    removeScrollListener = () => {
      window.removeEventListener("scroll", closePreviewOnScroll, true);
    };
  }

  activePreviewListeners.forEach((listener) => listener());
}

function getPreviewAlignOffset(trigger) {
  if (!trigger) return DEFAULT_PREVIEW_ALIGN_OFFSET;

  let scrollContainer = trigger.parentElement;

  while (scrollContainer) {
    const { overflowY } = getComputedStyle(scrollContainer);

    if (overflowY === "auto" || overflowY === "scroll") break;
    scrollContainer = scrollContainer.parentElement;
  }

  if (!scrollContainer) return DEFAULT_PREVIEW_ALIGN_OFFSET;

  const dayHeader = [
    ...scrollContainer.querySelectorAll('[role="rowgroup"]'),
  ].find((rowGroup) => rowGroup.querySelector('[role="columnheader"]'));

  if (!dayHeader) return DEFAULT_PREVIEW_ALIGN_OFFSET;

  const triggerRect = trigger.getBoundingClientRect();
  const scrollContainerRect = scrollContainer.getBoundingClientRect();
  const dayHeaderRect = dayHeader.getBoundingClientRect();
  const isDayHeaderVisible =
    dayHeaderRect.bottom > scrollContainerRect.top &&
    dayHeaderRect.top < scrollContainerRect.bottom;

  if (!isDayHeaderVisible) return DEFAULT_PREVIEW_ALIGN_OFFSET;

  const minimumPreviewTop = dayHeaderRect.bottom + PREVIEW_HEADER_GAP;

  return Math.max(
    DEFAULT_PREVIEW_ALIGN_OFFSET,
    minimumPreviewTop - triggerRect.top
  );
}

function capitalize(value) {
  if (!value) return "";

  return value.charAt(0).toUpperCase() + value.slice(1).replaceAll("-", " ");
}

function formatDate(date) {
  return date.toLocaleDateString([], {
    weekday: "long",
    month: "long",
    day: "numeric",
  });
}

function getParticipantName(participant) {
  return [participant.first_name, participant.last_name]
    .filter(Boolean)
    .join(" ");
}

function getParticipantInitials(participant) {
  const firstInitial = participant.first_name?.[0] ?? "";
  const lastInitial = participant.last_name?.[0] ?? "";

  return `${firstInitial}${lastInitial}` || "?";
}

function EventLabel({ eventInfo }) {
  const isListView = eventInfo.view.type.startsWith("list");
  const isMonthView = eventInfo.view.type === "dayGridMonth";

  if (isListView) {
    return (
      <div className="flex size-full min-w-0 items-center gap-4">
        {eventInfo.timeText && (
          <span
            className="w-28 shrink-0 text-sm text-gray-500 dark:text-gray-400"
          >
            {eventInfo.timeText}
          </span>
        )}
        <span className="min-w-0 truncate font-medium">
          {eventInfo.event.title}
        </span>
      </div>
    );
  }

  return (
    <div
      className={`flex size-full min-w-0 overflow-hidden
        ${isMonthView ? "flex-row items-center gap-1" : "flex-col"}`}
    >
      {eventInfo.timeText && (
        <span className="shrink-0 text-xs leading-tight font-medium">
          {eventInfo.timeText}
        </span>
      )}
      <span className="min-w-0 truncate text-xs leading-tight font-semibold">
        {eventInfo.event.title}
      </span>
    </div>
  );
}

function ParticipantAvatars({ participants }) {
  if (participants.length === 0) {
    return (
      <div
        className="bg-hvr_gray flex size-10 shrink-0 items-center justify-center
          rounded-full"
      >
        <Icon icon={User03Icon} styles="size-5 text-text_color/60" />
      </div>
    );
  }

  const visibleParticipants = participants.slice(0, 3);
  const remainingParticipants =
    participants.length - visibleParticipants.length;

  return (
    <div className="flex shrink-0 -space-x-4">
      {visibleParticipants.map((participant) => (
        <Avatar
          key={participant.id}
          img={participant.avatar_url}
          initials={getParticipantInitials(participant)}
          alt={getParticipantName(participant)}
          styles="size-10! rounded-full! border-2 border-layer_bg text-xs!"
        />
      ))}
      {remainingParticipants > 0 && (
        <div
          className="border-layer_bg bg-hvr_gray flex size-10 items-center
            justify-center rounded-full border-2 text-xs font-semibold"
        >
          +{remainingParticipants}
        </div>
      )}
    </div>
  );
}

function DetailRow({ icon, children }) {
  return (
    <div className="text-text_color/70 flex items-center gap-2.5 text-sm">
      <Icon icon={icon} styles="size-4 shrink-0" />
      <span>{children}</span>
    </div>
  );
}

function TeamMemberAvatars({ members }) {
  const visibleMembers = members.slice(0, 4);
  const remainingMembers = members.length - visibleMembers.length;

  return (
    <div className="flex shrink-0 -space-x-3" role="list">
      {visibleMembers.map((member) => {
        const name = [member.first_name, member.last_name]
          .filter(Boolean)
          .join(" ");
        const initials = `${member.first_name?.[0] ?? ""}${
          member.last_name?.[0] ?? ""
        }`;

        return (
          <span
            key={member.id}
            className="rounded-full"
            role="listitem"
            title={name || "Team member"}
          >
            <Avatar
              img={member.avatar_url}
              initials={initials || "?"}
              alt={name}
              styles="size-8! rounded-full! border-2 border-layer_bg
                text-[11px]!"
            />
          </span>
        );
      })}
      {remainingMembers > 0 && (
        <div
          className="border-layer_bg bg-hvr_gray flex size-8 items-center
            justify-center rounded-full border-2 text-[11px] font-semibold"
          role="listitem"
        >
          +{remainingMembers}
        </div>
      )}
    </div>
  );
}

function TeamMemberAvatarSkeleton({ count }) {
  return (
    <div
      aria-label="Loading assigned team"
      className="flex shrink-0 -space-x-3"
    >
      {Array.from({ length: Math.min(count, 4) }, (_, index) => (
        <div
          key={index}
          className="border-layer_bg bg-hvr_gray size-8 animate-pulse
            rounded-full border-2"
        />
      ))}
    </div>
  );
}

function AssignedTeam({ employeeIds, isPending, teamMembers }) {
  if (employeeIds.length === 0) return null;

  const assignedMembers = employeeIds.map(
    (employeeId) =>
      teamMembers.find(
        (member) => String(member.id) === String(employeeId)
      ) ?? {
        id: employeeId,
        first_name: null,
        last_name: null,
      }
  );

  return (
    <div className="border-border_color border-t px-4 py-3.5">
      <div className="flex items-center gap-3">
        {isPending ? (
          <TeamMemberAvatarSkeleton count={employeeIds.length} />
        ) : (
          <TeamMemberAvatars members={assignedMembers} />
        )}
        <span className="text-sm font-medium">Team members</span>
      </div>
    </div>
  );
}

function BookingPreview({ event, timeFormat }) {
  const { extendedProps } = event;
  const participants = extendedProps.participants ?? [];
  const isGroupBooking = extendedProps.booking_type !== "appointment";
  const customerName =
    participants.length > 0
      ? getParticipantName(participants[0]) || "Customer"
      : "Walk-in";
  const additionalParticipants = participants.length - 1;
  const status = extendedProps.booking_status;
  const groupSummary = `${participants.length} of ${extendedProps.max_participants} participants`;

  return (
    <div className="w-80 max-w-[calc(100vw-2rem)] overflow-hidden">
      <div className="flex items-center gap-3 p-4">
        <ParticipantAvatars participants={participants} />
        <div className="min-w-0 flex-1">
          <p className="truncate font-semibold">
            {customerName}
            {isGroupBooking && additionalParticipants > 0
              ? ` +${additionalParticipants}`
              : ""}
          </p>
          <div className="mt-1 flex min-w-0 items-center gap-2">
            <span
              className={`${STATUS_STYLES[status] ?? STATUS_STYLES["no-show"]}
                shrink-0 rounded-md px-1.5 py-0.5 text-[11px] font-medium`}
            >
              {capitalize(status)}
            </span>
            {isGroupBooking && (
              <span className="text-text_color/55 truncate text-xs">
                {groupSummary}
              </span>
            )}
          </div>
        </div>
      </div>

      <div className="border-border_color border-t px-4 py-3">
        <div className="flex items-center justify-between gap-4">
          <div className="flex min-w-0 items-center gap-2">
            <span
              className="size-2.5 shrink-0 rounded-full"
              style={{ backgroundColor: extendedProps.service_color }}
            />
            <span className="truncate text-sm font-semibold">
              {extendedProps.service_name}
            </span>
          </div>
          <span className="shrink-0 text-sm font-semibold">
            {getDisplayPrice(extendedProps.price, extendedProps.price_type)}
          </span>
        </div>
      </div>

      <div className="border-border_color flex flex-col gap-2.5 border-t p-4">
        <DetailRow icon={Calendar02Icon}>{formatDate(event.start)}</DetailRow>
        <DetailRow icon={Clock01Icon}>
          {timeStringFromDate(event.start, timeFormat)} -{" "}
          {timeStringFromDate(event.end, timeFormat)}
          <span className="text-text_color/45">
            {" "}
            · {formatDuration(extendedProps.duration)}
          </span>
        </DetailRow>
        {extendedProps.is_recurring && (
          <DetailRow icon={ArrowReloadHorizontalIcon}>
            Part of a repeating series
          </DetailRow>
        )}
      </div>
    </div>
  );
}

function BlockedTimePreview({ event, timeFormat }) {
  const { extendedProps } = event;
  const { merchantId } = useAuth();
  const employeeIds = extendedProps.employee_ids ?? [];
  const { data: teamMembers = [], isPending } = useQuery({
    ...calendarTeamMembersQueryOptions(merchantId),
    enabled: employeeIds.length > 0,
  });

  return (
    <div className="w-72 max-w-[calc(100vw-2rem)] overflow-hidden">
      <div className="flex items-center gap-3 p-4">
        {extendedProps.icon && (
          <div
            className="bg-hvr_gray flex size-10 shrink-0 items-center
              justify-center rounded-full text-lg"
          >
            {extendedProps.icon}
          </div>
        )}
        <div className="min-w-0">
          <p className="truncate font-semibold">{extendedProps.name}</p>
          <p className="text-text_color/60 mt-0.5 text-xs">Blocked time</p>
        </div>
      </div>

      <div className="border-border_color flex flex-col gap-2.5 border-t p-4">
        <DetailRow icon={Calendar02Icon}>{formatDate(event.start)}</DetailRow>
        <DetailRow icon={Clock01Icon}>
          {extendedProps.allDay
            ? "All day"
            : `${timeStringFromDate(event.start, timeFormat)} - ${timeStringFromDate(event.end, timeFormat)}`}
        </DetailRow>
      </div>
      <AssignedTeam
        employeeIds={employeeIds}
        isPending={isPending}
        teamMembers={teamMembers}
      />
    </div>
  );
}

export default function CalendarEventPreview({ eventInfo, timeFormat }) {
  const currentActivePreviewId = useSyncExternalStore(
    subscribeToActivePreview,
    getActivePreviewId,
    getServerActivePreviewId
  );
  const isBlockedTime = eventInfo.event.extendedProps.type === "blocked";
  const previewId = `${isBlockedTime ? "blocked" : "booking"}:${eventInfo.event.id}`;
  const isOpen = currentActivePreviewId === previewId;
  const [alignOffset, setAlignOffset] = useState(DEFAULT_PREVIEW_ALIGN_OFFSET);
  const suppressOpenRef = useRef(false);

  useEffect(() => {
    return () => {
      setActivePreviewId((currentPreviewId) =>
        currentPreviewId === previewId ? null : currentPreviewId
      );
    };
  }, [previewId]);

  function handleOpenChange(open, eventDetails) {
    if (open && suppressOpenRef.current) return;

    if (open) {
      setAlignOffset(getPreviewAlignOffset(eventDetails?.trigger));
      setActivePreviewId(previewId);
      return;
    }

    setActivePreviewId((currentPreviewId) =>
      currentPreviewId === previewId ? null : currentPreviewId
    );
  }

  function dismissPreview() {
    suppressOpenRef.current = true;
    setActivePreviewId((currentPreviewId) =>
      currentPreviewId === previewId ? null : currentPreviewId
    );
  }

  return (
    <PreviewCard open={isOpen} onOpenChange={handleOpenChange}>
      <PreviewCardTrigger asChild delay={350} closeDelay={120}>
        <div
          className="size-full min-w-0 cursor-pointer"
          onClick={dismissPreview}
          onPointerDown={dismissPreview}
          onPointerLeave={() => {
            suppressOpenRef.current = false;
          }}
        >
          <EventLabel eventInfo={eventInfo} />
        </div>
      </PreviewCardTrigger>
      <PreviewCardContent
        align="start"
        alignOffset={alignOffset}
        side="right"
        styles="overflow-hidden"
        sideOffset={12}
        collisionPadding={12}
      >
        {isBlockedTime ? (
          <BlockedTimePreview event={eventInfo.event} timeFormat={timeFormat} />
        ) : (
          <BookingPreview event={eventInfo.event} timeFormat={timeFormat} />
        )}
      </PreviewCardContent>
    </PreviewCard>
  );
}
