import Select from "@components/Select";
import Switch from "@components/Switch";
import PlusIcon from "@icons/PlusIcon";
import TrashBinIcon from "@icons/TrashBinIcon";

const generateTimeOptions = () => {
  const options = [];
  for (let hour = 0; hour < 24; hour++) {
    for (let minute of [0, 30]) {
      const label = `${hour}:${minute === 0 ? "00" : "30"}`;
      const value = `${hour.toString().padStart(2, "0")}:${minute === 0 ? "00" : "30"}:00`;
      options.push({ label, value });
    }
  }
  return options;
};

const timeOptions = generateTimeOptions();

const days = {
  1: "Monday",
  2: "Tuesday",
  3: "Wednesday",
  4: "Thursday",
  5: "Friday",
  6: "Saturday",
  7: "Sunday",
};

export default function BusinessHours({ data, setBusinessHours }) {
  const bhArray = Object.keys(days).map((day) => ({
    day: parseInt(day),
    isOpen: data[day]?.length > 0,
    timeSlots: data[day],
  }));

  const toggleDay = (day) => {
    setBusinessHours((prevHours) => {
      return {
        ...prevHours,
        [day]: prevHours?.[day]?.length
          ? []
          : [{ start_time: "09:00:00", end_time: "17:00:00" }],
      };
    });
  };

  function addTimePeriod(day) {
    setBusinessHours((prevHours) => ({
      ...prevHours,
      [day]: [
        ...(prevHours?.[day] || []),
        { start_time: "09:00:00", end_time: "17:00:00" },
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
      newHours[day][timeSlotIndex] = {
        ...newHours[day][timeSlotIndex],
        [field]: value,
      };
      return newHours;
    });
  }

  return (
    <div
      className="flex max-w-xl flex-col gap-4 rounded border border-gray-300 p-4
        dark:border-gray-500"
    >
      {bhArray.map((day) => (
        <div
          key={day.day}
          className="flex flex-col items-start gap-3 border-b border-b-gray-300 pb-3 last:border-b-0
            last:pb-0 lg:flex-row lg:gap-6 dark:border-b-gray-500"
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
              <div className="text-gray-500">
                <span>Closed</span>
              </div>
            )}
          </div>
          <div className="flex w-full flex-1 flex-col gap-4">
            {day.isOpen &&
              day.timeSlots.map((timeSlot, timeSlotIndex) => (
                <div key={timeSlotIndex} className="flex items-center gap-4">
                  <Select
                    options={timeOptions}
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
                  <span className="text-gray-500">to</span>
                  <Select
                    options={timeOptions}
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
                      className="cursor-pointer rounded-full border border-gray-500 p-1"
                    >
                      <PlusIcon styles="h-4 w-4 text-text_color" />
                    </button>
                  ) : (
                    <button
                      onClick={() => removeTimePeriod(day.day, timeSlotIndex)}
                      className="cursor-pointer rounded-full border border-gray-500 p-1"
                    >
                      <TrashBinIcon styles="h-4 w-4" />
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
