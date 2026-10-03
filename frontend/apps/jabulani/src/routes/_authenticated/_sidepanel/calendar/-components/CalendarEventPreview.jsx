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
import {
  getBookingStatusStyles,
  useAuth,
} from "@reservations/jabulani/lib";
import {
  calendarTeamMembersQueryOptions,
  formatDuration,
  getDisplayPrice,
  timeStringFromDate,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import {
  useCallback,
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";

const DEFAULT_PREVIEW_ALIGN_OFFSET = -8;
const PREVIEW_HEADER_GAP = 8;
const activePreviewListeners = new Map();
let activePreviewId = null;

function subscribeToPreview(previewId, listener) {
  const listeners = activePreviewListeners.get(previewId) ?? new Set();
  listeners.add(listener);
  activePreviewListeners.set(previewId, listeners);

  return () => {
    listeners.delete(listener);
    if (listeners.size === 0) activePreviewListeners.delete(previewId);
  };
}

function closeActivePreview() {
  setActivePreviewId(null);
}

function setActivePreviewId(nextPreviewId) {
  if (nextPreviewId === activePreviewId) return;

  const previousPreviewId = activePreviewId;
  activePreviewId = nextPreviewId;

  if (typeof window !== "undefined") {
    window.removeEventListener("scroll", closeActivePreview, true);

    if (activePreviewId) {
      window.addEventListener("scroll", closeActivePreview, {
        capture: true,
        once: true,
      });
    }
  }

  [previousPreviewId, activePreviewId].forEach((previewId) => {
    activePreviewListeners
      .get(previewId)
      ?.forEach((listener) => listener());
  });
}

function closePreview(previewId) {
  if (activePreviewId === previewId) setActivePreviewId(null);
}

function usePreviewOpen(previewId) {
  const subscribe = useCallback(
    (listener) => subscribeToPreview(previewId, listener),
    [previewId]
  );
  const getSnapshot = useCallback(
    () => activePreviewId === previewId,
    [previewId]
  );

  return useSyncExternalStore(subscribe, getSnapshot, () => false);
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

function getPersonName(person) {
  return [person.first_name, person.last_name].filter(Boolean).join(" ");
}

function getPersonInitials(person) {
  const firstInitial = person.first_name?.[0] ?? "";
  const lastInitial = person.last_name?.[0] ?? "";

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

function AvatarStack({
  people,
  limit,
  styles,
  avatarStyles,
  overflowStyles,
  fallbackTitle,
}) {
  const visiblePeople = people.slice(0, limit);
  const remainingPeople = people.length - visiblePeople.length;

  return (
    <div className={`flex shrink-0 ${styles}`} role="list">
      {visiblePeople.map((person) => {
        const name = getPersonName(person);

        return (
          <span
            key={person.id}
            className="rounded-full"
            role="listitem"
            title={name || fallbackTitle}
          >
            <Avatar
              img={person.avatar_url}
              initials={getPersonInitials(person)}
              alt={name}
              styles={avatarStyles}
            />
          </span>
        );
      })}
      {remainingPeople > 0 && (
        <div
          className={`${overflowStyles} border-layer_bg bg-hvr_gray flex
            items-center justify-center rounded-full border-2 font-semibold`}
          role="listitem"
        >
          +{remainingPeople}
        </div>
      )}
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

  return (
    <AvatarStack
      people={participants}
      limit={3}
      styles="-space-x-4"
      avatarStyles="size-10! rounded-full! border-2 border-layer_bg text-xs!"
      overflowStyles="size-10 text-xs"
      fallbackTitle="Customer"
    />
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

  const teamMembersById = new Map(
    teamMembers.map((member) => [String(member.id), member])
  );
  const assignedMembers = employeeIds.map(
    (employeeId) =>
      teamMembersById.get(String(employeeId)) ?? {
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
          <AvatarStack
            people={assignedMembers}
            limit={4}
            styles="-space-x-3"
            avatarStyles="size-8! rounded-full! border-2 border-layer_bg
              text-[11px]!"
            overflowStyles="size-8 text-[11px]"
            fallbackTitle="Team member"
          />
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
      ? getPersonName(participants[0]) || "Customer"
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
              className={`${getBookingStatusStyles(status)} shrink-0
                rounded-md px-1.5 py-0.5 text-[11px] font-medium`}
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

function BlockedTimePreview({ event, isOpen, timeFormat }) {
  const { extendedProps } = event;
  const { merchantId } = useAuth();
  const employeeIds = extendedProps.employee_ids ?? [];
  const { data: teamMembers = [], isPending } = useQuery({
    ...calendarTeamMembersQueryOptions(merchantId),
    enabled: isOpen && employeeIds.length > 0,
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
  const isBlockedTime = eventInfo.event.extendedProps.type === "blocked";
  const previewId = `${isBlockedTime ? "blocked" : "booking"}:${eventInfo.event.id}`;
  const isOpen = usePreviewOpen(previewId);
  const [alignOffset, setAlignOffset] = useState(DEFAULT_PREVIEW_ALIGN_OFFSET);
  const suppressOpenRef = useRef(false);

  useEffect(() => {
    return () => closePreview(previewId);
  }, [previewId]);

  function handleOpenChange(open, eventDetails) {
    if (open && suppressOpenRef.current) return;

    if (open) {
      setAlignOffset(getPreviewAlignOffset(eventDetails?.trigger));
      setActivePreviewId(previewId);
      return;
    }

    closePreview(previewId);
  }

  function dismissPreview() {
    suppressOpenRef.current = true;
    closePreview(previewId);
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
          <BlockedTimePreview
            event={eventInfo.event}
            isOpen={isOpen}
            timeFormat={timeFormat}
          />
        ) : (
          <BookingPreview event={eventInfo.event} timeFormat={timeFormat} />
        )}
      </PreviewCardContent>
    </PreviewCard>
  );
}
