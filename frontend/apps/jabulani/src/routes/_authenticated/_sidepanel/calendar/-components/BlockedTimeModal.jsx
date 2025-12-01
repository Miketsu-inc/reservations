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
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";

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

async function fetchEmployees() {
  const response = await fetch(`/api/v1/merchants/calendar/employees`, {
    method: "GET",
    headers: {
      Accept: "application/json",
      "constent-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    return result.data;
  }
}

export default function BlockedTimeModal({
  isOpen,
  onClose,
  blockedTime,
  preferences,
  onDeleted,
  onSubmitted,
}) {
  const { data: employees = [] } = useQuery({
    queryKey: ["calendar-employees"],
    queryFn: fetchEmployees,
    enabled: isOpen,
  });

  const isEditing = blockedTime !== null;
  const timeOptions = generateTimeOptions(preferences?.time_format);
  const [isDatepickerOpen, setIsDatepickerOpen] = useState(false);
  const [isSelectOpen, setIsSelectOpen] = useState(false);
  const { showToast } = useToast();
  const [formData, setFormData] = useState({
    id: blockedTime?.extendedProps?.id || null,
    name: blockedTime?.title || "",
    employee_id: blockedTime?.extendedProps.employee_id || "",
    from_date: blockedTime?.start || new Date(),
    to_date: blockedTime?.end || new Date(),
    from_time:
      !blockedTime?.extendedProps?.allDay && blockedTime?.start
        ? timeStringFromDate(blockedTime?.start).split(" ")[0]
        : "09:00",
    to_time:
      !blockedTime?.extendedProps?.allDay && blockedTime?.end
        ? timeStringFromDate(blockedTime?.end).split(" ")[0]
        : "17:00",
    all_day: blockedTime?.extendedProps?.allDay ?? true,
  });

  const employeeOptions = employees?.map((employee) => ({
    value: employee.id,
    label: employee.first_name + " " + employee.last_name,
  }));

  async function handleSubmit(e) {
    e.preventDefault();

    if (!e.target.checkValidity()) {
      return;
    }

    const body = {
      id: blockedTime?.extendedProps?.id ?? undefined,
      name: formData.name,
      all_day: formData.all_day,
    };

    if (formData.all_day) {
      body.from_date = startOfDay(formData.from_date).toISOString();

      const fromDateOnly = startOfDay(formData.from_date).getTime();
      const toDateOnly = startOfDay(formData.to_date).getTime();

      if (fromDateOnly === toDateOnly) {
        body.to_date = endOfDay(formData.to_date).toISOString();
      } else {
        body.to_date = startOfDay(formData.to_date).toISOString();
      }
    } else {
      body.from_date = combineDateTimeLocal(
        formData.from_date,
        formData.from_time
      ).toISOString();

      body.to_date = combineDateTimeLocal(
        formData.to_date,
        formData.to_time
      ).toISOString();
    }

    let url = "";
    let method = "";

    // when three is more than one employee the ids at insert are sent as an array but the updating is not (not implemented yet)
    // means that the blocek time was already added and now should be modified
    if (formData.id != null) {
      body.employee_id = blockedTime?.extendedProps?.employee_id;
      url = `/api/v1/merchants/blocked-times/${formData.id}`;
      method = "PUT";
    } else {
      // for correct json sending
      delete body.id;
      body.employee_ids = [formData.employee_id];
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
        body: JSON.stringify({
          employee_id: bt.employee_id,
        }),
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

  return (
    <Modal
      styles="w-full sm:w-fit"
      isOpen={isOpen}
      onClose={onClose}
      disableFocusTrap={true}
      suspendCloseOnClickOutside={isDatepickerOpen || isSelectOpen}
    >
      <form id="BlockedTimeForm" onSubmit={handleSubmit} className="">
        <div className="flex flex-col gap-3 p-3">
          <div className="flex items-center justify-between">
            <p className="text-lg md:text-xl">
              {isEditing ? "Edit Blocked Time" : "Add Blocked Time"}
            </p>
            <CloseButton onClick={onClose} />
          </div>
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
          <div className="text-text_color flex w-full items-center gap-4">
            <div className="flex w-full flex-col gap-1">
              <label className="text-sm">From Date</label>
              <DatePicker
                styles="sm:w-40 w-36"
                value={formData.from_date}
                disabledBefore={new Date()}
                onSelect={(date) => {
                  updateBlockedTimeData({ from_date: date });
                  if (date > formData.to_date) {
                    updateBlockedTimeData({ to_date: date });
                  }
                }}
                onOpenChange={(open) => setIsDatepickerOpen(open)}
              />
            </div>
            <div className="flex w-full flex-col gap-1">
              <label className="text-sm">To Date</label>
              <DatePicker
                styles="sm:w-40 w-36"
                value={formData.to_date}
                disabledBefore={formData.from_date}
                onSelect={(date) => {
                  updateBlockedTimeData({ to_date: date });
                }}
                onOpenChange={(open) => setIsDatepickerOpen(open)}
              />
            </div>
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
                <label className="text-sm">From Time</label>
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
                <label className="text-sm">To Time</label>
                <Select
                  options={timeOptions}
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
          <div className="flex w-full flex-col gap-1">
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
          </div>
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
                onClick={onClose}
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
