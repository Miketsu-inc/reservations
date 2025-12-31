import { EditIcon } from "@reservations/assets";
import {
  Button,
  CloseButton,
  DatePicker,
  Input,
  Modal,
  Select,
  Switch,
} from "@reservations/components";
import {
  combineDateTimeLocal,
  invalidateLocalStorageAuth,
  timeStringFromDate,
  useToast,
} from "@reservations/lib";
// import { queryOptions } from "@tanstack/react-query";
import {
  blockedTimeTypesQueryOptions,
  formatDuration,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";

const generateTimeOptions = (time_format) => {
  const options = [];

  for (let hour = 0; hour < 24; hour++) {
    for (let minute of [0, 30]) {
      const value = `${hour.toString().padStart(2, "0")}:${minute === 0 ? "00" : "30"}`;

      let label;
      if (time_format === "12-hour") {
        const period = hour >= 12 ? "PM" : "AM";
        const hour12 = hour % 12 || 12;
        label = `${hour12}:${minute === 0 ? "00" : "30"} ${period}`;
      } else {
        label = `${hour}:${minute === 0 ? "00" : "30"}`;
      }

      options.push({ label, value });
    }
  }
  return options;
};

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

// async function fetchEmployees() {
//   const response = await fetch(`/api/v1/merchants/calendar/employees`, {
//     method: "GET",
//     headers: {
//       Accept: "application/json",
//       "constent-type": "application/json",
//     },
//   });

//   const result = await response.json();
//   if (!response.ok) {
//     throw result.error;
//   } else {
//     return result.data;
//   }
// }

// function employeeQueryOptions() {
//   return queryOptions({
//     queryKey: ["calendar-employees"],
//     queryFn: fetchEmployees,
//   });
// }

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

export default function BlockedTimeModal({
  isOpen,
  onClose,
  blockedTime,
  preferences,
  onDeleted,
  onSubmitted,
}) {
  // const { data: employees = [] } = useQuery(employeeQueryOptions());

  const isEditing = blockedTime !== null;
  const timeOptions = generateTimeOptions(preferences?.time_format);
  const [isDatepickerOpen, setIsDatepickerOpen] = useState(false);
  const [isSelectOpen, setIsSelectOpen] = useState(false);
  const { showToast } = useToast();
  const [formData, setFormData] = useState({
    id: blockedTime?.extendedProps?.id || null,
    blocked_type_id: blockedTime?.extendedProps?.blocked_type_id || null,
    name: blockedTime?.extendedProps?.name || "",
    employee_id: blockedTime?.extendedProps.employee_id || "",
    date: blockedTime?.start || new Date(),
    from_time:
      !blockedTime?.extendedProps?.allDay && blockedTime?.start
        ? timeStringFromDate(blockedTime?.start).split(" ")[0]
        : "09:00",
    to_time:
      !blockedTime?.extendedProps?.allDay && blockedTime?.end
        ? timeStringFromDate(blockedTime?.end).split(" ")[0]
        : "17:00",
    all_day: blockedTime?.extendedProps?.allDay ?? false,
  });
  const initialActiveType =
    isEditing && formData?.blocked_type_id === null
      ? "custom"
      : formData?.blocked_type_id;
  const [activeType, setActiveType] = useState(initialActiveType);

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

    const body = {
      id: blockedTime?.extendedProps?.id ?? undefined,
      blocked_type_id: formData.blocked_type_id,
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
      const endTime = `${endHours.toString().padStart(2, "0")}:${endMins.toString().padStart(2, "0")}`;
      // because the select has 30 minute gaps if a blocked type has lets say 40m than the to select will be empty
      updateBlockedTimeData({ to_time: endTime });
    }
  }

  return (
    <Modal
      styles="w-full sm:w-fit"
      isOpen={isOpen}
      onClose={() => {
        onClose();
        setActiveType(null);
        setFormData(defaultFormData);
      }}
      disableFocusTrap={true}
      suspendCloseOnClickOutside={isDatepickerOpen || isSelectOpen}
    >
      <form id="BlockedTimeForm" onSubmit={handleSubmit} className="">
        <div className="flex flex-col gap-3 p-3">
          <div className="flex items-center justify-between">
            <p className="text-lg md:text-xl">
              {isEditing ? "Edit Blocked Time" : "Add Blocked Time"}
            </p>
            <CloseButton
              onClick={() => {
                onClose();
                setActiveType(null);
                setFormData(defaultFormData);
              }}
            />
          </div>
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
              styles="sm:w-80 w-36"
              value={formData.date}
              disabledBefore={new Date()}
              onSelect={(date) => {
                updateBlockedTimeData({ date: date });
              }}
              onOpenChange={(open) => setIsDatepickerOpen(open)}
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
            <div className="text-text_color flex w-full items-center gap-4">
              <div className="flex w-full flex-col gap-1">
                <label className="text-sm">From</label>
                <Select
                  options={timeOptions.filter(
                    (option) => option.value !== "23:30:00"
                  )}
                  value={formData.from_time}
                  onSelect={(option) =>
                    updateBlockedTimeData({ from_time: option.value })
                  }
                  styles="flex-1"
                  maxVisibleItems={7}
                  onOpenChange={(open) => setIsSelectOpen(open)}
                />
              </div>
              <div className="flex w-full flex-col gap-1">
                <label className="text-sm">To</label>
                <Select
                  options={timeOptions.filter(
                    (option) => option.value > formData.from_time
                  )}
                  value={formData.to_time}
                  onSelect={(option) =>
                    updateBlockedTimeData({ to_time: option.value })
                  }
                  styles="flex-1"
                  maxVisibleItems={7}
                  onOpenChange={(open) => setIsSelectOpen(open)}
                />
              </div>
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
          <div className="flex items-center justify-end gap-2 pt-2">
            {isEditing && (
              <Button
                styles="p-2"
                buttonText="Delete"
                variant="danger"
                type="button"
                onClick={() => handleDelete(formData)}
              />
            )}
            <div className="flex flex-1 justify-end gap-3">
              <Button
                styles="p-2"
                buttonText="Cancel"
                variant="tertiary"
                type="button"
                onClick={() => {
                  onClose();
                  setActiveType(null);
                  setFormData(defaultFormData);
                }}
              />
              <Button
                type="submit"
                variant="primary"
                styles="p-2"
                buttonText={isEditing ? "Update" : "Create"}
              />
            </div>
          </div>
        </div>
      </form>
    </Modal>
  );
}

