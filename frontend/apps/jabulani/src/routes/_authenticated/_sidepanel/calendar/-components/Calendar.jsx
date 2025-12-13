import dayGridPlugin from "@fullcalendar/daygrid";
import interactionPlugin from "@fullcalendar/interaction";
import listPlugin from "@fullcalendar/list";
import FullCalendar from "@fullcalendar/react";
import timeGridPlugin from "@fullcalendar/timegrid";
import { BackArrowIcon } from "@reservations/assets";
import {
  Button,
  DatePicker,
  Select,
  ServerError,
} from "@reservations/components";
import {
  businessHoursQueryOptions,
  formatToDateString,
  getMonthFromCalendarStart,
  preferencesQueryOptions,
  useWindowSize,
} from "@reservations/lib";
import {
  keepPreviousData,
  useQuery,
  useSuspenseQueries,
} from "@tanstack/react-query";
import { useCallback, useEffect, useRef, useState } from "react";
import { bookingsQueryOptions } from "..";
import BlockedTimeModal from "./BlockedTimeModal";
import CalendarModal from "./CalendarModal";
import CreateMenu from "./CreateMenu";
import DragConfirmationModal from "./DragConfirmationModal";
import NewBookingModal from "./NewBookingModal";

const calendarViewOptions = [
  { value: "dayGridMonth", label: "Month" },
  { value: "timeGridWeek", label: "Week" },
  { value: "timeGridDay", label: "Day" },
  { value: "listWeek", label: "List" },
];

function getContrastColor(color) {
  const hex = color.replace("#", "");

  const red = parseInt(hex.substring(0, 2), 16);
  const green = parseInt(hex.substring(2, 4), 16);
  const blue = parseInt(hex.substring(4, 6), 16);

  const brightness = red * 0.289 + green * 0.587 + blue * 0.114;
  return brightness > 186 ? "#000000" : "#ffffff";
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

  return data.map((booking) => ({
    id: booking.id,
    title:
      booking.first_name && booking.last_name
        ? `${booking.first_name} ${booking.last_name}`
        : "Walk-in",
    start: booking.from_date,
    end: booking.to_date,
    color: booking.service_color,
    textColor: getContrastColor(booking.service_color),
    durationEditable: false,
    startEditable: new Date(booking.to_date) > new Date() ? true : false,
    extendedProps: {
      // this is a number unlike the normal 'id' which get's converted to a string
      id: booking.id,
      type: "booking",
      first_name: booking.first_name,
      last_name: booking.last_name,
      phone_number: booking.phone_number,
      customer_note: booking.customer_note,
      merchant_note: booking.merchant_note,
      service_name: booking.service_name,
      service_duration: booking.service_duration,
      price: booking.price,
    },
  }));
}

function formatBlockedTimes(data) {
  if (data === undefined) return;

  return data.map((blockedTime) => ({
    id: blockedTime.id,
    title: blockedTime.name,
    start: blockedTime.from_date,
    end: blockedTime.to_date,
    color: "rgba(0, 0, 0, 0.6)",
    textColor: getContrastColor("#333333"),
    durationEditable: true,
    allDay: blockedTime.all_day,
    startEditable: new Date(blockedTime.to_date) > new Date() ? true : false,
    extendedProps: {
      id: blockedTime.id,
      type: "blocked",
      employee_id: blockedTime.employee_id,
      allDay: blockedTime.all_day,
    },
  }));
}

const defaultBookingInfo = {
  id: 0,
  title: "",
  start: new Date(),
  end: new Date(),
  startEditable: true,
  extendedProps: {
    id: 0,
    first_name: "",
    last_name: "",
    phone_number: "",
    customer_note: "",
    merchant_note: "",
    price: "",
  },
};

