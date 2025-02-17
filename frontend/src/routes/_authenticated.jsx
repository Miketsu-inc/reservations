import Loading from "@components/Loading";
import ServerError from "@components/ServerError";
import { isAuthenticated } from "@lib/lib";
import { createFileRoute, redirect } from "@tanstack/react-router";

const defaultPreferences = {
  first_day_of_week: "",
  time_format: "",
  calendar_view: "",
  calendar_view_mobile: "",
  start_hour: "",
  end_hour: "",
  time_frequency: "",
};

function getCachedPreferences() {
  const storedPreferences = localStorage.getItem("Preferences");
  return storedPreferences ? JSON.parse(storedPreferences) : defaultPreferences;
}

async function fetchAllPreference() {
  const response = await fetch(`/api/v1/merchants/preferences`, {
    method: "GET",
  });

  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    localStorage.setItem("Preferences", JSON.stringify(result.data));
    return result.data;
  }
}

export const Route = createFileRoute("/_authenticated")({
  beforeLoad: async () => {
    if (!(await isAuthenticated("/api/v1/auth/user"))) {
      throw redirect({
        to: "/login",
      });
    }
  },
  loader: async () => {
    const cachedPreferences = getCachedPreferences();
    if (cachedPreferences.first_day_of_week) {
      return cachedPreferences;
    }

    return await fetchAllPreference();
  },
  pendingComponent: Loading,
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});