function BlockedTypeSection({ onSelect, blockedTypes, activeType }) {
  const scrollRef = useRef(null);

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
    <div className="text-text_color flex flex-col gap-3 sm:w-80">
      <label className="text-sm font-medium">Block time type</label>

      <div
        ref={scrollRef}
        className="flex gap-3 overflow-x-auto pb-2 dark:scheme-dark"
      >
        <button
          type="button"
          onClick={() => onSelect("custom")}
          className={`flex h-26 w-28 shrink-0 flex-col items-center
            justify-center gap-2 rounded-md border-2 transition-all ${
              activeType === "custom"
                ? "border-primary"
                : "border-border_color hover:border-gray-400"
            }`}
        >
          <EditIcon styles="size-6 mt-3" />
          <div className="text-center">
            <div className="text-sm font-medium">Custom</div>
            <div className="text-xs text-gray-500 dark:text-gray-400">
              New blocked time
            </div>
          </div>
        </button>
        {blockedTypes.map((type) => (
          <button
            key={type.id}
            type="button"
            onClick={() => onSelect(type)}
            className={`flex h-26 w-28 shrink-0 flex-col items-center
            justify-center gap-2 rounded-lg border-2 transition-all ${
              activeType === type.id
                ? "border-primary"
                : " border-border_color hover:border-gray-400"
            }`}
          >
            <span className="text-3xl">{type.icon}</span>
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