export default function Calendar({ router, route, search }) {
  const [calendarTitle, setCalendarTitle] = useState("");
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [bookingInfo, setBookingInfo] = useState(defaultBookingInfo);
  const [isBlockedTimeModalOpen, setIsBlockedTimeModalOpen] = useState(false);
  const [blockedTimeModalData, setBlockedTimeModalData] = useState(null);
  const [dragModalOpen, setDragModalOpen] = useState(false);
  const [dragModalData, setDragModalData] = useState({
    booking: {},
    old_booking: {},
    revert: {},
  });
  const [isNewBookingModalOpen, setIsNewBookingModalOpen] = useState(false);

  const { queryClient } = route.useRouteContext({ from: route.id });
  const {
    data: events = { bookings: [], blocked_times: [] },
    isError,
    error,
  } = useQuery({
    ...bookingsQueryOptions(search.start, search.end),
    placeholderData: keepPreviousData,
  });
  const [{ data: preferences }, { data: businessHours }] = useSuspenseQueries({
    queries: [preferencesQueryOptions(), businessHoursQueryOptions()],
  });
  const [calendarView, setCalendarView] = useState(search.view);

  const windowSize = useWindowSize();
  const calendarRef = useRef();

  const isWindowSmall =
    windowSize === "sm" || windowSize === "md" || windowSize === "lg";

  const invalidateBookingsQuery = useCallback(async () => {
    await queryClient.invalidateQueries({
      queryKey: ["events", search.start, search.end],
    });
  }, [queryClient, search]);

  const datesChanged = useCallback(
    (api) => {
      const start = formatToDateString(api.view.activeStart);
      const end = formatToDateString(api.view.activeEnd);

      router.navigate({
        search: () => ({ view: api.view.type, start: start, end: end }),
        replace: true,
      });
      setCalendarTitle(api.view.title);
    },
    [router, setCalendarTitle]
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
    api.changeView(view);

    datesChanged(api);
    setCalendarView(view);
  }

  function changeDateHandler(date) {
    const api = calendarRef.current.getApi();
    const dateStr = formatToDateString(date);

    api.gotoDate(dateStr);
    datesChanged(api);
  }

  useEffect(() => {
    const api = calendarRef.current.getApi();
    setCalendarTitle(api.view.title);
  }, []);

  if (isError) {
    return <ServerError error={error} />;
  }

  return (
    <div
      className="flex h-[85svh] flex-col px-4 py-2 md:h-fit md:max-h-[90svh]
        md:px-0 md:py-0"
    >
      <div className="relative flex flex-col py-4 md:flex-row md:gap-2">
        <div
          className="flex w-full flex-col justify-between md:flex-row
            md:items-center"
        >
          <p className="text-2xl whitespace-nowrap md:text-3xl">
            {calendarTitle}
          </p>
          <div className="flex flex-row items-center justify-between gap-2">
            <div className="flex flex-row items-center gap-2">
              <CreateMenu
                isFloating={isWindowSmall}
                onCreateBlockedTime={() => setIsBlockedTimeModalOpen(true)}
                onCreateBooking={() => setIsNewBookingModalOpen(true)}
              />
              <DatePicker
                styles="w-fit"
                hideText={true}
                firstDayOfWeek={preferences.first_day_of_week}
                clearAfterClose={true}
                onSelect={changeDateHandler}
              />
              <button
                className="hover:bg-hvr_gray cursor-pointer rounded-lg"
                type="button"
                onClick={() => navButtonHandler("prev")}
              >
                <BackArrowIcon styles="size-8 stroke-current" />
              </button>
              <Button
                variant="primary"
                styles="p-2 text-sm"
                buttonText="today"
                onClick={todayButtonHandler}
              />
              <button
                className="hover:bg-hvr_gray cursor-pointer rounded-lg"
                type="button"
                onClick={() => navButtonHandler("next")}
              >
                <BackArrowIcon styles="size-8 stroke-current rotate-180" />
              </button>
            </div>
            <Select
              options={calendarViewOptions}
              value={calendarView}
              onSelect={(option) => changeViewHandler(option.value)}
              styles="w-28!"
            />
          </div>
        </div>
      </div>
      <div
        className="max-h-full w-full overflow-auto rounded-lg bg-white
          text-black"
      >
        <FullCalendar
          ref={calendarRef}
          plugins={[
            dayGridPlugin,
            interactionPlugin,
            timeGridPlugin,
            listPlugin,
          ]}
          locale="hu"
          editable={true}
          eventDurationEditable={true}
          selectable={true}
          initialView={search.view ? search.view : "timeGridWeek"}
          // dayGridMonth dates do not start or end with the current month's dates
          initialDate={
            search.view === "dayGridMonth"
              ? getMonthFromCalendarStart(search.start)
              : search.start
                ? search.start
                : undefined
          }
          height="auto"
          headerToolbar={false}
          events={[
            ...formatBookings(events.bookings),
            ...formatBlockedTimes(events.blocked_times),
          ]}
          eventClick={(e) => {
            const type = e.event.extendedProps.type;

            if (type === "blocked") {
              setBlockedTimeModalData(e.event);
              setTimeout(() => setIsBlockedTimeModalOpen(true), 0);
              return;
            }

            setBookingInfo(e.event);
            setTimeout(() => setIsModalOpen(true), 0);
          }}
          firstDay={preferences.first_day_of_week === "Monday" ? "1" : "0"}
          lazyFetching={true}
          slotLabelFormat={{
            hour: "numeric",
            minute: "numeric",
            hour12: preferences.time_format === "12-hour",
          }}
          slotDuration={preferences.time_frequency}
          slotMinTime={preferences.start_hour}
          slotMaxTime={preferences.end_hour}
          nowIndicator={true}
          titleFormat={{
            year: "numeric",
            month: "long",
            day: "numeric",
          }}
          titleRangeSeparator=" - "
          fixedWeekCount={false}
          allDaySlot={true}
          displayEventEnd={false}
          snapDuration={{ minutes: 5 }}
          eventDrop={(e) => {
            const type = e.event.extendedProps.type;

            if (type === "blocked") {
              e.revert();
              return;
            }

            setDragModalData({
              booking: e.event,
              old_booking: e.oldEvent,
              revert: e.revert,
            });
            setTimeout(() => setDragModalOpen(true), 0);
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
            dayGridMonth: {
              titleFormat: {
                year: "numeric",
                month: "long",
              },
              displayEventEnd: false,
            },
            timeGridDay: {
              dayHeaderFormat: {
                weekday: "long",
              },
              displayEventTime: false,
            },
            listWeek: {
              displayEventEnd: true,
            },
          }}
          businessHours={transformBusinessHours(businessHours)}
        />
      </div>
      <CalendarModal
        key={bookingInfo?.id}
        bookingInfo={bookingInfo}
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
        }}
        onDeleted={invalidateBookingsQuery}
        onEdit={invalidateBookingsQuery}
      />
      <DragConfirmationModal
        isOpen={dragModalOpen}
        bookingData={dragModalData}
        onClose={() => {
          dragModalData.revert();
          setDragModalOpen(false);
        }}
        onMoved={async () => {
          await invalidateBookingsQuery();
          setDragModalOpen(false);
        }}
      />
      <BlockedTimeModal
        key={blockedTimeModalData?.extendedProps?.id || "new"}
        isOpen={isBlockedTimeModalOpen}
        onClose={() => {
          setBlockedTimeModalData(null);
          setIsBlockedTimeModalOpen(false);
        }}
        blockedTime={blockedTimeModalData}
        preferences={preferences}
        onDeleted={invalidateBookingsQuery}
        onSubmitted={invalidateBookingsQuery}
      />
      <NewBookingModal
        isOpen={isNewBookingModalOpen}
        onClose={() => setIsNewBookingModalOpen(false)}
        onNewBooking={invalidateBookingsQuery}
      />
    </div>
  );
}
