import { Loading, Select, ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  invalidateLocalStorageAuth,
  preferencesQueryOptions,
  useToast,
} from "@reservations/lib";
import { useMutation, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import RadioInputGroup from "./-components/RadioInputGroup";
import SettingsPageHeader from "./-components/SettingsPageHeader";
import SettingsSaveAction from "./-components/SettingsSaveAction";
import SettingsSection from "./-components/SettingsSection";

const calendarViewOptions = [
  { value: "month", label: "Month View" },
  { value: "week", label: "Week View" },
  { value: "day", label: "Day View" },
  { value: "list", label: "List View" },
];

const TimeFrequencyOptions = [
  { value: "00:10", label: "10 minute" },
  { value: "00:15", label: "15 minute" },
  { value: "00:30", label: "30 minute" },
];

function convertTimeToMinutes(time) {
  const [hours, minutes] = time.split(":").map(Number);
  return hours * 60 + minutes;
}

async function updatePreferences(merchantId, employeeId, preferences) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/team/${employeeId}/preferences`,
    {
      method: "PATCH",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify(preferences),
    }
  );

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
  "/_authenticated/_sidepanel/settings/_topnav/calendar"
)({
  component: CalendarPage,
  loader: async ({
    context: {
      queryClient,
      authContext: { merchantId, employeeId },
    },
  }) => {
    await queryClient.ensureQueryData(
      preferencesQueryOptions(merchantId, employeeId)
    );
  },
});

function CalendarPage() {
  const [unsavedChanges, setUnsavedChanges] = useState({});
  const [errorMessage, setErrorMessage] = useState("");

  const { merchantId, employeeId } = useAuth();
  const { showToast } = useToast();
  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { data, isLoading, isError, error } = useQuery(
    preferencesQueryOptions(merchantId, employeeId)
  );

  const preferences = { ...(data || defaultPreferences), ...unsavedChanges };

  const updateMutation = useMutation({
    mutationFn: (preferences) =>
      updatePreferences(merchantId, employeeId, preferences),
    onSuccess: () => {
      queryClient.setQueryData(
        [merchantId, employeeId, "preferences"],
        preferences
      );
      setUnsavedChanges({});
      showToast({
        message: "Calendar preferences updated successfully",
        variant: "success",
      });
    },
  });

  function handleUpdate() {
    if (Object.keys(unsavedChanges).length === 0) {
      return;
    }

    updateMutation.mutate(preferences);
  }

  function handleInputChange(key, value) {
    setUnsavedChanges((prev) => {
      const newChanges = { ...prev, [key]: value };
      const newPreferences = { ...(data || defaultPreferences), ...newChanges };

      if (key === "start_hour" || key === "end_hour") {
        const startHour = newPreferences.start_hour;
        const endHour = newPreferences.end_hour;

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
      return newChanges;
    });
  }

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  return (
    <div className="flex w-full flex-col gap-12 pb-28 md:pb-8">
      <SettingsPageHeader
        title="Calendar"
        description="Personalize how the calendar is displayed for your account. These preferences do not affect other team members."
        action={
          <SettingsSaveAction
            buttonText="Save"
            onClick={handleUpdate}
            disabled={Object.keys(unsavedChanges).length === 0}
            isLoading={updateMutation.isPending}
            error={updateMutation.error}
          />
        }
      />
      <SettingsSection
        title="Date and time format"
        description="Choose how weeks and times are displayed throughout your calendar."
      >
        <div className="flex flex-col gap-6">
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
      </SettingsSection>

      <SettingsSection
        title="Default views"
        description="Set the view that opens first on each device."
      >
        <div className="grid gap-6 md:grid-cols-2">
          <Select
            options={calendarViewOptions}
            value={preferences.calendar_view}
            labelText="Desktop default view"
            required={false}
            onSelect={(option) =>
              handleInputChange("calendar_view", option.value)
            }
            placeholder=""
            styles="font-normal"
          />
          <Select
            options={calendarViewOptions}
            value={preferences.calendar_view_mobile}
            labelText="Mobile default view"
            required={false}
            onSelect={(option) =>
              handleInputChange("calendar_view_mobile", option.value)
            }
            placeholder=""
            styles="font-normal"
          />
        </div>
      </SettingsSection>

      <SettingsSection
        title="Calendar grid"
        description="Limit the visible part of the day and choose the time-grid interval."
      >
        <div className="grid items-end gap-6 md:grid-cols-3">
          <label htmlFor="start-hour" className="flex flex-col text-sm">
            <span className="pb-1">Starting hour</span>
            <input
              type="time"
              id="start-hour"
              value={preferences.start_hour}
              onChange={(event) =>
                handleInputChange("start_hour", event.target.value)
              }
              className="border-input_border_color bg-layer_bg min-h-10 rounded-lg
                border px-3 py-2 font-normal dark:scheme-dark"
              step="1800"
            />
          </label>
          <label htmlFor="end-hour" className="flex flex-col text-sm">
            <span className="pb-1">Ending hour</span>
            <input
              type="time"
              id="end-hour"
              value={preferences.end_hour}
              onChange={(event) =>
                handleInputChange("end_hour", event.target.value)
              }
              className="border-input_border_color bg-layer_bg min-h-10 rounded-lg
                border px-3 py-2 font-normal dark:scheme-dark"
              step="1800"
            />
          </label>
          <Select
            options={TimeFrequencyOptions}
            value={preferences.time_frequency}
            labelText="Time slot frequency"
            required={false}
            onSelect={(option) =>
              handleInputChange("time_frequency", option.value)
            }
            placeholder=""
            styles="font-normal"
          />
        </div>
        {errorMessage && (
          <p className="mt-3 text-sm text-red-600 dark:text-red-400">
            {errorMessage}
          </p>
        )}
      </SettingsSection>
    </div>
  );
}
