import Select from "@components/Select";

const cancelOptions = [
  { value: 0, label: "Anytime" },
  { value: 30, label: "30 minutes" },
  { value: 60, label: "1 hour" },
  { value: 120, label: "2 hours" },
  { value: 180, label: "3 hours" },
  { value: 240, label: "4 hours" },
  { value: 300, label: "5 hours" },
  { value: 360, label: "6 hours" },
  { value: 720, label: "12 hours" },
  { value: 1440, label: "1 day" },
  { value: 2880, label: "2 days" },
  { value: 5760, label: "3 days" },
];

const bookingOptions = [
  { value: 0, label: "Anytime" },
  { value: 15, label: "15 minutes" },
  { value: 30, label: "30 minutes" },
  { value: 60, label: "1 hour" },
  { value: 120, label: "2 hours" },
  { value: 180, label: "3 hours" },
  { value: 240, label: "4 hours" },
  { value: 300, label: "5 hours" },
  { value: 360, label: "6 hours" },
  { value: 720, label: "12 hours" },
  { value: 1440, label: "1 day" },
  { value: 2880, label: "2 days" },
  { value: 4320, label: "3 days" },
  { value: 5760, label: "4 days" },
  { value: 7200, label: "5 days" },
  { value: 8640, label: "6 days" },
  { value: 10080, label: "1 week" },
  { value: 20160, label: "2 weeks" },
];

const bookAheadOptions = [
  { value: 1, label: "1 month" },
  { value: 2, label: "2 months" },
  { value: 3, label: "3 months" },
  { value: 4, label: "4 months" },
  { value: 5, label: "5 months" },
  { value: 6, label: "6 months" },
  { value: 7, label: "7 months" },
  { value: 8, label: "8 months" },
  { value: 9, label: "9 months" },
  { value: 10, label: "10 months" },
  { value: 11, label: "11 months" },
  { value: 12, label: "1 year" },
];

const bufferOptions = [
  { value: 0, label: "No buffer time" },
  { value: 5, label: "5 minutes" },
  { value: 10, label: "10 minutes" },
  { value: 15, label: "15 minutes" },
  { value: 20, label: "20 minutes" },
  { value: 25, label: "25 minutes" },
  { value: 30, label: "30 minutes" },
];

export default function BookingSettings({ settings, onChange }) {
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
              options={cancelOptions}
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
                options={bookingOptions}
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
                options={bookAheadOptions}
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
              options={bufferOptions}
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
