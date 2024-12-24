import BackArrowIcon from "@icons/BackArrowIcon";
import { useClickOutside } from "@lib/hooks";
import { Children, cloneElement, useRef, useState } from "react";

export default function Selector({
  children,
  defaultValue,
  styles,
  dropdownStyles,
  onSelect,
}) {
  const selectorRef = useRef();
  const [value, setValue] = useState(defaultValue);
  const [showOptions, setShowOptions] = useState(false);
  useClickOutside(selectorRef, () => setShowOptions(false));

  function handleSelect(newValue) {
    setValue(newValue);
    setShowOptions(false);
    onSelect(newValue);
  }

  return (
    <div
      ref={selectorRef}
      className="relative flex w-full flex-col transition-all"
    >
      <button
        className={`relative inline-flex w-full flex-shrink-0 items-center ${styles}
          ${showOptions ? "border-[1px] border-gray-600 bg-hvr_gray" : ""}`}
        type="button"
        onClick={() => {
          setShowOptions(!showOptions);
        }}
      >
        <div className="flex w-full flex-nowrap justify-between gap-1 p-1">
          <span className="overflow-hidden text-nowrap">{value}</span>
          <BackArrowIcon styles="-rotate-90 w-5 h-5" />
        </div>
      </button>
      <ul
        className={`overflow-y-auto text-nowrap dark:[color-scheme:dark]
          ${showOptions ? "block" : "hidden"} ${dropdownStyles} `}
      >
        {Children.map(children, (child) =>
          cloneElement(child, {
            onClick: () => handleSelect(child.props.value),
          })
        )}
      </ul>
    </div>
  );
}
