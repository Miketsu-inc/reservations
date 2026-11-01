import { timeGridViews } from "./calendarViews";

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
export function scrollToDay(header) {
  const day = header.parentElement;
  const scroller = findScroller(document.body, day);
  if (!scroller) return;

  scroller.scrollTop +=
    day.getBoundingClientRect().top - scroller.getBoundingClientRect().top;
}

// Centers the current time on mount and when the grid resizes. Measures the
// visible grid height to offset FullCalendar's top-aligned scrolling.
export function observeCurrentTime({ el, view }) {
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
