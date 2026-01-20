import { BackArrowIcon, CalendarIcon } from "@reservations/assets";
import { Card, Select, Switch } from "@reservations/components";
import {
  BOOKING_WINDOW_MAX_OPTIONS,
  BOOKING_WINDOW_MIN_OPTIONS,
  BUFFER_TIME_OPTIONS,
  CANCEL_DEADLINE_OPTIONS,
} from "@reservations/lib";
import { Link } from "@tanstack/react-router";
import { useState } from "react";

export default function ServiceSchedulingSettings({ onUpdate, settings }) {
  const [isOpen, setIsOpen] = useState(false);
  const areSettingsNull = Object.values(settings).every(
    (value) => value === null
  );
  const [showCustomSettings, setShowCustomSettings] =
    useState(!areSettingsNull);

  function handleSwitch() {
    if (showCustomSettings) {
      onUpdate({
        settings: {
          cancel_deadline: null,
          buffer_time: null,
          booking_window_max: null,
          booking_window_min: null,
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
          <CalendarIcon styles="size-6 mb-0.5 text-text_color" />
          <p className="text-lg">Scheduling</p>
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
            <span className="font-medium">Use Custom Settings</span>
            <Switch defaultValue={!areSettingsNull} onSwitch={handleSwitch} />
          </div>
          <div className="text-text_color/70 text-sm">
            Create custom scheduling rules for this service. These settings will
            override your{" "}
            <Link
              to="/settings/merchant"
              className="text-primary hover:text-primary/80 font-medium
                hover:underline"
            >
              global settings
            </Link>
            .
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
          <div className="flex flex-col gap-2">
            <Select
              options={CANCEL_DEADLINE_OPTIONS}
              labelText="Minimum cancellation notice"
              required={false}
              value={settings.cancel_deadline}
              onSelect={(option) =>
                onUpdate({
                  settings: { ...settings, cancel_deadline: option.value },
                })
              }
              styles="w-full"
              placeholder="Select cancel deadline"
            />
            <p className="text-text_color/70 text-sm">
              Required notice time for cancellations
            </p>
          </div>

          <div className="flex flex-col gap-2">
            <Select
              options={BUFFER_TIME_OPTIONS}
              labelText="Buffer time between appointments"
              required={false}
              value={settings.buffer_time}
              onSelect={(option) =>
                onUpdate({
                  settings: { ...settings, buffer_time: option.value },
                })
              }
              styles="w-full"
              placeholder="Select buffer time"
            />
            <p className="text-text_color/70 text-sm">
              Break time added between consecutive bookings
            </p>
          </div>

          <div className="flex flex-col gap-2">
            <Select
              options={BOOKING_WINDOW_MIN_OPTIONS}
              labelText="Minimum advance booking"
              required={false}
              value={settings.booking_window_min}
              onSelect={(option) =>
                onUpdate({
                  settings: { ...settings, booking_window_min: option.value },
                })
              }
              styles="w-full"
              placeholder="Select minimum time"
            />
            <p className="text-text_color/70 text-sm">
              Earliest time customers can book ahead
            </p>
          </div>

          <div className="flex flex-col gap-2">
            <Select
              options={BOOKING_WINDOW_MAX_OPTIONS}
              labelText="Maximum advance booking"
              required={false}
              value={settings.booking_window_max}
              onSelect={(option) =>
                onUpdate({
                  settings: { ...settings, booking_window_max: option.value },
                })
              }
              styles="w-full"
              placeholder="Select maximum time"
            />
            <p className="text-text_color/70 text-sm">
              Latest time customers can book ahead
            </p>
          </div>
        </div>
      </div>
    </Card>
  );
}
