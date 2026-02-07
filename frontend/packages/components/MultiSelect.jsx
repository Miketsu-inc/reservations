import { BackArrowIcon } from "@reservations/assets";
import { useEffect, useRef, useState } from "react";
import { Avatar } from ".";
import CheckBox from "./CheckBox";
import { Popover, PopoverContent, PopoverTrigger } from "./Popover";

const itemHeight = 36;

export default function MultiSelect({
  options,
  values = [],
  onSelect,
  placeholder,
  labelText,
  displayText = "items",
  icon,
  required,
  styles,
  maxVisibleItems = 7,
  onOpenChange,
  disabled,
  emptyText,
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [highlightedIndex, setHighlightedIndex] = useState(0);
  const [isUsingKeyboard, setIsUsingKeyboard] = useState(false);
  const [triggerWidth, setTriggerWidth] = useState(null);

  const containerRef = useRef(null);
  const dropDownListRef = useRef(null);

  const allSelected = options.length > 0 && values.length === options.length;

  function handleOpen() {
    if (disabled) return;
    setIsOpen(true);
    setHighlightedIndex(0);

    if (containerRef.current) {
      setTriggerWidth(containerRef.current.offsetWidth);
    }
  }

  function handleClose() {
    setIsOpen(false);
  }

  function handleToggleAll() {
    if (allSelected) {
      onSelect([]);
    } else {
      onSelect(options.map((option) => option.value));
    }
  }

  function handleToggleOption(optionValue) {
    if (values.includes(optionValue)) {
      onSelect(values.filter((id) => id !== optionValue));
    } else {
      onSelect([...values, optionValue]);
    }
  }

  function getDisplayText() {
    if (!values || values.length === 0) return placeholder;
    if (values.length === options.length) return `All ${displayText}`;
    if (values.length === 1) {
      const selectedItem = options.find((o) => o.id === values[0]);
      return selectedItem ? selectedItem.label : `1 ${displayText} selected`;
    }
    return `${values.length} ${displayText} selected`;
  }

  function handleKeyDown(e) {
    // options + 1 for 'All' - 1 for 0-index
    const maxIndex = options.length;
    setIsUsingKeyboard(true);
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      if (!isOpen) {
        handleOpen();
      } else {
        if (highlightedIndex === 0) {
          handleToggleAll();
        } else {
          const option = options[highlightedIndex - 1];
          if (option) handleToggleOption(option.value);
        }
      }
    } else if (e.key === "ArrowUp" && isOpen) {
      e.preventDefault();
      setHighlightedIndex((prev) => (prev > 0 ? prev - 1 : prev));
    } else if (e.key === "ArrowDown" && isOpen) {
      e.preventDefault();
      setHighlightedIndex((prev) => (prev < maxIndex ? prev + 1 : prev));
    }
  }

  useEffect(() => {
    if (!isOpen || !isUsingKeyboard) return;

    const timeout = setTimeout(() => {
      if (dropDownListRef.current?.children[highlightedIndex]) {
        dropDownListRef.current.children[highlightedIndex].scrollIntoView({
          block: "center",
        });
      }
    }, 0);

    return () => clearTimeout(timeout);
  }, [highlightedIndex, isUsingKeyboard, isOpen]);

  return (
    <Popover
      open={isOpen}
      onOpenChange={(open) => {
        open ? handleOpen() : handleClose();
        onOpenChange?.(open);
      }}
    >
      <PopoverTrigger asChild>
        <label className={`w-full ${styles}`}>
          {labelText && (
            <span className="flex items-center gap-1 pb-1 text-sm">
              {labelText}
              {required !== false && (
                <span className="text-base leading-none text-red-500">*</span>
              )}
            </span>
          )}
          <button
            className={`${styles} border-input_border_color
              disabled:border-input_border_color/60 w-full min-w-fit rounded-lg
              border py-2 pr-2 pl-3 text-left transition-opacity
              disabled:bg-gray-200/60 disabled:dark:bg-gray-700/20`}
            type="button"
            ref={containerRef}
            disabled={disabled}
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                {icon && values.length > 0 && (
                  <span className="text-gray-500">{icon}</span>
                )}
                <span
                  className={`${values.length > 0 ? "text-text_color" : "text-gray-500"}
                    min-h-6 flex-1 truncate
                    ${disabled ? "text-text_color/70" : ""}`}
                >
                  {getDisplayText()}
                </span>
              </div>
              <BackArrowIcon
                styles={`stroke-gray-700 dark:stroke-gray-300
                  transition-transform -rotate-90 shrink-0
                  ${isOpen ? "rotate-90" : ""} size-5`}
              />
            </div>
          </button>
        </label>
      </PopoverTrigger>
      <PopoverContent
        forceMount
        styles="p-0! data-[side=top]:translate-y-6"
        onKeyDown={handleKeyDown}
        style={{
          width: triggerWidth || "auto",
        }}
      >
        <ul
          ref={dropDownListRef}
          style={{
            maxHeight: itemHeight
              ? `${itemHeight * maxVisibleItems + 8}px`
              : "auto",
          }}
          className="overflow-x-hidden overflow-y-auto p-1 transition-all
            dark:scheme-dark"
          onMouseMove={() => {
            setIsUsingKeyboard(false);
            setHighlightedIndex(null);
          }}
        >
          {options.length === 0 ? (
            <li
              className="px-4 py-6 text-center text-gray-500 select-none
                dark:text-gray-400"
            >
              {emptyText || "No results found"}
            </li>
          ) : (
            <>
              <li
                onClick={handleToggleAll}
                className={`${
                  highlightedIndex === 0 ? "bg-hvr_gray" : "hover:bg-hvr_gray"
                }
                  dark:text-text_color mb-1 cursor-pointer rounded-sm py-2
                  pr-0.5 pl-2 text-gray-700 select-none`}
                role="option"
                aria-selected={allSelected}
              >
                <div className="flex w-full items-center gap-3">
                  <CheckBox
                    checked={allSelected}
                    readOnly
                    styles="outline-none"
                  />
                  <span className="text-sm font-medium">All {displayText}</span>
                </div>
              </li>

              {options.map((option, index) => {
                const listIndex = index + 1;
                const isSelected = values.includes(option.value);
                const isHighlighted = highlightedIndex === listIndex;

                return (
                  <li
                    key={index}
                    onClick={() => handleToggleOption(option.value)}
                    className={`${isHighlighted ? "bg-hvr_gray" : isUsingKeyboard ? "" : "hover:bg-hvr_gray"}
                      dark:text-text_color cursor-pointer rounded-sm py-2 pr-0.5
                      pl-2 text-gray-700 select-none`}
                    role="option"
                    aria-selected={isSelected}
                  >
                    <div
                      className="flex w-full cursor-pointer items-center gap-3"
                    >
                      <CheckBox
                        checked={isSelected}
                        readOnly
                        styles="outline-none"
                      />
                      <div className="flex min-w-0 flex-1 items-center gap-2">
                        <Avatar
                          img={option.img}
                          initials={
                            option.initials || option.label?.substring(0, 2)
                          }
                          styles="!size-6 !text-[10px] shrink-0 !rounded-full"
                        />
                        <span className="mb-0.5 truncate text-sm">
                          {option.label}
                        </span>
                      </div>
                    </div>
                  </li>
                );
              })}
            </>
          )}
        </ul>
      </PopoverContent>
    </Popover>
  );
}
