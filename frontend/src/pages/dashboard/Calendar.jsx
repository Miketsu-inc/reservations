import dayGridPlugin from "@fullcalendar/daygrid";
import interactionPlugin from "@fullcalendar/interaction";
import listPlugin from "@fullcalendar/list";
import FullCalendar from "@fullcalendar/react";
import timeGridPlugin from "@fullcalendar/timegrid";

export default function Calendar() {
    return (
        <FullCalendar
            plugins={[dayGridPlugin, interactionPlugin, timeGridPlugin, listPlugin]}
            weekNumberCalculation={"ISO"}
            locale={'hu'}
            editable={true}
            selectable={true}
            initialView="dayGridMonth"
            weekNumbers={true}
            navLinks={true}
            height={"auto"}
            views={{
                dayGridMonth: {
                    fixedWeekCount: false
                },
                timeGridWeek: {
                    titleFormat: {
                        year: "numeric",
                        month: "long",
                        day: "2-digit"
                    },
                    slotLabelFormat: {
                        hour: "numeric",
                        minute: "2-digit",
                    },
                    slotDuration: "00:15:00",
                    slotMinTime: "08:00:00",
                    slotMaxTime: "17:30:00",
                    nowIndicator: true
                },
                timeGridDay: {
                    slotLabelFormat: {
                        hour: "numeric",
                        minute: "2-digit",
                    },
                    slotDuration: "00:15:00",
                    slotMinTime: "08:00:00",
                    slotMaxTime: "17:30:00",
                    nowIndicator: true
                }
            }}
            headerToolbar={{
                left: "dayGridMonth,timeGridWeek,timeGridDay",
                center: "title",
                right: "today,prev,next",
            }}
            allDaySlot={false}
            eventTimeFormat={{
                hour: "numeric",
                minute: "2-digit",
                second: "2-digit",
                meridiem: false
            }}
            buttonText={{
                month: "hónap",
                today: "ma",
                week: "hét",
                day: "nap",
                list: "lista"
            }}
        />
    )
}