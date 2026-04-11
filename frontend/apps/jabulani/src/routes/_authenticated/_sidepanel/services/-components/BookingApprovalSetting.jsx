import { BackArrowIcon } from "@reservations/assets";
import { Card, Switch } from "@reservations/components";
import { useState } from "react";

const approvalOptions = [
  {
    value: "auto",
    name: "Automatic",
    desc: "All bookings confirmed instantly, no action needed.",
  },
  {
    value: "manual",
    name: "Manual",
    desc: "Every request waits for your approval before confirming.",
  },
  {
    value: "manual_for_new",
    name: "Manual for new customers",
    desc: "Returning customers auto-approved, new ones need review.",
  },
];

export default function BookingApprovalSetting({ onUpdate, settings }) {
  const [isOpen, setIsOpen] = useState(false);
  const isOverridden = settings.approval_policy !== null;
  const [showCustomSettings, setShowCustomSettings] = useState(isOverridden);

  function handleSwitch() {
    if (showCustomSettings) {
      onUpdate({
        settings: {
          ...settings,
          approval_policy: null,
        },
      });
    }
    setShowCustomSettings(!showCustomSettings);
  }

  return (
    <Card styles="p-0! flex flex-col">
      <div
        role="button"
        onClick={() => setIsOpen(!isOpen)}
        className={`${isOpen ? "border-border_color border-b" : ""} flex
          cursor-pointer items-center justify-between p-4`}
      >
        <div className="flex items-center justify-center gap-2">
          <p className="text-lg">Other</p>
        </div>
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          className="hover:bg-hvr_gray cursor-pointer rounded-lg p-2"
        >
          <BackArrowIcon
            styles={`size-6 stroke-text_color transition-transform
              ${isOpen ? "rotate-90" : "-rotate-90"}`}
          />
        </button>
      </div>
      <div
        className={`flex flex-col gap-6 px-4 transition-[max-height,opacity]
          duration-200 ease-in-out ${
            isOpen
              ? "max-h-250 pb-4 opacity-100"
              : "max-h-0 overflow-hidden opacity-0"
          }`}
      >
        <div className="flex flex-col gap-4 pt-4">
          <div className="flex items-center gap-4">
            <Switch defaultValue={isOverridden} onSwitch={handleSwitch} />
            Override your bookng approval policy for this service
          </div>
        </div>

        <div
          className={`grid grid-cols-1 gap-6 transition-[max-height,opacity]
            ease-in-out lg:grid-cols-2 ${
              showCustomSettings
                ? "max-h-250 pb-4 opacity-100"
                : "max-h-0 overflow-hidden opacity-0"
            }`}
        >
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
            {approvalOptions.map((option) => {
              const active = settings.approval_policy === option.value;
              return (
                <button
                  key={option.value}
                  onClick={() =>
                    onUpdate({
                      settings: { ...settings, approval_policy: option.value },
                    })
                  }
                  className={`flex flex-col gap-1 rounded-md border p-4
                  text-left transition-colors duration-150 ${
                    active
                      ? "border-primary bg-primary/5"
                      : `bg-layer_bg border-gray-300 hover:border-gray-300
                        hover:bg-gray-100 dark:border-gray-500
                        dark:hover:border-gray-400 dark:hover:bg-gray-600/5`
                  }`}
                >
                  <span className={"text-text_color text-sm font-medium"}>
                    {option.name}
                  </span>
                  <span
                    className={
                      "text-xs leading-relaxed text-gray-500 dark:text-gray-400"
                    }
                  >
                    {option.desc}
                  </span>
                </button>
              );
            })}
          </div>
        </div>
      </div>
    </Card>
  );
}
