import { Delete02Icon, PlusSignIcon } from "@hugeicons/core-free-icons";
import { Icon, Select, Switch } from "@reservations/components";
import { generateTimeOptions } from "@reservations/lib";

const days = {
  1: "Monday",
  2: "Tuesday",
  3: "Wednesday",
  4: "Thursday",
  5: "Friday",
  6: "Saturday",
  0: "Sunday",
};

export default function BusinessHours({ data, setBusinessHours, preferences }) {
  const timeOptions = generateTimeOptions(preferences?.time_format);
  const dayOrder =
    preferences?.first_day_of_week === "Sunday"
      ? [0, 1, 2, 3, 4, 5, 6]
      : [1, 2, 3, 4, 5, 6, 0];

  const bhArray = dayOrder.map((day) => {
    const timeSlots = data?.[day] ?? [];

    return {
      day: day,
      isOpen: timeSlots.length > 0,
      timeSlots,
    };
  });

  const toggleDay = (day) => {
    setBusinessHours((prevHours) => {
      return {
        ...prevHours,
        [day]: prevHours?.[day]?.length
          ? []
          : [{ start_time: "09:00", end_time: "17:00" }],
      };
    });
  };

  function addTimePeriod(day) {
    setBusinessHours((prevHours) => ({
      ...prevHours,
      [day]: [
        ...(prevHours?.[day] || []),
        { start_time: "09:00", end_time: "17:00" },
      ],
    }));
  }

  function removeTimePeriod(day, timeSlotIndex) {
    setBusinessHours((prevHours) => {
      const newHours = { ...prevHours };
      if (newHours[day]) {
        newHours[day] = newHours[day].filter((_, i) => i !== timeSlotIndex);
      }
      return newHours;
    });
  }

  function updateTime(day, timeSlotIndex, field, value) {
    setBusinessHours((prevHours) => {
      const newHours = { ...prevHours };
      newHours[day] = [...newHours[day]];

      let updatedSlot = { ...newHours[day][timeSlotIndex], [field]: value };

      if (field === "start_time") {
        const newStartTime = value;
        const currentEndTime = newHours[day][timeSlotIndex].end_time;

        if (!currentEndTime || currentEndTime <= newStartTime) {
          const nextValidTime = timeOptions.find(
            (option) => option.value > newStartTime
          )?.value;

          updatedSlot.end_time = nextValidTime;
        }
      }

      newHours[day][timeSlotIndex] = updatedSlot;
      return newHours;
    });
  }

  return (
    <div
      className="border-border bg-bg-raised shadow-card flex max-w-xl flex-col
        gap-3 rounded-xl border px-3 py-4 sm:px-4"
    >
      {bhArray.map((day) => (
        <div
          key={day.day}
          className={`flex flex-col items-start
          ${day.isOpen ? "gap-3" : "gap-0 pb-5"} border-b-border-subtle border-b
          pb-3 last:border-b-0 last:pb-0 lg:flex-row lg:gap-6`}
        >
          <div className="flex items-center gap-10 md:mt-2 md:gap-20">
            <label className="inline-flex cursor-pointer items-center gap-3">
              <Switch
                onSwitch={() => toggleDay(day.day)}
                size="medium"
                defaultValue={day.isOpen}
              />
              <span className="w-20 text-sm font-medium">{days[day.day]}</span>
            </label>

            {!day.isOpen && (
              <div className="text-fg-muted text-sm">
                <span>Closed</span>
              </div>
            )}
          </div>
          <div className="flex w-full flex-1 flex-col gap-2">
            {day.isOpen &&
              day.timeSlots.map((timeSlot, timeSlotIndex) => (
                <div key={timeSlotIndex} className="flex items-center gap-2">
                  <Select
                    options={timeOptions.filter(
                      (option) => option.value !== "23:30"
                    )}
                    value={timeSlot.start_time}
                    onSelect={(option) =>
                      updateTime(
                        day.day,
                        timeSlotIndex,
                        "start_time",
                        option.value
                      )
                    }
                    styles="flex-1 lg:min-w-28"
                    maxVisibleItems={7}
                  />
                  <span className="text-fg-muted text-sm">to</span>
                  <Select
                    options={timeOptions.filter(
                      (option) => option.value > timeSlot.start_time
                    )}
                    value={timeSlot.end_time}
                    onSelect={(option) =>
                      updateTime(
                        day.day,
                        timeSlotIndex,
                        "end_time",
                        option.value
                      )
                    }
                    styles="flex-1 lg:min-w-28"
                    maxVisibleItems={7}
                  />
                  {timeSlotIndex === 0 ? (
                    <button
                      onClick={() => addTimePeriod(day.day)}
                      className={`${
                        day.timeSlots.length >= 2
                          ? "text-fg-disabled"
                          : `text-fg-muted hover:bg-hover hover:text-fg
                            cursor-pointer`
                        } rounded-lg p-2`}
                      disabled={day.timeSlots.length >= 2}
                    >
                      <Icon icon={PlusSignIcon} styles="size-4" />
                    </button>
                  ) : (
                    <button
                      onClick={() => removeTimePeriod(day.day, timeSlotIndex)}
                      className="hover:bg-danger-subtle cursor-pointer
                        rounded-lg p-2"
                    >
                      <Icon
                        icon={Delete02Icon}
                        styles="size-4 text-fg-danger"
                      />
                    </button>
                  )}
                </div>
              ))}
          </div>
        </div>
      ))}
    </div>
  );
}
