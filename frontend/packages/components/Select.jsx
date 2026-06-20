import { ArrowLeft01Icon, Tick02Icon } from "@hugeicons/core-free-icons";
import { memo, useEffect, useRef, useState } from "react";
import { Icon } from ".";
import { Popover, PopoverContent, PopoverTrigger } from "./Popover";

const itemHeight = 34;

export default function Select({
  options,
  allOptions,
  value,
  onSelect,
  placeholder,
  labelText,
  required,
  styles,
  maxVisibleItems = 7,
  extraContent,
  onClose,
  emptyText,
  onOpenChange,
  disabled,
  dropDownSameWidth = true,
  showOnlyIcon = false,
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [highlightedIndex, setHighlightedIndex] = useState(null);
  const [isUsingKeyboard, setIsUsingKeyboard] = useState(false);
  const [triggerWidth, setTriggerWidth] = useState(null);

  const containerRef = useRef(null);
  const dropDownListRef = useRef(null);

  const fullOptions = allOptions || options;
  const selectedOption = fullOptions?.find((option) => option.value === value);
  //index should be from the actually rendered options
  const selectedIndex = options?.findIndex((option) => option.value === value);

  function handleOpen() {
    setIsOpen(true);
    setHighlightedIndex(selectedIndex > -1 ? selectedIndex : 0);

    if (dropDownSameWidth && containerRef.current) {
      setTriggerWidth(containerRef.current.offsetWidth);
    }
  }

  function handleClose() {
    setIsOpen(false);
    onClose?.();
  }

  function handleKeyDown(e) {
    setIsUsingKeyboard(true);
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      if (!isOpen) {
        handleOpen();
      } else if (isOpen && highlightedIndex !== null) {
        onSelect(options[highlightedIndex]);
        handleClose();
      }
    } else if (e.key === "ArrowUp" && isOpen) {
      e.preventDefault();
      setHighlightedIndex((prev) => (prev > 0 ? prev - 1 : prev));
    } else if (e.key === "ArrowDown" && isOpen) {
      e.preventDefault();
      setHighlightedIndex((prev) =>
        prev < options.length - 1 ? prev + 1 : prev
      );
    }
  }

  useEffect(() => {
    if (!isOpen || selectedIndex < 0) return;

    const timeout = setTimeout(() => {
      if (dropDownListRef.current?.children) {
        dropDownListRef.current.children[selectedIndex]?.scrollIntoView({
          block: "center",
          behavior: "smooth",
        });
      }
    }, 0);

    return () => clearTimeout(timeout);
  }, [isOpen, selectedIndex]);

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
            <div className="flex items-center justify-between gap-2">
              <span
                className={`${selectedOption ? "text-text_color" : "text-gray-500"}
                  min-h-6 ${disabled ? "text-text_color/70" : ""}`}
              >
                {!selectedOption ? (
                  placeholder
                ) : selectedOption.icon ? (
                  <span className="flex items-center gap-2">
                    <span className={"shrink-0"}>{selectedOption.icon}</span>
                    {!showOnlyIcon && (
                      <span className="truncate">{selectedOption.label}</span>
                    )}
                  </span>
                ) : (
                  selectedOption.label
                )}
              </span>
              <Icon
                icon={ArrowLeft01Icon}
                styles={`text-gray-700 dark:text-gray-300 transition-transform
                  -rotate-90 shrink-0 ${isOpen ? "rotate-90" : ""} size-6`}
              />
            </div>
          </button>
        </label>
      </PopoverTrigger>
      <PopoverContent
        forceMount
        styles={`p-0! ${labelText && "data-[side=top]:translate-y-6"}`}
        align={dropDownSameWidth ? "center" : "start"}
        onKeyDown={handleKeyDown}
        style={{
          width: dropDownSameWidth ? triggerWidth : "auto",
        }}
      >
        {extraContent && <div className="p-2">{extraContent}</div>}
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
          {options?.length === 0 ? (
            <li
              className="px-4 py-6 text-center text-gray-500 select-none
                dark:text-gray-400"
            >
              {emptyText || "No results found"}
            </li>
          ) : (
            options?.map((option, index) => (
              <SelectItem
                key={index}
                option={option}
                isSelected={value === option.value}
                isHighlighted={index === highlightedIndex}
                isUsingKeyboard={isUsingKeyboard}
                onSelect={onSelect}
                handleClose={handleClose}
              />
            ))
          )}
        </ul>
      </PopoverContent>
    </Popover>
  );
}

const SelectItem = memo(
  ({
    option,
    isSelected,
    isHighlighted,
    isUsingKeyboard,
    onSelect,
    handleClose,
  }) => {
    return (
      <li
        onClick={() => {
          onSelect(option);
          handleClose();
        }}
        className={`${isHighlighted ? "bg-hvr_gray" : isUsingKeyboard ? "" : "hover:bg-hvr_gray"}
          dark:text-text_color cursor-pointer rounded-sm py-1 pr-0.5 pl-2
          text-gray-700 select-none`}
        role="option"
        aria-selected={isSelected}
      >
        <div className="flex w-full items-center justify-between">
          <div className="flex min-w-0 flex-1 items-center gap-2">
            {option.icon && (
              <span className="flex shrink-0 items-center justify-center">
                {option.icon}
              </span>
            )}
            <span className="truncate">{option.label}</span>
          </div>
          {isSelected && (
            <Icon icon={Tick02Icon} styles="size-6 text-text_color shrink-0" />
          )}
        </div>
      </li>
    );
  }
);

SelectItem.displayName = "SelectItem";
