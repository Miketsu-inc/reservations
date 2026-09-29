import { JABULANI_URL, TANGO_URL } from "./constants";

export function invalidateLocalStorageAuth(responseCode) {
  if (responseCode === 401) {
    localStorage.setItem("loggedIn", false);
  }
}

export function getDisplayPrice(price, type) {
  if (type === "free") {
    return "Free";
  }

  if (type === "from") {
    return `from ${price}`;
  }

  return price;
}

export function getSafeRedirectUrl(value) {
  if (!value) return null;

  try {
    const target = new URL(value, window.location.origin);

    const allowedOrigins = new Set([
      new URL(JABULANI_URL).origin,
      new URL(TANGO_URL).origin,
    ]);

    return allowedOrigins.has(target.origin) ? target.href : null;
  } catch {
    return null;
  }
}
