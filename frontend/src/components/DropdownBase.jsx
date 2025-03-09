import BackArrowIcon from "@icons/BackArrowIcon";
import TickIcon from "@icons/TickIcon";
import { useClickOutside } from "@lib/hooks";
import { useEffect, useRef, useState } from "react";

export default function DropdownBase({
  options,
  value,
  onSelect,
  placeholder,
  styles,
  maxVisibleItems = 7,
  extraContent,
  onClose,
emptyText,
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [highlightedIndex, setHighlightedIndex] = useState(null);
  const [itemHeight, setItemHeight] = useState(0);
  const [isUsingKeyboard, setIsUsingKeyboard] = useState(false);
  const containerRef = useRef(null);
  const dropDownListRef = useRef(null);
  const listElementRef = useRef(null);

  const selectedIndex = options?.findIndex((option) => option.value === value);
  const selectedOption = options?.[selectedIndex];

  useClickOutside(containerRef, () => {
    handleClose();
  });

  function handleOpen() {
    setIsOpen(true);
    setHighlightedIndex(selectedIndex > -1 ? selectedIndex : 0);
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
    if (isOpen && selectedIndex > -1 && dropDownListRef.current?.children) {
      dropDownListRef.current.children[selectedIndex].scrollIntoView({
        block: "nearest", //when center scrolling weird shit happens
        behavior: "smooth",
      });
    }
  }, [isOpen, selectedIndex]);

  useEffect(() => {
    if (
      isOpen &&
      isUsingKeyboard &&
      dropDownListRef.current?.children[highlightedIndex]
    ) {
      dropDownListRef.current.children[highlightedIndex].scrollIntoView({
        block: "nearest",
      });
    }
  }, [highlightedIndex, isUsingKeyboard, isOpen]);

  useEffect(() => {
    if (isOpen && listElementRef.current) {
      const height = listElementRef.current.getBoundingClientRect().height;
      setItemHeight(height);
    }
  }, [isOpen]);

  return (
    <div
      className={`${styles} relative`}
      ref={containerRef}
      onKeyDown={handleKeyDown}
    >
      <button
        onClick={() => (isOpen ? handleClose() : handleOpen())}
        className="focus:border-text_color w-full rounded-md border border-gray-400 px-4 py-2
          text-left text-gray-900 focus:outline-none dark:border-gray-500
          dark:bg-neutral-950 dark:focus:border-white"
        type="button"
      >
        <div className="flex items-center justify-between">
          <span
            className={`${selectedOption ? "text-text_color" : "text-gray-500"} min-h-6 flex-1 truncate`}
          >
            {!selectedOption ? (
              placeholder
            ) : selectedOption.icon ? (
              <span className="flex items-center gap-2">
                <span className="shrink-0">{selectedOption.icon}</span>
                <span className="truncate">{selectedOption.label}</span>
              </span>
            ) : (
              selectedOption.label
            )}
          </span>
          <BackArrowIcon
            styles={`dark:stroke-gray-300 stroke-gray-500 transition-transform -rotate-90 shrink-0
              ${isOpen ? "rotate-90" : ""} h-5 w-5`}
          />
        </div>
      </button>
      {isOpen && (
        <div
          className={`z-10 ${
          window.innerHeight -
              containerRef.current.getBoundingClientRect().bottom <
            itemHeight * maxVisibleItems + (extraContent ? 48 : 0) // height of the search input
              ? "bottom-full mb-1"
              : "top-full mt-1"
          } absolute flex w-full flex-col rounded-md border border-gray-300 bg-white
          shadow-lg dark:border-gray-500 dark:bg-neutral-950`}
        >
          {extraContent}
          <ul
            ref={dropDownListRef}
            style={{
              maxHeight: itemHeight
                ? `${itemHeight * maxVisibleItems + 8}px`
                : "auto",
            }}
            className="overflow-x-hidden overflow-y-auto p-1 transition-all dark:[color-scheme:dark]"
            onMouseMove={() => {
              setIsUsingKeyboard(false);
              setHighlightedIndex(null);
            }}
          >
            {options.length === 0 ? (
              <li className="px-4 py-6 text-center text-gray-500 select-none dark:text-gray-400">
{emptyText || "No results found"}
              </li>
            ) : (
              options.map((option, index) => {
                const isSelected = value === option.value;
                const isHighlighted = index === highlightedIndex;

                return (
                  <li
                    ref={index === 0 ? listElementRef : null}
                    onClick={() => {
                      onSelect(option);
                      handleClose();
                    }}
                    key={index}
                    className={`${isHighlighted ? "bg-hvr_gray" : isUsingKeyboard ? "" : "hover:bg-hvr_gray"}
                      dark:text-text_color cursor-pointer rounded-sm py-1 pr-3 pl-3 text-gray-700
                      select-none`}
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
                        <TickIcon styles="h-6 w-6 fill-text_color shrink-0" />
                      )}
                    </div>
                  </li>
                );
              })
            )}
          </ul>
        </div>
      )}
    </div>
  );
}
