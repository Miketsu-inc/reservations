export async function isAuthenticated(path) {
  try {
    const response = await fetch(path, {
      method: "GET",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    });

    if (response.ok) {
      return true;
    } else {
      return false;
    }
  } catch (err) {
    // Should be return false in production with logging
    throw new Error("Something went wrong while checking user auth: " + err);
  }
}
