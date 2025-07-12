export async function isAuthenticated(path) {
  if (localStorage.getItem("loggedIn") === "true") {
    return true;
  }

  try {
    const response = await fetch(path, {
      method: "GET",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    });

    if (response.ok) {
      localStorage.setItem("loggedIn", true);
      return true;
    } else {
      return false;
    }
  } catch (err) {
    // Should be return false in production with logging
    throw new Error("Something went wrong while checking user auth: " + err);
  }
}

export function invalidateLocalStorageAuth(responseCode) {
  if (responseCode === 401) {
    localStorage.setItem("loggedIn", false);
  }
}

export function getStoredPreferences() {
  const storedPreferences = localStorage.getItem("Preferences");
  return storedPreferences ? JSON.parse(storedPreferences) : {};
}

export function fillStatisticsWithDate(data, fromDateStr, toDateStr) {
  let filledStats = [];

  const fromDate = new Date(fromDateStr);
  const toDate = new Date(toDateStr);

  const convertedMap = new Map(
    data.map(({ day, value }) => [new Date(day).toDateString(), value])
  );

  for (let d = fromDate; d <= toDate; d.setDate(d.getDate() + 1)) {
    const value = convertedMap.get(d.toDateString()) || 0;

    filledStats.push({
      value,
      day: d.toLocaleDateString([], {
        month: "2-digit",
        day: "2-digit",
      }),
    });
  }

  return filledStats;
}
