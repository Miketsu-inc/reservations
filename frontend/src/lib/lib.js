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

export function invalidateLocalSotrageAuth(responseCode) {
  if (responseCode === 401) {
    localStorage.setItem("loggedIn", false);
  }
}

export function getStoredPreferences() {
  const storedPreferences = localStorage.getItem("Preferences");
  return storedPreferences ? JSON.parse(storedPreferences) : {};
}
