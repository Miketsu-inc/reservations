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
        isPressed
          ? "text-white [transition-delay:150ms] dark:text-black"
          : "text-text_color [transition-delay:0ms]"
        } cursor-pointer rounded-3xl px-4 py-2 font-semibold`}
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

    const activeElement = itemRefs.current[currentValue];
    if (!activeElement) return;

    const groupRect = groupRef.current.getBoundingClientRect();
    const activeRect = activeElement.getBoundingClientRect();

    setPillStyle({
      width: `${activeRect.width}px`,
      transform: `translateX(${activeRect.left - groupRect.left}px)`,
      opacity: 1,
      // Suppress transition on the very first render
      transition: isReady.current ? "" : "none",
    });

    // Re-enable transitions after the first paint
    if (!isReady.current) {
      requestAnimationFrame(() => {
        isReady.current = true;
        setPillStyle((prev) => ({ ...prev, transition: "" }));
      });
    }
  }, [currentValue, multiple]);

  function registerRef(itemValue, element) {
    if (element) itemRefs.current[itemValue] = element;
  }

  function onToggle(itemValue) {
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
      <div ref={groupRef} role="group" className={`${styles} relative flex`}>
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
