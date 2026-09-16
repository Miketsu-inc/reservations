import {
  createContext,
  useContext,
  useLayoutEffect,
  useRef,
  useState,
} from "react";

const ToggleGroupContext = createContext();

export function Toggle({
  value,
  styles,
  pressed,
  onPressedChange,
  defaultPressed = false,
  disabled = false,
  badgeText = null,
  children,
}) {
  const group = useContext(ToggleGroupContext);
  const [isPressedInternal, setIsPressedInternal] = useState(defaultPressed);

  const isPressed = group
    ? group.multiple
      ? group.value.includes(value)
      : group.value === value
    : pressed !== undefined
      ? pressed
      : isPressedInternal;

  return (
    <button
      ref={(element) => (group ? group.registerRef(value, element) : {})}
      aria-pressed={isPressed}
      type="button"
      disabled={disabled}
      className={`${styles} ${
        !group || group.multiple
          ? isPressed
            ? "bg-black dark:bg-white"
            : ""
          : "relative z-10 transition-colors duration-150"
        } ${
        isPressed ? "text-white delay-150 dark:text-black" : "text-text_color"
      }
        shrink-0 cursor-pointer rounded-3xl px-4 py-2 font-semibold text-nowrap`}
      onClick={() => {
        if (group) {
          group.onToggle(value);
        } else {
          // called before internal state is set due to asynchronous react state updates
          onPressedChange?.(!isPressed);
          if (pressed === undefined) setIsPressedInternal(!isPressed);
        }
      }}
    >
      {children}
      {badgeText !== null && (
        <span
          className={`ml-2 inline-flex items-center justify-center rounded-full
          px-2 py-1 text-sm leading-none font-bold transition-colors
          duration-150 ${
            isPressed
              ? "bg-white/20 text-white dark:bg-black/15 dark:text-black"
              : `text-text_color dark:text-text_color bg-black/10
                dark:bg-white/10`
          }`}
        >
          {badgeText}
        </span>
      )}
    </button>
  );
}

export function ToggleGroup({
  styles,
  value,
  defaultValue,
  onValueChange,
  multiple = false,
  disableDeselect = true,
  children,
}) {
  const initial = defaultValue ?? (multiple ? [] : null);
  const [internalValue, setInternalValue] = useState(initial);

  const isControlled = value !== undefined;
  const currentValue = isControlled ? value : internalValue;

  const itemRefs = useRef({});
  const groupRef = useRef(null);

  const isReady = useRef(false);
  const [pillStyle, setPillStyle] = useState({ opacity: 0 });

  useLayoutEffect(() => {
    if (multiple || !currentValue || !groupRef.current) return;

    function updatePillPosition() {
      const activeElement = itemRefs.current[currentValue];
      if (!activeElement) return;

      setPillStyle({
        width: `${activeElement.offsetWidth}px`,
        transform: `translateX(${activeElement.offsetLeft}px)`,
        opacity: 1,
        // Suppress transition on the very first render
        transition: isReady.current ? "" : "none",
      });
    }

    updatePillPosition();

    // Badge text and other asynchronous content can resize the active toggle
    // or a preceding toggle, which also changes the active toggle's offset.
    const resizeObserver = new ResizeObserver(updatePillPosition);
    Object.values(itemRefs.current).forEach((element) => {
      resizeObserver.observe(element);
    });

    // Re-enable transitions after the first paint
    let animationFrame;
    if (!isReady.current) {
      animationFrame = requestAnimationFrame(() => {
        isReady.current = true;
        setPillStyle((prev) => ({ ...prev, transition: "" }));
      });
    }

    return () => {
      resizeObserver.disconnect();
      if (animationFrame !== undefined) {
        cancelAnimationFrame(animationFrame);
      }
    };
  }, [children, currentValue, multiple]);

  function registerRef(itemValue, element) {
    if (element) {
      itemRefs.current[itemValue] = element;
    } else {
      delete itemRefs.current[itemValue];
    }
  }

  function onToggle(itemValue) {
    itemRefs.current[itemValue]?.scrollIntoView({
      behavior: "smooth",
      block: "nearest",
      container: "nearest",
      inline: "center",
    });

    let val;

    if (multiple) {
      const arr = currentValue;

      if (disableDeselect && arr.includes(itemValue)) return;

      val = arr.includes(itemValue)
        ? arr.filter((v) => v !== itemValue)
        : [...arr, itemValue];
    } else {
      if (disableDeselect && currentValue === itemValue) return;

      val = currentValue === itemValue ? null : itemValue;
    }

    if (!isControlled) setInternalValue(val);
    onValueChange?.(val);
  }

  return (
    <ToggleGroupContext.Provider
      value={{ multiple, value: currentValue, registerRef, onToggle }}
    >
      <div
        ref={groupRef}
        role="group"
        className={`${styles} relative flex h-fit scrollbar-thin
          overflow-x-auto`}
      >
        {!multiple && (
          <div
            aria-hidden="true"
            style={pillStyle}
            className={`pointer-events-none absolute inset-y-0 rounded-3xl
            bg-black transition-[transform,width,opacity] duration-300
            dark:bg-white`}
          ></div>
        )}
        {children}
      </div>
    </ToggleGroupContext.Provider>
  );
}
