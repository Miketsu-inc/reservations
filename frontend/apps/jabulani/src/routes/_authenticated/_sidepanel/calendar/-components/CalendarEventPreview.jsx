import {
  ArrowReloadHorizontalIcon,
  Calendar02Icon,
  Clock01Icon,
  User03Icon,
} from "@hugeicons/core-free-icons";
import {
  AvatarStack,
  Icon,
  PreviewCard,
  PreviewCardContent,
  PreviewCardTrigger,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  calendarTeamMembersQueryOptions,
  formatDuration,
  getBookingStatusStyles,
  getDisplayPrice,
  timeStringFromDate,
  toAvatar,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import {
  useCallback,
  useEffect,
  useId,
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

function setActivePreviewId(nextPreviewId) {
  if (nextPreviewId === activePreviewId) return;

  const previousPreviewId = activePreviewId;
  activePreviewId = nextPreviewId;

  [previousPreviewId, activePreviewId].forEach((previewId) => {
    activePreviewListeners.get(previewId)?.forEach((listener) => listener());
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

function formatDate(date) {
  return date.toLocaleDateString([], {
    weekday: "long",
    month: "long",
    day: "numeric",
  });
}

function ParticipantAvatars({ participants }) {
  if (participants.length === 0) {
    return (
      <div
        className="bg-hover flex size-10 shrink-0 items-center justify-center
          rounded-full"
      >
        <Icon icon={User03Icon} styles="size-5 text-fg-muted" />
      </div>
    );
  }

  return (
    <AvatarStack
      avatars={participants.map((person) => toAvatar(person, "Customer"))}
      size="lg"
    />
  );
}

function DetailRow({ icon, children }) {
  return (
    <div className="text-fg-muted flex items-center gap-2.5 text-sm">
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
          className="border-bg-raised bg-hover size-8 animate-pulse rounded-full
            border-2"
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
    <div className="border-border border-t px-4 py-3.5">
      <div className="flex items-center gap-3">
        {isPending ? (
          <TeamMemberAvatarSkeleton count={employeeIds.length} />
        ) : (
          <AvatarStack
            avatars={assignedMembers.map((member) =>
              toAvatar(member, "Team member")
            )}
            max={4}
            size="md"
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
      ? toAvatar(participants[0], "Customer").name
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
              className={`${getBookingStatusStyles(status)} shrink-0 rounded-md
                px-1.5 py-0.5 text-[11px] font-medium`}
            >
              {status}
            </span>
            {isGroupBooking && (
              <span className="text-fg-muted truncate text-xs">
                {groupSummary}
              </span>
            )}
          </div>
        </div>
      </div>

      <div className="border-border border-t px-4 py-3">
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

      <div className="border-border flex flex-col gap-2.5 border-t p-4">
        <DetailRow icon={Calendar02Icon}>{formatDate(event.start)}</DetailRow>
        <DetailRow icon={Clock01Icon}>
          {timeStringFromDate(event.start, timeFormat)} -{" "}
          {timeStringFromDate(event.end, timeFormat)}
          <span className="text-fg-muted">
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
      <div className="p-4">
        <p className="truncate font-semibold">
          {extendedProps.name} {extendedProps.icon}
        </p>
        <p className="text-fg-muted mt-0.5 text-xs">Blocked time</p>
      </div>

      <div className="border-border flex flex-col gap-2.5 border-t p-4">
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

export default function CalendarEventPreview({
  event,
  timeFormat,
  children,
  canOpen,
}) {
  const isBlockedTime = event.extendedProps.type === "blocked";
  const previewId = useId();
  const isOpen = usePreviewOpen(previewId);
  const [alignOffset, setAlignOffset] = useState(DEFAULT_PREVIEW_ALIGN_OFFSET);
  const suppressOpenRef = useRef(false);

  useEffect(() => {
    return () => closePreview(previewId);
  }, [previewId]);

  function handleOpenChange(open, eventDetails) {
    if (open && (suppressOpenRef.current || canOpen?.() === false)) return;

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
          {children}
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
            event={event}
            isOpen={isOpen}
            timeFormat={timeFormat}
          />
        ) : (
          <BookingPreview event={event} timeFormat={timeFormat} />
        )}
      </PreviewCardContent>
    </PreviewCard>
  );
}
