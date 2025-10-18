import Select from "@components/Select";
import {
  BOOKING_WINDOW_MAX_OPTIONS,
  BOOKING_WINDOW_MIN_OPTIONS,
  BUFFER_TIME_OPTIONS,
  CANCEL_DEADLINE_OPTIONS,
} from "@lib/constants";

export default function SchedulingSettings({ settings, onChange }) {
  return (
    <>
      <div className="flex flex-col gap-8">
        <div className="flex flex-col gap-5">
          <div className="text-text_color text-lg font-medium">
            Cancellation Policy
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-text_color block text-sm font-medium">
              Minimum time required for cancellation
            </label>
            <Select
              options={CANCEL_DEADLINE_OPTIONS}
              value={settings.cancel_deadline}
              onSelect={(option) =>
                onChange({ name: "cancel_deadline", value: option.value })
              }
              styles="w-full sm:w-2/3"
            />
            <p className="text-text_color/70 text-sm">
              Customers must cancel at least this amount of time before their
              appointment
            </p>
          </div>
        </div>

        <div className="flex flex-col gap-4">
          <div>
            <div className="text-text_color text-lg font-medium">
              Booking Window
            </div>
            <p className="text-text_color/70 mt-1 text-sm">
              Control when customers can book appointments
            </p>
          </div>

          <div className="flex w-full flex-col gap-8 md:flex-row">
            <div className="flex flex-col gap-2">
              <label className="text-text_color block text-sm font-medium">
                Minimum advance booking time
              </label>
              <Select
                options={BOOKING_WINDOW_MIN_OPTIONS}
                value={settings.booking_window_min}
                onSelect={(option) =>
                  onChange({ name: "booking_window_min", value: option.value })
                }
                styles="w-full"
              />
              <p className="text-text_color/70 text-sm">
                Customers must book at least this far in advance
              </p>
            </div>

            <div className="flex flex-col gap-2">
              <label className="text-text_color block text-sm font-medium">
                Maximum advance booking time
              </label>
              <Select
                options={BOOKING_WINDOW_MAX_OPTIONS}
                value={settings.booking_window_max}
                onSelect={(option) =>
                  onChange({ name: "booking_window_max", value: option.value })
                }
                styles="w-full"
              />
              <p className="text-text_color/70 text-sm">
                Customers can book appointments up to this far ahead
              </p>
            </div>
          </div>
        </div>

        <div className="flex flex-col gap-5">
          <div className="text-text_color text-lg font-medium">
            Schedule Management
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-text_color block text-sm font-medium">
              Buffer time between appointments
            </label>
            <Select
              options={BUFFER_TIME_OPTIONS}
              value={settings.buffer_time}
              onSelect={(option) =>
                onChange({ name: "buffer_time", value: option.value })
              }
              styles="w-full sm:w-2/3"
            />
            <p className="text-text_color/70 text-sm">
              Automatic break time added between consecutive appointments
            </p>
          </div>
        </div>
      </div>
    </>
  );
}
