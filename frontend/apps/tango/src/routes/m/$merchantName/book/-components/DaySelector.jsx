import { ArrowLeft01Icon, ArrowRight01Icon } from "@hugeicons/core-free-icons";
import { Icon } from "@reservations/components";
import { useCallback, useEffect, useRef } from "react";

export default function DaySelector({ days, selectedDate, onSelect }) {
  const containerRef = useRef(null);
  const buttonRefs = useRef(new Map());

  const registerRef = (date, el) => {
    if (el) buttonRefs.current.set(date, el);
    else buttonRefs.current.delete(date);
  };

  const scrollToDate = useCallback((date) => {
    const el = buttonRefs.current.get(date);
    const container = containerRef.current;
    if (!el || !container) return;

    const containerRect = container.getBoundingClientRect();
    const elRect = el.getBoundingClientRect();
    const isFullyVisible =
      elRect.left >= containerRect.left && elRect.right <= containerRect.right;

    if (!isFullyVisible) {
      el.scrollIntoView({
        behavior: "smooth",
        inline: "center",
        block: "nearest",
      });
    }
  }, []);

  useEffect(() => {
    if (!selectedDate) return;
    scrollToDate(selectedDate);
  }, [selectedDate, scrollToDate]);

  function scrollByPage(direction) {
    const container = containerRef.current;
    if (!container) return;
    container.scrollBy({
      left: direction * container.clientWidth * 0.8,
      behavior: "smooth",
    });
  }

  return (
    <div className="group relative flex w-full min-w-0 items-center gap-2">
      <button
        type="button"
        aria-label="Scroll to earlier days"
        onClick={() => scrollByPage(-1)}
        className="border-border_color bg-layer_bg pointer-events-none absolute
          left-0 z-10 hidden size-10 -translate-x-1/4 items-center
          justify-center rounded-full border opacity-0 shadow-lg
          transition-opacity duration-200 group-hover:pointer-events-auto
          group-hover:opacity-100 hover:size-11 md:flex"
      >
        <Icon icon={ArrowLeft01Icon} styles="size-5" />
      </button>

      <div
        ref={containerRef}
        className="flex flex-1 snap-x snap-mandatory scrollbar-none gap-3
          overflow-x-auto scroll-smooth px-1 py-1"
      >
        {days.map((day) => (
          <DayButton
            key={day.date}
            day={day}
            isSelected={day.date === selectedDate}
            isUnavailable={!day.is_available}
            onClick={onSelect}
            registerRef={registerRef}
          />
        ))}
      </div>

      <button
        type="button"
        aria-label="Scroll to later days"
        onClick={() => scrollByPage(1)}
        className="border-border_color bg-layer_bg pointer-events-none absolute
          right-0 z-10 hidden size-10 translate-x-1/4 items-center
          justify-center rounded-full border opacity-0 shadow-lg
          transition-opacity duration-200 group-hover:pointer-events-auto
          group-hover:opacity-100 hover:size-11 md:flex"
      >
        <Icon icon={ArrowRight01Icon} styles="size-5" />
      </button>
    </div>
  );
}

function DayButton({ day, isSelected, isUnavailable, onClick, registerRef }) {
  const date = new Date(day.date);
  const weekday = date.toLocaleDateString("en-US", { weekday: "short" });
  const dayNumber = date.getDate();
  const month = date.toLocaleDateString("en-US", { month: "short" });

  return (
    <button
      ref={(el) => registerRef(day.date, el)}
      type="button"
      onClick={() => onClick(day.date)}
      aria-pressed={isSelected}
      aria-label={
        isUnavailable
          ? `${date.toDateString()}, no availability`
          : date.toDateString()
      }
      className={`flex w-16 shrink-0 snap-center scroll-mx-4 flex-col
        items-center gap-1 rounded-lg border py-3 transition-colors ${
          isSelected
            ? "bg-primary border-primary text-white"
            : isUnavailable
              ? `border-border_color text-text_color/50 hover:bg-bg_color
                bg-gray-200/50 dark:bg-gray-200/5`
              : `border-border_color bg-layer_bg text-text_color
                hover:bg-bg_color`
        }`}
    >
      <span className="text-xs font-medium uppercase opacity-70">
        {weekday}
      </span>
      <span className="text-xl font-bold">{dayNumber}</span>
      <span className="text-[11px] opacity-70">{month}</span>
    </button>
  );
}
