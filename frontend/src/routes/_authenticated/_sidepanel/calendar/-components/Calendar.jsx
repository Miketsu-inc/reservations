import ServerError from "@components/ServerError";
import dayGridPlugin from "@fullcalendar/daygrid";
import interactionPlugin from "@fullcalendar/interaction";
import listPlugin from "@fullcalendar/list";
import FullCalendar from "@fullcalendar/react";
import timeGridPlugin from "@fullcalendar/timegrid";
import { useClickOutside } from "@lib/hooks";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { useCallback, useRef, useState } from "react";
import CalendarModal from "./CalendarModal";

const defaultEventInfo = {
  extendedProps: {
    appointment_id: 0,
    first_name: "",
    last_name: "",
    phone_number: "",
    user_comment: "",
    merchant_comment: "",
    price: 0,
  },
};

export default function Calendar() {
  const modalRef = useRef();
  const [serverError, setServerError] = useState();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [eventInfo, setEventInfo] = useState(defaultEventInfo);
  useClickOutside(modalRef, () => setIsModalOpen(false));

  const formatData = (data) => {
    return data.map((event) => ({
      id: event.id,
      title: event.service_name,
      start: event.from_date,
      end: event.to_date,
      extendedProps: {
        appointment_id: event.id,
        first_name: event.first_name,
        last_name: event.last_name,
        phone_number: event.phone_number,
        user_comment: event.user_comment,
        merchant_comment: event.merchant_comment,
        price: event.price,
      },
    }));
  };

  const fetchEvents = useCallback(
    async (fetchInfo, successCallback, failureCallback) => {
      try {
        const response = await fetch(
          `/api/v1/appointments/calendar?start=${fetchInfo.startStr}&end=${fetchInfo.endStr}`,
          {
            method: "GET",
          }
        );

        const result = await response.json();

        if (!response.ok) {
          invalidateLocalSotrageAuth(response.status);
          setServerError(result.error.message);
          failureCallback(result.error.message);
        } else {
          setServerError("");

          if (result.data !== null) {
            const events = formatData(result.data);
            successCallback(events);
          } else {
            failureCallback("No appointments found");
          }
        }
      } catch (err) {
        setServerError(err.message);
        failureCallback(err.message);
      }
    },
    []
  );

  function handleClick(e) {
    setEventInfo(e.event);
    setIsModalOpen(true);
  }

  return (
    <>
      <ServerError styles="mt-4 mb-2" error={serverError} />
      <div className="flex items-center justify-center">
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
            eventClick={handleClick}
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
        {isModalOpen && (
          <CalendarModal
            ref={modalRef}
            eventInfo={eventInfo}
            close={() => {
              setIsModalOpen(false);
            }}
            setError={setServerError}
          />
        )}
      </div>
    </>
  );
}
