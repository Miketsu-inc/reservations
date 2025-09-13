import Button from "@components/Button";
import DatePicker from "@components/DatePicker";
import Select from "@components/Select";
import ServerError from "@components/ServerError";
import dayGridPlugin from "@fullcalendar/daygrid";
import interactionPlugin from "@fullcalendar/interaction";
import listPlugin from "@fullcalendar/list";
import FullCalendar from "@fullcalendar/react";
import timeGridPlugin from "@fullcalendar/timegrid";
import BackArrowIcon from "@icons/BackArrowIcon";
import { formatToDateString, getMonthFromCalendarStart } from "@lib/datetime";
import { useWindowSize } from "@lib/hooks";
import {
  businessHoursQueryOptions,
  preferencesQueryOptions,
} from "@lib/queries";
import {
  keepPreviousData,
  useQuery,
  useSuspenseQueries,
} from "@tanstack/react-query";
import { useCallback, useEffect, useRef, useState } from "react";
import { bookingsQueryOptions } from "..";
import CalendarModal from "./CalendarModal";
import DragConfirmationModal from "./DragConfirmationModal";

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

function formatData(data) {
  if (data === undefined) return;

  return data.map((booking) => ({
    id: booking.id,
    title: booking.first_name + " " + booking.last_name,
    start: booking.from_date,
    end: booking.to_date,
    color: booking.service_color,
    textColor: getContrastColor(booking.service_color),
    durationEditable: false,
    startEditable: new Date(booking.to_date) > new Date() ? true : false,
    extendedProps: {
      // this is a number unlike the normal 'id' which get's converted to a string
      booking_id: booking.id,
      group_id: booking.group_id,
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

const defaultBookingInfo = {
  id: 0,
  group_id: 0,
  title: "",
  start: new Date(),
  end: new Date(),
  startEditable: true,
  extendedProps: {
    booking_id: 0,
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
  const [dragModalOpen, setDragModalOpen] = useState(false);
  const [dragModalData, setDragModalData] = useState({
    booking: {},
    old_booking: {},
    revert: {},
  });

  const { queryClient } = route.useRouteContext({ from: route.id });
  const {
    data: bookings,
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
      queryKey: ["bookings", search.start, search.end],
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
      <div className="flex flex-col py-4 md:flex-row md:gap-2">
        <div
          className="flex w-full flex-col justify-between md:flex-row
            md:items-center"
        >
          <p className="text-2xl whitespace-nowrap md:text-3xl">
            {calendarTitle}
          </p>
          <div className="flex flex-row items-center justify-between gap-2">
            <div className="flex flex-row items-center gap-2">
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
        className="light bg-bg_color text-text_color max-h-full w-full
          overflow-auto rounded-lg"
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
          events={formatData(bookings)}
          eventClick={(e) => {
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
          allDaySlot={false}
          displayEventEnd={false}
          snapDuration={{ minutes: 5 }}
          eventDrop={(e) => {
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
    </div>
  );
}
