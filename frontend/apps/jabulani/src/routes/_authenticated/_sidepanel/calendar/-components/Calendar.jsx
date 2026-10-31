import { Calendar as FullCalendar } from "@fullcalendar/react";
import interactionPlugin from "@fullcalendar/react/interaction";
import listPlugin from "@fullcalendar/react/list";
import { ArrowLeft01Icon } from "@hugeicons/core-free-icons";
import {
  Avatar,
  Button,
  DatePicker,
  Icon,
  Select,
  ServerError,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  businessHoursQueryOptions,
  calendarTeamMembersQueryOptions,
  dateStringToLocalDate,
  DEFAULT_SERVICE_COLOR,
  formatToDateString,
  preferencesQueryOptions,
  useWindowSize,
} from "@reservations/lib";
import {
  keepPreviousData,
  useQuery,
  useSuspenseQueries,
} from "@tanstack/react-query";
import { useParams } from "@tanstack/react-router";
import {
  useCallback,
  useEffect,
  useEffectEvent,
  useMemo,
  useRef,
  useState,
} from "react";
import AgendaEventContent from "./AgendaEventContent";
import CalendarEventContent from "./CalendarEventContent";
import CalendarEventPreview from "./CalendarEventPreview";
import CalendarMemberFilter from "./CalendarMemberFilter";
import {
  calendarBookingQueryOptions,
  calendarBookingsQueryOptions,
} from "./calendarQueries";
import CalendarSidePanel from "./CalendarSidePanel";
import calendarTheme from "./calendarTheme";
import {
  canViewTeamCalendar,
  getCalendarRange,
  getCalendarRanges,
  getCalendarView,
  hasTeamColumns,
  resolveCalendarRange,
} from "./calendarViews";
import CreateMenu, { CalendarCreateMenu } from "./CreateMenu";
import resourceTimeGridPlugin from "./resourceTimeGridPlugin";

import "@fullcalendar/react/skeleton.css";
import "@fullcalendar/react/themes/breezy/theme.css";

const timeGridViews = new Set([
  "resourceTimeGridDay",
  "resourceTimeGrid2Day",
  "resourceTimeGridWeek",
]);

function getContrastColor(color) {
  const hex = color.replace("#", "");

  const red = parseInt(hex.substring(0, 2), 16);
  const green = parseInt(hex.substring(2, 4), 16);
  const blue = parseInt(hex.substring(4, 6), 16);

  const [r, g, b] = [red, green, blue].map((channel) => {
    const value = channel / 255;
    return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
  });
  const luminance = r * 0.2126 + g * 0.7152 + b * 0.0722;
  return luminance > 0.179 ? "#000000" : "#ffffff";
}

function transformBusinessHours(businessHours) {
  if (!businessHours) return {};

  return Object.entries(businessHours).map(([day, times]) => ({
    daysOfWeek: [parseInt(day)],
    startTime: times.start_time.slice(0, 5),
    endTime: times.end_time.slice(0, 5),
  }));
}

function formatBookings(data) {
  if (data === undefined) return;

  return data.map((booking) => {
    let title = "Walk-in";
    const isGroup = booking.booking_type !== "appointment";

    if (isGroup) {
      title = `${booking.service_name} - ${booking.participants.length}/${booking.max_participants}`;
    } else if (booking.participants.length > 0) {
      const participant = booking.participants[0];
      title = `${participant.first_name} ${participant.last_name}`;
    }

    const serviceColor = booking.service_color ?? DEFAULT_SERVICE_COLOR;

    return {
      id: booking.id,
      resourceId: booking.employee_id,
      title: title,
      start: booking.from_date,
      end: booking.to_date,
      color: serviceColor,
      contrastColor: getContrastColor(serviceColor),
      durationEditable: false,
      startEditable: new Date(booking.to_date) > new Date() ? true : false,
      extendedProps: {
        // this is a number unlike the normal 'id' which get's converted to a string
        id: booking.id,
        event_type: "booking",
        booking_type: booking.booking_type,
        booking_status: booking.booking_status,
        is_recurring: booking.is_recurring,
        max_participants: booking.max_participants,
        participants: booking.participants,
        merchant_note: booking.merchant_note,
        service_name: booking.service_name,
        service_color: serviceColor,
        service_id: booking.service_id,
        employee_id: booking.employee_id,
        duration: booking.duration,
        phases: booking.phases ?? [],
        price: booking.price,
        price_type: booking.price_type,
      },
    };
  });
}

