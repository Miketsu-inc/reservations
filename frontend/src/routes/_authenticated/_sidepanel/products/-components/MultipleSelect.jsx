import Button from "@components/Button";
import Select from "@components/Select";
import XIcon from "@icons/XIcon";
import { useEffect, useState } from "react";

export default function MultipleSelect({
  options,
  onAddition,
  onDeletion,
  initialItems = [],
  styles,
  setDefault,
}) {
  const [selectedItems, setSelectedItems] = useState(initialItems);
  const [currentSelection, setCurrentSelection] = useState(null);

  useEffect(() => {
    if (initialItems.length > 0) {
      const selectedServices = initialItems.map((id) =>
        options.find((option) => option.value === id)
      );
      setSelectedItems(selectedServices);
    }
  }, [options, initialItems]);

  useEffect(() => {
    if (setDefault) {
      setSelectedItems([]);
      setCurrentSelection();
    }
  }, [setDefault]);

  function handleAdd() {
    if (
      currentSelection &&
      !selectedItems.some((item) => item.value === currentSelection.value)
    ) {
      const newItem = currentSelection;
      const newItems = [...selectedItems, newItem];

      setSelectedItems(newItems);
      onAddition?.(newItem);
      setCurrentSelection(null);
    }
  }

  function handleRemove(item) {
    const newItems = selectedItems.filter(
      (selected) => selected.value !== item.value
    );
    setSelectedItems(newItems);
    onDeletion?.(item);
  }

  return (
    <div className={`flex flex-col gap-2 ${styles}`}>
      {/* Select and Add Button */}
      <div className="flex items-center gap-2">
        <Select
          options={options.filter(
            (option) =>
              !selectedItems.some((item) => item.value === option.value)
          )}
          value={currentSelection?.value}
          onSelect={(item) => setCurrentSelection(item)}
          placeholder="Select a service"
          styles="w-full"
          emptyText="No more services found"
        />
        <Button
          variant="primary"
          onClick={handleAdd}
          disabled={!currentSelection}
          buttonText="Add"
          styles="p-2"
        />
      </div>

      {selectedItems.length > 0 && (
        <div
          className="mt-1 flex gap-2 overflow-x-auto scroll-smooth rounded-lg pb-2 outline-none
            md:max-h-24 md:flex-wrap md:overflow-y-auto dark:[color-scheme:dark]"
        >
          {selectedItems.map((item) =>
            item?.value ? (
              <div
                key={item.value}
                className="bg-hvr_gray flex max-w-44 items-center gap-2 rounded-full px-3 py-1 text-sm
                  md:max-w-36"
              >
                {item.icon && item.icon}
                <span className="text-text_color truncate">{item.label}</span>
                <XIcon
                  styles="h-5 w-5 fill-text_color cursor-pointer"
                  onClick={() => handleRemove(item)}
                />
              </div>
            ) : undefined
          )}
        </div>
      )}
    </div>
  );
}
