import { EditIcon, PlusIcon, TrashBinIcon } from "@reservations/assets";
import {
  Button,
  CloseButton,
  DatePicker,
  Input,
  Select,
  Switch,
} from "@reservations/components";
import {
  blockedTimeTypesQueryOptions,
  combineDateTimeLocal,
  formatDuration,
  GenerateTimeOptions,
  invalidateLocalStorageAuth,
  timeStringFromDate,
  useToast,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";

function getFormattedLabel(timeValue, timeFormat) {
  if (!timeValue) return "";
  const [hours, minutes] = timeValue.split(":").map(Number);

  if (timeFormat === "12-hour") {
    const period = hours >= 12 ? "PM" : "AM";
    const hour12 = hours % 12 || 12;
    return `${hour12}:${minutes.toString().padStart(2, "0")} ${period}`;
  }

  return `${hours}:${minutes.toString().padStart(2, "0")}`;
}

function startOfDay(date) {
  const d = new Date(date);
  d.setHours(0, 0, 0, 0);
  return d;
}

function endOfDay(date) {
  const d = new Date(date);
  d.setHours(0, 0, 0, 0);
  d.setDate(d.getDate() + 1);
  return d;
}

const defaultFormData = {
  id: null,
  blocked_type_id: null,
  name: "",
  employee_id: "",
  date: new Date(),
  from_time: "09:00",
  to_time: "17:00",
  all_day: false,
};

export default function BlockedTimePanel({
  onClose,
  blockedTime,
  preferences,
  onDeleted,
  onSubmitted,
  isWindowSmall,
}) {
  const isEditing = blockedTime !== null;
  const originalTimeOptions = GenerateTimeOptions(preferences?.time_format);
  const initialToTime =
    !blockedTime?.extendedProps?.allDay && blockedTime?.end
      ? timeStringFromDate(blockedTime?.end).split(" ")[0]
      : "17:00";
  const [timeOptions, setTimeOptions] = useState(() => {
    const options = originalTimeOptions;
    const timeExists = options.some((opt) => opt.value === initialToTime);

    if (!timeExists) {
      options.push({
        value: initialToTime,
        label: getFormattedLabel(initialToTime, preferences?.time_format),
      });
    }
    return options;
  });

  const { showToast } = useToast();
  const [formData, setFormData] = useState({
    id: blockedTime?.extendedProps?.id || null,
    blocked_type_id: blockedTime?.extendedProps?.blocked_type_id || "custom",
    name: blockedTime?.extendedProps?.name || "",
    employee_id: blockedTime?.extendedProps.employee_id || "",
    date: blockedTime?.start || new Date(),
    from_time:
      !blockedTime?.extendedProps?.allDay && blockedTime?.start
        ? timeStringFromDate(blockedTime?.start).split(" ")[0]
        : "09:00",
    to_time: initialToTime,
    all_day: blockedTime?.extendedProps?.allDay ?? false,
  });

  const [activeType, setActiveType] = useState(formData.blocked_type_id);

  const { data: blockedTypes = [] } = useQuery(blockedTimeTypesQueryOptions());

  // const employeeOptions = employees?.map((employee) => ({
  //   value: employee.id,
  //   label: employee.first_name + " " + employee.last_name,
  // }));

  async function handleSubmit(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    if (!activeType) {
      showToast({
        message: "Please select a blocked time type",
        variant: "error",
      });
      return;
    }

    let blockedTypeId =
      formData.blocked_type_id === "custom"
        ? undefined
        : formData.blocked_type_id;

    const body = {
      id: blockedTime?.extendedProps?.id ?? undefined,
      blocked_type_id: blockedTypeId,
      name: formData.name,
      all_day: formData.all_day,
    };

    if (formData.all_day) {
      body.from_date = startOfDay(formData.date).toISOString();
      body.to_date = endOfDay(formData.date).toISOString();
    } else {
      body.from_date = combineDateTimeLocal(
        formData.date,
        formData.from_time
      ).toISOString();

      body.to_date = combineDateTimeLocal(
        formData.date,
        formData.to_time
      ).toISOString();
    }

    let url = "";
    let method = "";

    // when three is more than one employee the ids at insert are sent as an array but the updating is not (not implemented yet)
    // means that the blocek time was already added and now should be modified
    if (formData.id != null) {
      // body.employee_id = blockedTime?.extendedProps?.employee_id;
      url = `/api/v1/merchants/blocked-times/${formData.id}`;
      method = "PUT";
    } else {
      // for correct json sending
      delete body.id;
      // body.employee_ids = [formData.employee_id];
      url = "/api/v1/merchants/blocked-times";
      method = "POST";
    }

    try {
      const response = await fetch(url, {
        method: method,
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        showToast({ message: result.error.message, variant: "error" });
      } else {
        showToast({
          message:
            method === "POST"
              ? "Blocked Time added successfully"
              : "Blocked Time modified successfully",
          variant: "success",
        });
        setActiveType(null);
        setFormData(defaultFormData);
        onSubmitted();
        onClose();
      }
    } catch (err) {
      showToast({ message: err.message, variant: "error" });
    }
  }

  async function handleDelete(bt) {
    try {
      const response = await fetch(`/api/v1/merchants/blocked-times/${bt.id}`, {
        method: "DELETE",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        // body: JSON.stringify({
        //   employee_id: bt.employee_id,
        // }),
      });

      if (!response.ok) {
        const result = await response.json();
        invalidateLocalStorageAuth(response.status);
        showToast({ message: result.error.message, variant: "error" });
      } else {
        showToast({
          message: "Blocked Time deleted successfully",
          variant: "success",
        });
        setActiveType(null);
        setFormData(defaultFormData);
        onDeleted();
        onClose();
      }
    } catch (err) {
      showToast({ message: err.message, variant: "error" });
    }
  }

  function updateBlockedTimeData(data) {
    setFormData((prev) => {
      const newData = { ...prev, ...data };
      return newData;
    });
  }

  function handleTypeSelect(type) {
    if (type === "custom") {
      updateBlockedTimeData({ blocked_type_id: null });
      updateBlockedTimeData({ name: "" });
      setActiveType("custom");
    } else {
      updateBlockedTimeData({ blocked_type_id: type.id });
      updateBlockedTimeData({ name: type.name });
      setActiveType(type.id);

      const [hours, minutes] = formData.from_time.split(":").map(Number);
      const durationMinutes = type.duration;

      const totalMinutes = hours * 60 + minutes + durationMinutes;
      const endHours = Math.floor(totalMinutes / 60) % 24;
      const endMins = totalMinutes % 60;
      const endTimeValue = `${endHours.toString().padStart(2, "0")}:${endMins
        .toString()
        .padStart(2, "0")}`;

      const timeExists = timeOptions.some((opt) => opt.value === endTimeValue);

      if (!timeExists) {
        const label = getFormattedLabel(endTimeValue, preferences?.time_format);
        setTimeOptions((prev) => [...prev, { label, value: endTimeValue }]);
      }

      updateBlockedTimeData({ to_time: endTimeValue });
    }
  }
  return (
    <form
      className={`no-scrollbar relative h-full w-full justify-between
        overflow-y-auto ${isWindowSmall ? "pt-0" : "pt-10"} md:w-110`}
      d="BlockedTimeForm"
      onSubmit={handleSubmit}
    >
      {isWindowSmall && (
        <div className="flex w-full items-center justify-end px-4 pt-5">
          <CloseButton
            onClick={() => {
              onClose();
              setActiveType(null);
              setFormData(defaultFormData);
            }}
            styles="size-8"
          />
        </div>
      )}

      <div className="flex flex-col gap-5 px-6 pb-6">
        <p className="text-2xl font-semibold">
          {isEditing ? "Edit Blocked Time" : "Add Blocked Time"}
        </p>
        <BlockedTypeSection
          onSelect={handleTypeSelect}
          blockedTypes={blockedTypes}
          activeType={activeType}
        />
        {activeType === "custom" && (
          <Input
            styles="w-full p-2"
            type="text"
            id="name"
            name="name"
            labelText="Name"
            placeholder="Lunch Break"
            value={formData.name}
            inputData={(data) => {
              updateBlockedTimeData({ name: data.value });
            }}
          />
        )}

        <div className="flex w-full flex-col gap-1">
          <label className="text-sm">Date</label>
          <DatePicker
            styles="w-full"
            value={formData.date}
            disabledBefore={new Date()}
            onSelect={(date) => {
              updateBlockedTimeData({ date: date });
            }}
          />
        </div>
        <div className="my-1 flex items-center gap-4">
          <span className="text-sm">All day</span>
          <Switch
            defaultValue={formData.all_day}
            onSwitch={() =>
              updateBlockedTimeData({ all_day: !formData.all_day })
            }
          />
        </div>
        {!formData?.all_day && (
          <div className="text-text_color grid grid-cols-2 gap-4">
            <Select
              allOptions={timeOptions}
              options={originalTimeOptions.filter(
                (option) => option.value !== "23:30:00"
              )}
              value={formData.from_time}
              labelText="From"
              required={false}
              onSelect={(option) =>
                updateBlockedTimeData({ from_time: option.value })
              }
              styles="flex-1"
              maxVisibleItems={7}
            />

            <Select
              allOptions={timeOptions}
              options={originalTimeOptions.filter(
                (option) => option.value > formData.from_time
              )}
              value={formData.to_time}
              labelText="To"
              required={false}
              onSelect={(option) =>
                updateBlockedTimeData({ to_time: option.value })
              }
              styles="flex-1"
              maxVisibleItems={7}
            />
          </div>
        )}
        {/* <div className="flex w-full flex-col gap-1">
                        <label className="text-text_color text-sm">Team members</label>
                        <Select
                          options={employeeOptions}
                          value={formData.employee_id}
                          onSelect={(selected) =>
                            updateBlockedTimeData({ employee_id: selected.value })
                            }
                            styles="w-full"
                            placeholder="Select employee"
                            disabled={isEditing}
                          onOpenChange={(open) => setIsSelectOpen(open)}
                          />
                      </div> */}
      </div>

      <div
        className="border-border_color bg-layer_bg fixed bottom-0 flex w-full
          items-center gap-2 border-t px-6 py-4"
      >
        {isEditing && (
          <Button
            styles="p-2 w-fit"
            variant="tertiary"
            type="button"
            onClick={() => handleDelete(formData)}
          >
            <TrashBinIcon styles="size-6" />
          </Button>
        )}
        <div className="flex flex-1 justify-end gap-3">
          <Button
            type="submit"
            variant="primary"
            styles="p-2 w-full"
            buttonText={isEditing ? "Update" : "Create"}
          />
        </div>
      </div>
    </form>
  );
}

function BlockedTypeSection({ onSelect, blockedTypes, activeType }) {
  const scrollRef = useRef(null);
  const hasTypes = blockedTypes && blockedTypes.length > 0;

  let selectedIndex = -1;

  if (activeType === "custom") {
    selectedIndex = 0;
  } else if (activeType) {
    const typeIndex = blockedTypes.findIndex((t) => t.id === activeType);
    if (typeIndex !== -1) {
      // Add 1 because "Custom"
      selectedIndex = typeIndex + 1;
    }
  }

  useEffect(() => {
    if (selectedIndex < 0) return;

    const timeout = setTimeout(() => {
      if (scrollRef.current?.children[selectedIndex]) {
        scrollRef.current.children[selectedIndex].scrollIntoView({
          behavior: "smooth",
          inline: "center",
          block: "nearest",
        });
      }
    }, 0);
    return () => clearTimeout(timeout);
  }, [activeType, selectedIndex]);

  return (
    <div className="text-text_color flex w-full flex-col gap-3">
      <label className="text-sm">Block time type</label>

      <div
        ref={scrollRef}
        className={'"overflow-x-auto" flex gap-4 pb-2 dark:scheme-dark'}
      >
        <button
          type="button"
          onClick={() => onSelect("custom")}
          className={`flex h-28 w-36 shrink-0 flex-col items-center
            justify-center gap-2 rounded-md border-2 transition-all ${
              activeType === "custom"
                ? "border-primary"
                : "border-border_color hover:border-gray-400"
            } `}
        >
          <EditIcon styles="size-7 mt-3" />

          <div className="text-center">
            <div className="text-sm font-medium">Custom</div>
            <div className="text-xs text-gray-500 dark:text-gray-400">
              New blocked time
            </div>
          </div>
        </button>
        {!hasTypes && (
          <Link
            to="/settings/scheduling"
            className="border-border_color flex h-28 w-36 shrink-0 flex-col
              items-center justify-center gap-2 rounded-lg border-2
              transition-all hover:border-gray-400"
          >
            <div
              className="bg-primary/20 text-primary flex size-12 items-center
                justify-center rounded-full"
            >
              <PlusIcon styles="size-6" />
            </div>
            <div
              className="text-sm font-medium text-gray-500 dark:text-gray-400"
            >
              Add new types
            </div>
          </Link>
        )}
        {blockedTypes.map((type) => (
          <button
            key={type.id}
            type="button"
            onClick={() => onSelect(type)}
            className={`flex h-28 w-36 shrink-0 flex-col items-center
            justify-center gap-2 rounded-lg border-2 transition-all ${
              activeType === type.id
                ? "border-primary"
                : " border-border_color hover:border-gray-400"
            }`}
          >
            <span className="text-4xl">{type.icon}</span>
            <div className="text-center">
              <div className="text-sm font-medium">{type.name}</div>
              <div className="text-xs text-gray-500 dark:text-gray-400">
                {formatDuration(type.duration)}
              </div>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