function formatBlockedTimes(data) {
  if (data === undefined) return;

  return data.map((blockedTime) => {
    const start = blockedTime.is_all_day
      ? blockedTime.blocked_day
      : blockedTime.from_date;
    const end = blockedTime.is_all_day ? undefined : blockedTime.to_date;
    const editableUntil = blockedTime.is_all_day
      ? new Date(`${blockedTime.blocked_day}T23:59:59`)
      : new Date(blockedTime.to_date);

    return {
      id: blockedTime.id,
      resourceIds: blockedTime.employee_ids,
      title: `${blockedTime.name} ${blockedTime?.icon || ""}`,
      start: start,
      end: end,
      color: "#6b6b78",
      contrastColor: getContrastColor("#6b6b78"),
      durationEditable: !blockedTime.is_all_day,
      allDay: blockedTime.is_all_day,
      startEditable: editableUntil > new Date(),
      extendedProps: {
        id: blockedTime.id,
        type: "blocked",
        name: blockedTime.name,
        icon: blockedTime.icon,
        blocked_type_id: blockedTime.blocked_type_id,
        employee_ids: blockedTime.employee_ids,
        allDay: blockedTime.is_all_day,
      },
    };
  });
}

// Matches the column order of the team views. Defined once, so the query keeps
// returning the same array between renders.
function sortMembersByName(members) {
  return members.toSorted((a, b) =>
    `${a.first_name} ${a.last_name}`.localeCompare(
      `${b.first_name} ${b.last_name}`
    )
  );
}

function formatResources(teamMembers) {
  return teamMembers.map((member) => ({
    id: member.id,
    title: `${member.first_name} ${member.last_name}`,
    avatarUrl: member.avatar_url,
    initials: `${member.first_name[0]}${member.last_name[0]}`,
  }));
}

function ResourceHeader({ resource }) {
  return (
    <div className="flex min-w-0 flex-col items-center gap-1">
      <Avatar
        styles="size-8! text-[11px]!"
        img={resource.extendedProps.avatarUrl}
        initials={resource.extendedProps.initials}
        alt={resource.title}
      />
      <span className="text-fg max-w-full truncate">{resource.title}</span>
    </div>
  );
}

