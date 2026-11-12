import { ArrowLeft01Icon } from "@hugeicons/core-free-icons";
import { useEffect, useRef, useState } from "react";
import Avatar from "./Avatar.jsx";
import AvatarStack from "./AvatarStack.jsx";
import CheckBox from "./CheckBox.jsx";
import Icon from "./Icon.jsx";
import { Popover, PopoverContent, PopoverTrigger } from "./Popover";

const itemHeight = 36;

export default function MultiSelect({
  options = [],
  values = [],
  onSelect,
  placeholder,
  labelText,
  displayText = "item",
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

  const selectedValues = new Set(values);
  const allSelected =
    options.length > 0 &&
    selectedValues.size === options.length &&
    options.every((option) => selectedValues.has(option.value));

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
    if (selectedValues.has(optionValue)) {
      onSelect(values.filter((id) => id !== optionValue));
    } else {
      onSelect([...values, optionValue]);
    }
  }

  function handleKeyDown(e) {
    // options + 1 for the 'All' option and - 1 for 0-index
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
      <div className={`w-full min-w-0 ${styles}`}>
        {labelText && (
          <span className="text-fg flex items-center gap-1 pb-1.5 text-sm">
            {labelText}
            {required !== false && (
              <span
                aria-hidden="true"
                className="text-fg-danger text-base leading-none"
              >
                *
              </span>
            )}
          </span>
        )}
        <PopoverTrigger asChild>
          <button
            className={`${styles} bg-bg-input border-border-strong
              hover:not-disabled:not-focus-visible:not-data-popup-open:border-border-hover
              focus-visible:border-primary focus-visible:ring-primary-focus
              data-popup-open:border-primary data-popup-open:ring-primary-focus
              disabled:border-border disabled:bg-bg-sunken h-[38px] w-full
              min-w-0 rounded-lg border px-3 text-left outline-hidden
              transition-[border-color,box-shadow] duration-150
              focus-visible:ring-3 data-popup-open:ring-3 md:text-sm`}
            type="button"
            ref={containerRef}
            disabled={disabled}
          >
            <div
              className="flex w-full min-w-0 items-center justify-between gap-2"
            >
              <div className="flex min-w-0 flex-1 items-center gap-2">
                <TriggerContent
                  selectedValues={selectedValues}
                  options={options}
                  placeholder={placeholder}
                  displayText={displayText}
                  disabled={disabled}
                  allSelected={allSelected}
                />
              </div>
              <Icon
                icon={ArrowLeft01Icon}
                styles={`text-fg-muted transition-transform -rotate-90 shrink-0
                  ${isOpen ? "rotate-90" : ""} size-5`}
              />
            </div>
          </button>
        </PopoverTrigger>
      </div>
      <PopoverContent
        styles="p-0!"
        onKeyDown={handleKeyDown}
        style={{
          width: triggerWidth || "auto",
        }}
      >
        <ul
          role="listbox"
          aria-multiselectable="true"
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
            <li className="text-fg-muted px-4 py-6 text-center select-none">
              {emptyText || "No results found"}
            </li>
          ) : (
            <>
              <li
                onClick={handleToggleAll}
                className={`${
                  highlightedIndex === 0 ? "bg-hover" : "hover:bg-hover"
                }
                  text-fg mb-1 cursor-pointer rounded-sm py-2 pr-1 pl-2
                  select-none`}
                role="option"
                aria-selected={allSelected}
              >
                <div className="flex w-full items-center gap-3">
                  <CheckBox
                    checked={allSelected}
                    readOnly
                    styles="outline-none"
                  />
                  <span className="text-sm font-medium">
                    All {displayText}s
                  </span>
                </div>
              </li>

              {options.map((option, index) => {
                const listIndex = index + 1;
                const isSelected = selectedValues.has(option.value);
                const isHighlighted = highlightedIndex === listIndex;

                return (
                  <li
                    key={option.value}
                    onClick={() => handleToggleOption(option.value)}
                    className={`${isHighlighted ? "bg-hover" : isUsingKeyboard ? "" : "hover:bg-hover"}
                      text-fg cursor-pointer rounded-sm py-2 pr-0.5 pl-2
                      select-none`}
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
                        {(option.img || option.initials) && (
                          <Avatar
                            img={option.img}
                            initials={option.initials}
                            styles="size-6! text-[10px]! shrink-0"
                          />
                        )}
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

function TriggerContent({
  options,
  selectedValues,
  placeholder,
  displayText,
  disabled,
  allSelected,
}) {
  const selectedOptions = options.filter((option) =>
    selectedValues.has(option.value)
  );

  if (selectedOptions.length === 0) {
    return (
      <span
        className={`${disabled ? "text-fg-disabled" : "text-fg-muted"} flex-1
          truncate text-sm`}
      >
        {placeholder}
      </span>
    );
  }

  let text = `${selectedOptions.length} ${displayText}s selected`;
  if (allSelected) {
    text = `All ${displayText}s`;
  } else if (selectedOptions.length === 1) {
    text = selectedOptions[0].label;
  }

  const avatars = selectedOptions
    .filter((option) => option.img || option.initials)
    .map((option) => ({
      id: option.value,
      img: option.img,
      initials: option.initials,
      name: option.label,
    }));

  return (
    <div className="flex min-w-0 flex-1 items-center gap-2">
      {avatars.length > 0 && <AvatarStack avatars={avatars} max={2} />}
      <span
        className={`truncate text-sm ${
          disabled ? "text-fg-disabled" : "text-fg"
        }`}
      >
        {text}
      </span>
    </div>
  );
}
