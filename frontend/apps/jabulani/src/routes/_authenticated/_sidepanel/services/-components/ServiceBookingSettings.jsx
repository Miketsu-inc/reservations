import { Select, Switch } from "@reservations/components";
import {
  BOOKING_WINDOW_MAX_OPTIONS,
  BOOKING_WINDOW_MIN_OPTIONS,
  BUFFER_TIME_OPTIONS,
  CANCEL_DEADLINE_OPTIONS,
} from "@reservations/lib";
import { Link } from "@tanstack/react-router";
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

export default function ServiceBookingSettings({ onUpdate, settings }) {
  const areSchedulingSettingsNull = Object.values(settings).every(
    (value) => value === null
  );
  const isApprovalPolicyNull = settings.approval_policy === null;
  const [showScheduling, setShowScheduling] = useState(
    !areSchedulingSettingsNull
  );
  const [showApprovalPolicy, setShowApprovalPolicy] =
    useState(!isApprovalPolicyNull);

  function handleSchedulingSwitch() {
    if (showScheduling) {
      onUpdate({
        settings: {
          ...settings,
          cancel_deadline: null,
          buffer_time: null,
          booking_window_max: null,
          booking_window_min: null,
        },
      });
    }
    setShowScheduling(!showScheduling);
  }

  function handleApprovalPolicySwitch() {
    if (showApprovalPolicy) {
      onUpdate({
        settings: {
          ...settings,
          approval_policy: null,
        },
      });
    }
    setShowApprovalPolicy(!showApprovalPolicy);
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-4">
          <div className="flex items-center gap-4">
            <Switch
              defaultValue={!areSchedulingSettingsNull}
              onSwitch={handleSchedulingSwitch}
            />
            <span className="font-medium">
              Override booking scheduling settings
            </span>
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
              showScheduling
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
      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-4">
          <div className="flex items-center gap-4">
            <Switch
              defaultValue={!isApprovalPolicyNull}
              onSwitch={handleApprovalPolicySwitch}
            />
            Override bookng approval policy
          </div>
        </div>
        <div
          className={`grid gap-6 transition-[max-height,opacity] ease-in-out ${
            showApprovalPolicy
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
    </div>
  );
}