function ListDayHeader({ info }) {
  const month = info.textParts.find((part) => part.type === "month")?.value;

  return (
    <>
      <span
        className={`shrink-0 pr-1 text-xl font-semibold tabular-nums
          ${info.isToday ? "text-fg-primary" : ""}`}
      >
        {info.date.getDate()}
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

/**
 * Dropping an event into another employee's column only moves its resource,
 * so mirror the change into the employee fields the side panels read.
 */
function applyDroppedEmployee({ event, oldResource, newResource }) {
  if (!newResource) return;

  const oldEmployeeId = Number(oldResource.id);
  const newEmployeeId = Number(newResource.id);

  if (event.extendedProps.type === "blocked") {
    const employeeIds = event.extendedProps.employee_ids.map((employeeId) =>
      employeeId === oldEmployeeId ? newEmployeeId : employeeId
    );
    event.setExtendedProp("employee_ids", [...new Set(employeeIds)]);
    return;
  }

  event.setExtendedProp("employee_id", newEmployeeId);
}

const defaultSidePanelState = {
  isOpen: false,
  type: null,
  data: null,
  initialStart: null,
  initialEmployeeId: null,
  initialAllDay: false,
  bookingId: null,
  panelKey: null,
};

function createBookingPanelState(bookingId, event = null) {
  return {
    isOpen: true,
    type: "edit-booking",
    data: event,
    initialStart: null,
    initialEmployeeId: null,
    initialAllDay: false,
    bookingId,
    panelKey: `edit-booking:${bookingId}`,
  };
}

function getCalendarCreateStart(date) {
  const now = new Date();
  const calendarDate = new Date(date);

  calendarDate.setHours(now.getHours(), now.getMinutes(), 0, 0);
  return calendarDate < now ? now : calendarDate;
}

function getInitialCalendarDate(search) {
  return dateStringToLocalDate(search.start) ?? new Date();
}

function findScroller(el, child) {
  let scroller = child;
  while (scroller && scroller !== el) {
    if (["auto", "scroll"].includes(getComputedStyle(scroller).overflowY))
      return scroller;
    scroller = scroller.parentElement;
  }
  return null;
}

// The header is sticky, so measure its day instead.
function scrollToDay(header) {
  const day = header.parentElement;
  const scroller = findScroller(document.body, day);
  if (!scroller) return;

  scroller.scrollTop +=
    day.getBoundingClientRect().top - scroller.getBoundingClientRect().top;
}

/**
 * Centers the current time in day/week views on mount and when the grid resizes.
 * Measures the visible grid height to offset FullCalendar's top-aligned scrolling.
 * Returns methods to recenter on date changes and clean up when the view unmounts.
 */
function observeCurrentTime({ el, view }) {
  if (!timeGridViews.has(view.type)) return;

  const scroller = findScroller(el, el.querySelector("[data-time]"));
  if (!scroller) return;

  // Runs synchronously, since a view switch paints the new view before the
  // next animation frame. FullCalendar keeps the requested time and reapplies
  // it once the slot heights are measured.
  function center() {
    if (!scroller.clientHeight || !scroller.scrollHeight) return;

    const now = new Date();
    const currentMinutes =
      now.getHours() * 60 + now.getMinutes() + now.getSeconds() / 60;
    const visibleMinutes =
      (scroller.clientHeight / scroller.scrollHeight) * 24 * 60;

    view.calendar.scrollToTime({
      minutes: Math.max(0, currentMinutes - visibleMinutes / 2),
    });
  }

  const observer = new ResizeObserver(center);
  observer.observe(scroller);
  observer.observe(scroller.firstElementChild);

  return {
    center,
    disconnect() {
      observer.disconnect();
    },
  };
}

export default function Calendar({ router, route, search }) {
  const { bookingId } = useParams({ strict: false });
  const [sidePanelState, setSidePanelState] = useState(() =>
    bookingId ? createBookingPanelState(bookingId) : defaultSidePanelState
  );
  const [calendarTitle, setCalendarTitle] = useState("");
  const [calendarDate, setCalendarDate] = useState(() =>
    getInitialCalendarDate(search)
  );
  const [createSelection, setCreateSelection] = useState(null);

  // Keep the old booking only when the route disappears, so its panel can
  // animate out before the transition callback clears it.
  if (bookingId && sidePanelState.bookingId !== bookingId) {
    setSidePanelState(createBookingPanelState(bookingId));
  }

  const { merchantId, employeeId, role } = useAuth();
  const canViewTeam = canViewTeamCalendar(role);
  const { queryClient } = route.useRouteContext({ from: route.id });
  const {
    data: events = { bookings: [], blocked_times: [] },
    isError,
    error,
  } = useQuery({
    ...calendarBookingsQueryOptions(merchantId, search.start, search.end),
    placeholderData: keepPreviousData,
  });
  const { data: selectedBooking } = useQuery({
    ...calendarBookingQueryOptions(merchantId, sidePanelState.bookingId),
    enabled: Boolean(sidePanelState.bookingId),
  });
  const { data: teamMembers = [] } = useQuery({
    ...calendarTeamMembersQueryOptions(merchantId),
    enabled: canViewTeam,
    select: sortMembersByName,
  });
  const [{ data: preferences }, { data: businessHours }] = useSuspenseQueries({
    queries: [
      preferencesQueryOptions(merchantId, employeeId),
      businessHoursQueryOptions(merchantId),
    ],
  });
  // null shows everyone, so new team members and unassigned bookings show up.
  const [memberFilter, setMemberFilter] = useState(null);
  // Kept in state rather than read from the URL, so the resources update in
  // the same render as the view instead of after the navigation.
  const [selectedRange, setSelectedRange] = useState(search.view);

  // The agenda takes over where the side navigation collapses.
  const { isWindowSmall } = useWindowSize();
  const calendarRanges = getCalendarRanges(role, isWindowSmall);
  const calendarRef = useRef();
  const calendarScrollRef = useRef();
  const dragRevertRef = useRef(null);
  const isDraggingRef = useRef(false);

  // Drops members that have since left the team.
  const filteredEmployeeIds = useMemo(() => {
    if (!canViewTeam) return [employeeId];
    if (!memberFilter) return null;

    const ids = teamMembers
      .map((member) => member.id)
      .filter((id) => memberFilter.includes(id));
    return ids.length > 0 ? ids : null;
  }, [canViewTeam, employeeId, memberFilter, teamMembers]);

  const membersById = useMemo(
    () =>
      canViewTeam
        ? new Map(teamMembers.map((member) => [member.id, member]))
        : null,
    [canViewTeam, teamMembers]
  );

  const selectedEmployeeIds = useMemo(
    () => filteredEmployeeIds ?? teamMembers.map((member) => member.id),
    [filteredEmployeeIds, teamMembers]
  );

  const calendarEvents = useMemo(() => {
    const filter = filteredEmployeeIds && new Set(filteredEmployeeIds);
    const bookings = filter
      ? events.bookings.filter((booking) => filter.has(booking.employee_id))
      : events.bookings;
    const blockedTimes = filter
      ? events.blocked_times.filter((blockedTime) =>
          blockedTime.employee_ids.some((id) => filter.has(id))
        )
      : events.blocked_times;

    return [...formatBookings(bookings), ...formatBlockedTimes(blockedTimes)];
  }, [filteredEmployeeIds, events.bookings, events.blocked_times]);

  const calendarRange = resolveCalendarRange(
    calendarRanges,
    selectedRange,
    preferences.calendar_view
  );
  // Crossing the mobile breakpoint can switch to the agenda.
  const resolvedView = getCalendarView(calendarRange);
  const showTeamColumns =
    hasTeamColumns(calendarRange) && selectedEmployeeIds.length > 1;

  // Without resources the time grid merges everyone into one column per day.
  const resources = useMemo(() => {
    if (!showTeamColumns) return [];

    const selected = new Set(selectedEmployeeIds);
    return formatResources(
      teamMembers.filter((member) => selected.has(member.id))
    );
  }, [showTeamColumns, teamMembers, selectedEmployeeIds]);

  const selectedBookingEvent = useMemo(() => {
    if (!selectedBooking) return null;

    const event = formatBookings([selectedBooking])[0];
    return {
      ...event,
      start: new Date(event.start),
      end: new Date(event.end),
    };
  }, [selectedBooking]);

  const panelData = sidePanelState.bookingId
    ? (sidePanelState.data ?? selectedBookingEvent)
    : sidePanelState.data;

  const isPanelOpen = sidePanelState.bookingId
    ? Boolean(bookingId && panelData)
    : sidePanelState.isOpen;

  const panelType =
    sidePanelState.type === "edit-booking" && !panelData
      ? null
      : sidePanelState.type;

  const invalidateBookingsQuery = useCallback(async () => {
    await queryClient.invalidateQueries(
      calendarBookingsQueryOptions(merchantId, search.start, search.end)
    );
  }, [merchantId, queryClient, search]);

  const invalidateSelectedBookingQuery = useCallback(async () => {
    if (!sidePanelState.bookingId) return;

    await queryClient.invalidateQueries(
      calendarBookingQueryOptions(merchantId, sidePanelState.bookingId)
    );
  }, [sidePanelState.bookingId, merchantId, queryClient]);

  const closeCreateMenu = useCallback(() => {
    setCreateSelection(null);
  }, []);

  function openSidePanel(
    type,
    data,
    id,
    {
      revert = null,
      initialStart = null,
      initialEmployeeId = null,
      initialAllDay = false,
    } = {}
  ) {
    dragRevertRef.current = revert;
    setCreateSelection(null);

    setSidePanelState({
      isOpen: true,
      type,
      data,
      initialStart,
      initialEmployeeId,
      initialAllDay,
      bookingId: null,
      panelKey: `${type}:${id}`,
    });
  }

  function openBookingSidePanel(event, revert = null) {
    dragRevertRef.current = revert;
    setCreateSelection(null);

    setSidePanelState(createBookingPanelState(event.id, event));

    router.navigate({
      to: "/calendar/bookings/$bookingId",
      params: { bookingId: event.id },
      search,
    });
  }

  function openSidePanelAtTime(
    type,
    {
      start = getCalendarCreateStart(calendarRef.current.getApi().getDate()),
      employeeId = null,
      allDay = false,
    } = {}
  ) {
    openSidePanel(type, null, start.toISOString(), {
      initialStart: start,
      initialEmployeeId: employeeId,
      initialAllDay: allDay,
    });
  }

  function closeSidePanel() {
    dragRevertRef.current?.();
    dragRevertRef.current = null;

    if (sidePanelState.bookingId) {
      router.navigate({ to: "/calendar", search });
      return;
    }

    setSidePanelState((prev) => ({
      ...prev,
      isOpen: false,
    }));
  }

  function saveSidePanel() {
    dragRevertRef.current = null;
    invalidateBookingsQuery();

    if (sidePanelState.bookingId) {
      invalidateSelectedBookingQuery();
      router.navigate({ to: "/calendar", search });
      return;
    }

    setSidePanelState((prev) => ({
      ...prev,
      isOpen: false,
    }));
  }

  function finishSidePanelClose() {
    if (sidePanelState.bookingId) {
      dragRevertRef.current?.();
      dragRevertRef.current = null;
    }

    setSidePanelState(defaultSidePanelState);
  }

  const datesChanged = useCallback(
    (api) => {
      const start = formatToDateString(api.view.activeStart);
      const end = formatToDateString(api.view.activeEnd);

      router.navigate({
        to: "/calendar",
        search: (prev) => ({
          ...prev,
          view: getCalendarRange(api.view.type),
          start: start,
          end: end,
        }),
        replace: true,
      });
    },
    [router]
  );

  function navButtonHandler(dir) {
    const api = calendarRef.current.getApi();

    if (dir === "prev") {
      api.prev();
    } else if (dir === "next") {
      api.next();
    }

    datesChanged(api);
  }

  function todayButtonHandler() {
    const api = calendarRef.current.getApi();

    const today = new Date().getTime();
    if (
      api.view.currentStart.getTime() <= today &&
      api.view.currentEnd.getTime() >= today
    ) {
      return;
    }

    api.today();

    datesChanged(api);
  }

  function changeViewHandler(view) {
    const api = calendarRef.current.getApi();

    if (view === api.view.type) return;

    // Otherwise switching from a week to a shorter range lands on its first day.
    const now = new Date();
    const showsToday = api.view.activeStart <= now && now < api.view.activeEnd;
    api.changeView(view, showsToday ? now : api.view.currentStart);

    datesChanged(api);
  }

  const syncCalendarView = useEffectEvent(() => {
    changeViewHandler(resolvedView);
  });

  useEffect(() => {
    syncCalendarView();
  }, [resolvedView]);

  function changeRangeHandler(range) {
    setSelectedRange(range);
    changeViewHandler(getCalendarView(range));
  }

  function changeMembersHandler(employeeIds) {
    setMemberFilter(
      employeeIds.length === teamMembers.length ? null : employeeIds
    );
  }

  function changeDateHandler(date) {
    const api = calendarRef.current.getApi();
    const dateStr = formatToDateString(date);

    api.gotoDate(dateStr);
    datesChanged(api);
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col px-4 pt-4">
      <CalendarSidePanel
        isOpen={isPanelOpen}
        type={panelType}
        data={panelData}
        initialStart={sidePanelState.initialStart}
        initialEmployeeId={sidePanelState.initialEmployeeId}
        initialAllDay={sidePanelState.initialAllDay}
        panelKey={sidePanelState.panelKey}
        onClose={closeSidePanel}
        onTransitionEnd={finishSidePanelClose}
        onSave={saveSidePanel}
        onSoftUpdate={() => {
          invalidateBookingsQuery();
          invalidateSelectedBookingQuery();
        }}
        preferences={preferences}
      />
      {createSelection && (
        <CalendarCreateMenu
          selection={createSelection}
          onClose={closeCreateMenu}
          onCreateBlockedTime={() =>
            openSidePanelAtTime("blocked-time", createSelection)
          }
          onCreateBooking={() =>
            openSidePanelAtTime("new-booking", {
              ...createSelection,
              // A booking needs a time, so the all-day row starts it now.
              start: createSelection.allDay
                ? getCalendarCreateStart(createSelection.start)
                : createSelection.start,
            })
          }
        />
      )}
      {/* Mobile puts the title on its own row above the controls. */}
      <div className="flex shrink-0 flex-wrap items-center gap-x-4 gap-y-2 pb-4">
        <p
          className="w-full text-xl font-semibold whitespace-nowrap md:order-2
            md:w-auto md:text-2xl"
        >
          {calendarTitle}
        </p>
        <div className="flex flex-row items-center gap-1 md:order-1">
          <DatePicker
            styles="w-fit"
            value={calendarDate}
            hideText={true}
            firstDayOfWeek={preferences.first_day_of_week}
            onSelect={changeDateHandler}
          />
          <button
            className="text-fg-muted hover:bg-hover hover:text-fg cursor-pointer
              rounded-lg p-2"
            type="button"
            onClick={() => navButtonHandler("prev")}
          >
            <Icon icon={ArrowLeft01Icon} styles="size-5" />
          </button>
          <Button
            variant="secondary"
            styles="px-3 py-2 text-sm"
            buttonText="Today"
            onClick={todayButtonHandler}
          />
          <button
            className="text-fg-muted hover:bg-hover hover:text-fg cursor-pointer
              rounded-lg p-2"
            type="button"
            onClick={() => navButtonHandler("next")}
          >
            <Icon icon={ArrowLeft01Icon} styles="size-5 rotate-180" />
          </button>
        </div>
        <div
          className="ml-auto flex min-w-0 flex-1 flex-row items-center
            justify-end gap-2 md:order-3 md:flex-none"
        >
          {canViewTeam && (
            <CalendarMemberFilter
              members={teamMembers}
              selectedIds={selectedEmployeeIds}
              onChange={changeMembersHandler}
              currentEmployeeId={employeeId}
            />
          )}
          {calendarRanges.length > 1 && (
            <Select
              options={calendarRanges}
              value={calendarRange}
              onSelect={(option) => changeRangeHandler(option.value)}
              styles="w-28!"
            />
          )}
          <CreateMenu
            isFloating={isWindowSmall}
            onCreateBlockedTime={() => openSidePanelAtTime("blocked-time")}
            onCreateBooking={() => openSidePanelAtTime("new-booking")}
          />
        </div>
      </div>
      <div className="min-h-0 w-full flex-1 overflow-hidden">
        <FullCalendar
          ref={calendarRef}
          plugins={[
            calendarTheme,
            interactionPlugin,
            resourceTimeGridPlugin,
            listPlugin,
          ]}
          resources={resources}
          resourceOrder="title"
          resourceDayHeaderContent={(info) => (
            <ResourceHeader resource={info.resource} />
          )}
          locale="hu"
          editable={true}
          eventDurationEditable={true}
          selectable={true}
          initialView={resolvedView}
          initialDate={getInitialCalendarDate(search)}
          height="100%"
          // Lets the agenda sit on the page instead of a white card.
          className={
            resolvedView === "listWeek" ? "bg-transparent!" : undefined
          }
          headerToolbar={false}
          viewDidMount={(info) => {
            calendarScrollRef.current = observeCurrentTime(info);
          }}
          viewWillUnmount={() => {
            calendarScrollRef.current?.disconnect();
            calendarScrollRef.current = null;
          }}
          events={calendarEvents}
          blockEventClass="calendar-event bg-(--fc-event-color)!"
          blockEventInnerClass="text-(--fc-event-contrast-color)!"
          eventContent={(eventInfo) =>
            eventInfo.view.type === "listWeek" ? (
              // The agenda already shows everything the preview would.
              <AgendaEventContent
                eventInfo={eventInfo}
                membersById={membersById}
                timeFormat={preferences.time_format}
              />
            ) : (
              <CalendarEventPreview
                event={eventInfo.event}
                timeFormat={preferences.time_format}
                canOpen={() =>
                  !eventInfo.isMirror &&
                  !eventInfo.isDragging &&
                  !isDraggingRef.current
                }
              >
                <CalendarEventContent eventInfo={eventInfo} />
              </CalendarEventPreview>
            )
          }
          datesSet={({ view }) => {
            calendarScrollRef.current?.center();
            setCalendarTitle(view.title);

            const currentDate = view.calendar.getDate();
            setCalendarDate((previousDate) =>
              previousDate.getTime() === currentDate.getTime()
                ? previousDate
                : new Date(currentDate)
            );
          }}
          eventClick={(e) => {
            const type = e.event.extendedProps.type;

            if (type === "blocked") {
              openSidePanel("blocked-time", e.event, e.event.id);
              return;
            }

            openBookingSidePanel(e.event);
          }}
          dateClick={(info) => {
            // The all-day row stays clickable until the day is over.
            const end = new Date(info.date);
            if (info.allDay) end.setDate(end.getDate() + 1);

            if (
              !timeGridViews.has(info.view.type) ||
              end.getTime() < Date.now() ||
              info.jsEvent.target.closest(".fc-event")
            ) {
              return;
            }

            setCreateSelection({
              start: info.date,
              allDay: info.allDay,
              employeeId: info.resource ? Number(info.resource.id) : null,
              x: info.jsEvent.clientX,
              y: info.jsEvent.clientY,
            });
          }}
          firstDay={preferences.first_day_of_week === "Monday" ? "1" : "0"}
          lazyFetching={true}
          slotHeaderFormat={{
            hour: "numeric",
            minute: "numeric",
            hour12: preferences.time_format === "12-hour",
          }}
          slotDuration={preferences.time_frequency}
          nowIndicator={true}
          titleFormat={{
            year: "numeric",
            month: "long",
            day: "numeric",
          }}
          allDaySlot={true}
          displayEventEnd={false}
          snapDuration={{ minutes: 5 }}
          eventDragStart={() => {
            isDraggingRef.current = true;
          }}
          eventDragStop={() => {
            isDraggingRef.current = false;
          }}
          eventDrop={(e) => {
            applyDroppedEmployee(e);
            const type = e.event.extendedProps.type;

            if (type === "blocked") {
              openSidePanel("blocked-time", e.event, e.event.id, {
                revert: e.revert,
              });
              return;
            }

            openBookingSidePanel(e.event, e.revert);
          }}
          eventAllow={(dropInfo) => {
            if (dropInfo.start.getTime() < Date.now()) {
              return false;
            } else {
              return true;
            }
          }}
          dayHeaderFormat={{
            weekday: "short",
            day: isWindowSmall ? undefined : "numeric",
            omitCommas: true,
          }}
          eventTimeFormat={{
            hour: "2-digit",
            minute: "2-digit",
            hour12: preferences.time_format === "12-hour",
          }}
          views={{
            resourceTimeGrid2Day: {
              type: "resourceTimeGrid",
              duration: { days: 2 },
            },
            listWeek: {
              borderless: true,
              noEventsText: "Nothing scheduled this week",
              noEventsClass: "flex flex-1 items-center justify-center",
              noEventsInnerClass: "text-fg-muted text-sm",
              listDayFormat: {
                weekday: "long",
                month: "short",
                day: "numeric",
              },
              listDayHeaderContent: (info) => <ListDayHeader info={info} />,
              listDayHeaderDidMount: ({ el, isToday }) => {
                if (isToday) scrollToDay(el);
              },
              // Leaves room for the floating create button.
              listDaysClass: "flex flex-col gap-3 pb-24",
              listDayClass: "flex flex-col",
              listDayHeaderClass: "bg-bg z-10 pt-1 pb-2",
              // FullCalendar always renders a second date line, so hide it.
              listDayHeaderInnerClass: (info) =>
                info.level ? "hidden" : "flex items-baseline gap-2",
              listDayBodyClass: "flex flex-col gap-2",
              listItemEventClass: (info) =>
                `relative flex cursor-pointer overflow-hidden rounded-xl border
                border-border bg-bg-raised py-3 pr-4 pl-5 shadow-card
                hover:bg-hover ${info.isPast ? "opacity-60" : ""}`,
              listItemEventBeforeClass:
                "absolute inset-y-0 left-0 w-1.5 bg-(--fc-event-color)",
              listItemEventInnerClass: "min-w-0 flex-1",
            },
            resourceTimeGrid: {
              dayMinWidth: 120,
            },
          }}
          businessHours={transformBusinessHours(businessHours)}
        />
      </div>
    </div>
  );
}
