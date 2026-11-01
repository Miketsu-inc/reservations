import {
  createContext,
  useCallback,
  useContext,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";

const ToggleGroupContext = createContext();

const variants = {
  pill: {
    group: "scrollbar-thin overflow-x-auto",
    item: "shrink-0 rounded-3xl px-4 py-2",
    pill: "bg-fg inset-y-0 rounded-3xl",
    pressedBg: "bg-fg",
    pressedText: "text-bg delay-25",
    badge: "text-fg bg-bg-inactive",
    pressedBadge: "bg-bg/15 text-bg delay-25",
  },
  segmented: {
    // rings instead of borders so the pill's offsetLeft math stays unaffected
    group: "bg-bg-sunken ring-border rounded-lg p-1.25 ring-1 ring-inset",
    item: "min-w-0 flex-1 rounded-md px-2.5 py-1 ",
    // inset matches the group's padding so the pill sits inside the track
    pill: "bg-bg-overlay shadow-card inset-y-1.25 rounded-md",
    pressedBg: "bg-bg-overlay shadow-card",
    pressedText: "text-fg font-medium",
    badge: "text-fg bg-bg-inactive",
    pressedBadge: "text-fg bg-bg-sunken",
  },
};

export function Toggle({
  value,
  styles,
  pressed,
  onPressedChange,
  defaultPressed = false,
  disabled = false,
  badgeText = null,
  children,
  ...props
}) {
  const group = useContext(ToggleGroupContext);
  const variant = variants[group?.variant ?? "pill"];
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
      {...props}
      ref={(element) => {
        group?.registerRef(value, element);
      }}
      aria-pressed={isPressed}
      type="button"
      disabled={disabled}
      className={`${styles} ${
        !group || group.multiple
          ? isPressed
            ? variant.pressedBg
            : ""
          : "relative z-10"
        } ${isPressed ? variant.pressedText : "text-fg-muted hover:text-fg"}
        ${variant.item} cursor-pointer text-nowrap`}
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
          px-2 py-1 text-sm leading-none ${
            isPressed ? variant.pressedBadge : variant.badge
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
  variant = "pill",
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
    if (multiple || !currentValue || !groupRef.current) {
      setPillStyle({ opacity: 0 });
      return;
    }

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

  const registerRef = useCallback((itemValue, element) => {
    if (element) {
      itemRefs.current[itemValue] = element;
    } else {
      delete itemRefs.current[itemValue];
    }
  }, []);

  const onToggle = useCallback(
    (itemValue) => {
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
    },
    [currentValue, disableDeselect, isControlled, multiple, onValueChange]
  );

  const contextValue = useMemo(
    () => ({
      multiple,
      value: currentValue,
      registerRef,
      onToggle,
      variant,
    }),
    [currentValue, multiple, onToggle, registerRef, variant]
  );

  return (
    <ToggleGroupContext.Provider value={contextValue}>
      <div
        ref={groupRef}
        role="group"
        className={`${styles} ${variants[variant].group} relative flex h-fit`}
      >
        {!multiple && (
          <div
            aria-hidden="true"
            style={pillStyle}
            className={`${variants[variant].pill} pointer-events-none absolute
            left-0 transition-[transform,width,opacity] duration-150`}
          ></div>
        )}
        {children}
      </div>
    </ToggleGroupContext.Provider>
  );
}
