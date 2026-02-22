import { queryOptions } from "@tanstack/react-query";
import { invalidateLocalStorageAuth } from "./lib";

async function fetchPreferences() {
  const response = await fetch(`/api/v1/merchant/preferences`, {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

export function preferencesQueryOptions() {
  return queryOptions({
    queryKey: ["preferences"],
    queryFn: fetchPreferences,
  });
}

async function fetchBusinessHours() {
  const response = await fetch(
    `/api/v1/merchant/settings/business-hours/normalized`,
    {
      method: "GET",
      headers: {
        Accept: "application/json",
        "constent-type": "application/json",
      },
    }
  );

  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    return result.data;
  }
}

export function businessHoursQueryOptions() {
  return queryOptions({
    queryKey: ["normalized-business-hours"],
    queryFn: fetchBusinessHours,
  });
}

async function fetchCustomers() {
  const response = await fetch(`/api/v1/merchant/customers`, {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

export function customersQueryOptions() {
  return queryOptions({
    queryKey: ["customers"],
    queryFn: fetchCustomers,
  });
}

async function fetchBlockedTimeTypes() {
  const response = await fetch(`/api/v1/merchant/blocked-time-types`, {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

export function blockedTimeTypesQueryOptions() {
  return queryOptions({
    queryKey: ["blocked-time-types"],
    queryFn: fetchBlockedTimeTypes,
  });
}

async function fetchServiceFormOptions() {
  const response = await fetch("/api/v1/merchant/services/form-options", {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

export function serviceFormOptionsQueryOptions() {
  return queryOptions({
    queryKey: ["service-from-options"],
    queryFn: fetchServiceFormOptions,
  });
}
