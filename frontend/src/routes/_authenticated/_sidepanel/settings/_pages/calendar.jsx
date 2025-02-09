import Button from "@components/Button";
import ServerError from "@components/ServerError";
import { useToast } from "@lib/hooks";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import RadioInputGroup from "../-components/RadioInputGroup";
import SectionHeader from "../-components/SectionHeader";

async function fetchPreferences() {
  const response = await fetch(`/api/v1/merchants/preferences`, {
    method: "GET",
  });

  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    return result.data;
  }
}

const defaultPreferences = {
  first_day_of_week: "",
  time_format: "",
  calendar_view: "",
  calendar_view_mobile: "",
};

const calendarViewOptions = [
  { value: "month", label: "Month View" },
  { value: "week", label: "Week View" },
  { value: "day", label: "Day View" },
  { value: "list", label: "List View" },
];

function hasChanges(changedData, originalData) {
  return JSON.stringify(changedData) !== JSON.stringify(originalData);
}

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_pages/calendar"
)({
  component: CalendarPage,
  loader: () => fetchPreferences(),
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function CalendarPage() {
  const [preferences, setPreferences] = useState(defaultPreferences);
  const [originalPref, setOriginalPref] = useState(defaultPreferences);
  const loaderData = Route.useLoaderData();
  const [serverError, setServerError] = useState("");
  const { showToast } = useToast();
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);

  useEffect(() => {
    if (loaderData) {
      setPreferences(loaderData);
      setOriginalPref(loaderData);
    }
  }, [loaderData]);

  useEffect(() => {
    setHasUnsavedChanges(hasChanges(preferences, originalPref));
  }, [preferences, originalPref]);

  function handleSelect(e, type) {
    const value = e.target.value;

    if (type === "desktop") {
      setPreferences((prev) => ({
        ...prev,
        calendar_view: value,
      }));
    } else if (type === "mobile") {
      setPreferences((prev) => ({
        ...prev,
        calendar_view_mobile: value,
      }));
    }
  }

  async function updateHandler() {
    //if

    try {
      const response = await fetch("/api/v1/merchants/preferences", {
        method: "PATCH",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify(preferences),
      });

      if (!response.ok) {
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        setOriginalPref(preferences);
        setServerError("");
        showToast({
          message: "Calendar preferences updated successfully!",
          variant: "success",
        });
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  function handleRadioChange(key, value) {
    setPreferences((prev) => ({
      ...prev,
      [key]: value,
    }));
  }

  return (
    <div className="flex w-full flex-col gap-8">
      <div className="flex flex-col gap-4">
        <SectionHeader title="Appearance" styles="" />
        <RadioInputGroup
          title="First day of the week"
          name="firstDayOfWeek"
          value={preferences.first_day_of_week}
          onChange={(value) => handleRadioChange("first_day_of_week", value)}
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
          onChange={(value) => handleRadioChange("time_format", value)}
          options={[
            { value: "24-hour", label: "24-Hour Format" },
            { value: "12-hour", label: "12-Hour Format" },
          ]}
          description="Select how time is displayed in your calendar. The 24-hour format is common in Europe, while the 12-hour AM/PM format is standard in the U.S."
        />
      </div>
      <div className="flex flex-col gap-4">
        <label
          htmlFor="desktop-view"
          className="mb-2 flex flex-col gap-2 font-semibold"
        >
          Desktop Default View
          <select
            id="desktop-view"
            value={preferences.calendar_view}
            onChange={(e) => handleSelect(e, "desktop")}
            className="bg-hvr_gray rounded-lg border p-2 md:w-2/3 dark:[color-scheme:dark]"
          >
            {calendarViewOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
        <label
          htmlFor="mobile-view"
          className="mb-2 flex flex-col gap-2 font-semibold"
        >
          Mobile Default View
          <select
            id="mobile-view"
            value={preferences.calendar_view_mobile}
            onChange={(e) => handleSelect(e, "mobile")}
            className="bg-hvr_gray rounded-lg border p-2 md:w-2/3 dark:[color-scheme:dark]"
          >
            {calendarViewOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </div>
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
          onClick={updateHandler}
          disabled={!hasUnsavedChanges}
        />
        <ServerError error={serverError} styles="mt-2" />
      </div>
    </div>
  );
}
