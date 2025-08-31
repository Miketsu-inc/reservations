import Button from "@components/Button";
import Loading from "@components/Loading";
import Select from "@components/Select";
import ServerError from "@components/ServerError";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { preferencesQueryOptions } from "@lib/queries";
import { useMutation, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import RadioInputGroup from "../-components/RadioInputGroup";
import SectionHeader from "../-components/SectionHeader";

const calendarViewOptions = [
  { value: "month", label: "Month View" },
  { value: "week", label: "Week View" },
  { value: "day", label: "Day View" },
  { value: "list", label: "List View" },
];

const TimeFrequencyOptions = [
  { value: "00:10:00", label: "10 minute" },
  { value: "00:15:00", label: "15 minute" },
  { value: "00:30:00", label: "30 minute" },
];

function convertTimeToMinutes(time) {
  const [hours, minutes] = time.split(":").map(Number);
  return hours * 60 + minutes;
}

async function updatePreferences(preferences) {
  const response = await fetch("/api/v1/merchants/preferences", {
    method: "PATCH",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
    body: JSON.stringify(preferences),
  });

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    const result = await response.json();
    throw result.error;
  }
}

function validateTimeRange(startHour, endHour) {
  if (!startHour || !endHour) return true;

  const startTime = convertTimeToMinutes(startHour);
  const endTime = convertTimeToMinutes(endHour);

  return startTime < endTime;
}

const defaultPreferences = {
  first_day_of_week: "",
  time_format: "",
  calendar_view: "",
  calendar_view_mobile: "",
  start_hour: "",
  end_hour: "",
  time_frequency: "",
};

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_pages/calendar"
)({
  component: CalendarPage,
  loader: ({ context: { queryClient } }) => {
    return queryClient.ensureQueryData(preferencesQueryOptions());
  },
});

function CalendarPage() {
  const [preferences, setPreferences] = useState(defaultPreferences);
  const [serverError, setServerError] = useState("");
  const [errorMessage, setErrorMessage] = useState("");

  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { data, isLoading, isError, error } = useQuery(
    preferencesQueryOptions()
  );

  const updateMutation = useMutation({
    mutationFn: updatePreferences,
    onSuccess: () => {
      queryClient.setQueryData(["preferences"], preferences);
      setServerError("");
    },
    onError: (error) => {
      setServerError(error.message);
    },
  });

  useEffect(() => {
    if (data) {
      setPreferences(data);
    }
  }, [data]);

  function handleUpdate() {
    if (JSON.stringify(preferences) === JSON.stringify(data)) {
      return;
    }

    updateMutation.mutate(preferences);
  }

  function handleInputChange(key, value) {
    setPreferences((prev) => {
      const newPreferences = { ...prev, [key]: value };

      if (key === "start_hour" || key === "end_hour") {
        const startHour = key === "start_hour" ? value : prev.start_hour;
        const endHour = key === "end_hour" ? value : prev.end_hour;

        if (!validateTimeRange(startHour, endHour)) {
          const errorMsg =
            key === "start_hour"
              ? "Start time must be before end time"
              : "End time must be after start time";
          setErrorMessage(errorMsg);
          // Don't update if validation fails
          return prev;
        }
      }

      setErrorMessage("");
      return newPreferences;
    });
  }

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  return (
    <div className="flex w-full flex-col gap-8">
      <div className="flex flex-col gap-4">
        <SectionHeader title="Appearance" styles="" />
        <RadioInputGroup
          title="First day of the week"
          name="firstDayOfWeek"
          value={preferences.first_day_of_week}
          onChange={(value) => handleInputChange("first_day_of_week", value)}
          options={[
            { value: "Monday", label: "Monday" },
            { value: "Sunday", label: "Sunday" },
          ]}
          description="Choose which day your calendar week starts on. This setting will affect how dates are displayed in your scheduling system."
        />
        <RadioInputGroup
          title="Time Format"
          name="timeFormat"
          value={preferences.time_format}
          onChange={(value) => handleInputChange("time_format", value)}
          options={[
            { value: "24-hour", label: "24-Hour Format" },
            { value: "12-hour", label: "12-Hour Format" },
          ]}
          description="Select how time is displayed in your calendar. The 24-hour format is common in Europe, while the 12-hour AM/PM format is standard in the U.S."
        />
      </div>
      <div className="flex flex-col gap-6">
        <label
          htmlFor="desktop-view"
          className="flex flex-col gap-2 font-semibold"
        >
          Desktop Default View
          <Select
            options={calendarViewOptions}
            value={preferences.calendar_view}
            onSelect={(option) =>
              handleInputChange("calendar_view", option.value)
            }
            placeholder=""
            styles="font-normal md:w-2/3"
          />
        </label>
        <label
          htmlFor="mobile-view"
          className="flex flex-col gap-2 font-semibold"
        >
          Mobile Default View
          <Select
            options={calendarViewOptions}
            value={preferences.calendar_view_mobile}
            onSelect={(option) =>
              handleInputChange("calendar_view_mobile", option.value)
            }
            placeholder=""
            styles="font-normal md:w-2/3"
          />
        </label>
      </div>

      <div className="flex justify-between gap-10 md:flex-row">
        <label
          htmlFor="start-hour"
          className="flex w-full flex-col gap-2 font-semibold"
        >
          Starting hour
          <input
            type="time"
            id="start-hour"
            value={preferences.start_hour}
            onChange={(e) => handleInputChange("start_hour", e.target.value)}
            className="bg-hvr_gray rounded-lg border p-2 font-normal
              dark:scheme-dark"
            step="1800"
          />
        </label>
        <label
          htmlFor="end-hour"
          className="flex w-full flex-col gap-2 font-semibold"
        >
          Ending hour
          <input
            type="time"
            id="end-hour"
            value={preferences.end_hour}
            onChange={(e) => handleInputChange("end_hour", e.target.value)}
            className="bg-hvr_gray rounded-lg border p-2 font-normal
              dark:scheme-dark"
            step="1800"
          />
        </label>
      </div>
      {errorMessage && (
        <div className="text-sm text-red-500">{errorMessage}</div>
      )}
      <label htmlFor="time-slot" className="flex flex-col gap-2 font-semibold">
        Time slot frequency
        <Select
          options={TimeFrequencyOptions}
          value={preferences.time_frequency}
          onSelect={(option) =>
            handleInputChange("time_frequency", option.value)
          }
          placeholder=""
          styles="font-normal md:w-2/3"
        />
      </label>
      <div className="flex flex-col gap-3">
        <span className="text-text_color/70 text-sm md:w-2/3">
          Update your calendar preferences below. All fields are optional, and
          your settings will apply to your scheduling system immediately.
        </span>
        <Button
          styles="w-min px-2 text-nowrap py-1"
          variant="primary"
          buttonText="Update fields"
          type="button"
          onClick={handleUpdate}
        />
        <ServerError error={serverError} styles="mt-2" />
      </div>
    </div>
  );
}
