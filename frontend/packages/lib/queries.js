import { queryOptions } from "@tanstack/react-query";
import { invalidateLocalStorageAuth } from "./lib";

async function fetchPreferences(merchantId) {
  const response = await fetch(`/api/v1/merchants/${merchantId}/preferences`, {
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

export function preferencesQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "preferences"],
    queryFn: () => fetchPreferences(merchantId),
  });
}

async function fetchBusinessHours(merchantId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/settings/business-hours/normalized`,
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

export function businessHoursQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "normalized-business-hours"],
    queryFn: () => fetchBusinessHours(merchantId),
  });
}

async function fetchCustomers(merchantId) {
  const response = await fetch(`/api/v1/merchants/${merchantId}/customers`, {
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

export function customersQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "customers"],
    queryFn: () => fetchCustomers(merchantId),
  });
}

async function fetchBlockedTimeTypes(merchantId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/blocked-time-types`,
    {
      method: "GET",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    }
  );

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

export function blockedTimeTypesQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "blocked-time-types"],
    queryFn: () => fetchBlockedTimeTypes(merchantId),
  });
}

async function fetchServiceFormOptions(merchantId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/services/form-options`,
    {
      method: "GET",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    }
  );

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

export function serviceFormOptionsQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "service-from-options"],
    queryFn: () => fetchServiceFormOptions(merchantId),
  });
}
