import {
  Button,
  CloseButton,
  ResponsiveDialog,
  ResponsiveDialogClose,
  ResponsiveDialogContent,
} from "@reservations/components";
import { useState } from "react";

const options = [
  {
    id: "this",
    label: "This booking",
    description: "Only this occurrence will be updated.",
  },
  {
    id: "future",
    label: "All future occurrences",
    description: "This and all future occurrences will be updated.",
  },
];

export default function UpdateRecurringModal({ onClose, isOpen, onSave }) {
  const [selected, setSelected] = useState("this");
  return (
    <ResponsiveDialog
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <ResponsiveDialogContent styles="w-full p-5 flex flex-col gap-5">
        <div
          className="border-border flex items-start justify-between gap-4
            border-b pb-3"
        >
          <p className="text-fg text-lg font-semibold">
            Edit recurring booking
          </p>

          <ResponsiveDialogClose asChild>
            <CloseButton styles="size-4 hidden lg:block" />
          </ResponsiveDialogClose>
        </div>
        <div className="flex w-full flex-col gap-3 lg:flex-row">
          {options.map((opt) => {
            const active = selected === opt.id;
            return (
              <button
                key={opt.id}
                onClick={() => setSelected(opt.id)}
                className={`flex flex-1 cursor-pointer flex-col gap-2 rounded-lg
                border px-4 py-4 text-left transition-all ${
                  active
                    ? "border-primary bg-primary-subtle "
                    : "border-border-strong hover:bg-hover "
                }`}
              >
                <span className="flex flex-col gap-1">
                  <span className={"text-fg text-sm font-semibold"}>
                    {opt.label}
                  </span>
                  <span className="text-fg-muted text-xs leading-relaxed">
                    {opt.description}
                  </span>
                </span>
              </button>
            );
          })}
        </div>
        <div className="flex w-full justify-end gap-2">
          <ResponsiveDialogClose asChild>
            <Button
              variant="tertiary"
              styles="py-1 px-3 hidden lg:block"
              buttonText="Cancel"
            />
          </ResponsiveDialogClose>

          <Button
            variant="primary"
            onClick={() => onSave(selected)}
            styles="py-1 px-4 w-full lg:w-auto"
            buttonText="Save Changes"
          />
        </div>
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}
