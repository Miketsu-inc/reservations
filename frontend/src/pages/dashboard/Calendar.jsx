import dayGridPlugin from "@fullcalendar/daygrid";
import interactionPlugin from "@fullcalendar/interaction";
import listPlugin from "@fullcalendar/list";
import FullCalendar from "@fullcalendar/react";
import timeGridPlugin from "@fullcalendar/timegrid";
import { useCallback, useState } from "react";


export default function Calendar() {
const [isLoading, setIsLoading] = useState(false);
  const [serverError, setServerError] = useState(undefined);
 
  const formatData = (data) => {
    return data.map((event) => ({
      id: event.location,
      title: event.appointment_type,
      start: event.from_date,
      end: event.to_date,
      extendedProps: {
        name: event.user,
        //anything basically
      },
    }));
  };

  const fetchEvents = useCallback(
    async (fetchInfo, successCallback, failureCallback) => {
      setIsLoading(true);
      try {
        const response = await fetch(
          `/api/v1/appointments/calendar?start=${fetchInfo.startStr}&end=${fetchInfo.endStr}`,
          {
            method: "GET",
          }
        );
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        console.log(data);
        if (data.error) {
          setServerError(data.error);
          failureCallback(data.error);
        } else {
          const events = formatData(data);
          successCallback(events);
        }
      } catch (err) {
        setServerError(
          "Error occured. Please try again by refreshing the page"
        );
        failureCallback(err);
      } finally {
        setServerError(undefined);
        setIsLoading(false);
      }
    },
    []
  );
  return (
<div className="flex items-center justify-center">
      <div className="w-1/2">
        {isLoading && <div>Loading the calendar</div>}
        {serverError && <span className="text-red-600">{serverError}</span>}
      </div>
    <div className="bg-bg_color text-text_color">
      <FullCalendar
        plugins={[
dayGridPlugin,
interactionPlugin,
timeGridPlugin,
listPlugin,
]}
        weekNumberCalculation="ISO"
        locale="hu"
timeZone="UTC"
        editable={true}
        eventDurationEditable={true}
        selectable={true}
        initialView="timeGridWeek"
        weekNumbers={true}
        navLinks={true}
        height="auto"
        events={fetchEvents}
        eventClick={(e) => {
          const id = e.event.id;
          const title = e.event.title;
          const date = e.event.start;
          const end = e.event.end;

          console.log(id);
          console.log(title);
          console.log(date);
          console.log(end);
        }}
lazyFetching={true}
        // views={{
        //   dayGridMonth: {
        //     fixedWeekCount: false,
        //   },
        //   timeGridWeek: {
        //     titleFormat: {
        //       year: "numeric",
        //       month: "long",
        //       day: "2-digit",
        //     },
        //     slotLabelFormat: {
        //       hour: "numeric",
        //       minute: "2-digit",
        //     },
        //     slotDuration: "00:15:00",
        //     slotMinTime: "08:00:00",
        //     slotMaxTime: "17:30:00",
        //     nowIndicator: true,
        //   },
        //   timeGridDay: {
        //     slotLabelFormat: {
        //       hour: "numeric",
        //       minute: "2-digit",
        //     },
        //     slotDuration: "00:15:00",
        //     slotMinTime: "08:00:00",
        //     slotMaxTime: "17:30:00",
        //     nowIndicator: true,
        //   },
        // }}
        // headerToolbar={{
        //   left: "dayGridMonth,timeGridWeek,timeGridDay,list",
        //   center: "title",
        //   right: "today,prev,next",
        // }}
        // allDaySlot={false}
        // eventTimeFormat={{
        //   hour: "numeric",
        //   minute: "2-digit",
        //   second: "2-digit",
        //   meridiem: false,
        // }}
        // buttonText={{
        //   month: "hónap",
        //   today: "ma",
        //   week: "hét",
        //   day: "nap",
        //   list: "lista",
        // }}
      />
    </div>
  );
}
