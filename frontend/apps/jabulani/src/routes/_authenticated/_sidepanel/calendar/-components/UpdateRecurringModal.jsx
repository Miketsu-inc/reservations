import { Button, CloseButton, Modal } from "@reservations/components";
import { useState } from "react";

const options = [
  {
    id: "this",
    label: "This booking",
    description: "Only this occurrence will be updated.",
  },
  {
    id: "future",
    label: "All future occurences",
    description: "This and all future occurrences will be updated.",
  },
];

export default function UpdateRecurringModal({ onClose, isOpen, onSave }) {
  const [selected, setSelected] = useState("this");
  return (
    <Modal
      styles="w-full sm:w-80 p-5 flex flex-col gap-5"
      isOpen={isOpen}
      onClose={onClose}
      zindex={60}
      disableFocusTrap={true}
    >
      <div
        className="border-border_color flex items-start justify-between gap-4
          border-b pb-3"
      >
        <p className="text-lg font-semibold text-gray-900 dark:text-gray-100">
          Edit recurring booking
        </p>

        <CloseButton onClick={onClose} styles="size-4" />
      </div>
      <div className="flex flex-col gap-3 sm:flex-row">
        {options.map((opt) => {
          const active = selected === opt.id;
          return (
            <button
              key={opt.id}
              onClick={() => setSelected(opt.id)}
              className={`flex flex-1 cursor-pointer flex-col gap-2 rounded-lg
              border px-4 py-4 text-left transition-all ${
                active
                  ? "border-primary bg-primary/5 "
                  : `border-input_border_color hover:bg-gray-50
                    dark:hover:bg-gray-700/10`
              }`}
            >
              <span className="flex flex-col gap-1">
                <span className={"text-text_color/90 text-sm font-semibold"}>
                  {opt.label}
                </span>
                <span
                  className="text-xs leading-relaxed text-gray-500
                    dark:text-gray-400"
                >
                  {opt.description}
                </span>
              </span>
            </button>
          );
        })}
      </div>
      <div className="flex justify-end gap-2">
        <Button
          variant="tertiary"
          onClick={onClose}
          styles="py-1 px-3"
          buttonText="Cancel"
        />

        <Button
          variant="primary"
          onClick={() => onSave(selected)}
          styles="py-1 px-4"
          buttonText="Save Changes"
        />
      </div>
    </Modal>
  );
}
