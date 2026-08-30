import { queryOptions } from "@tanstack/react-query";
import { invalidateLocalStorageAuth } from "./lib";

async function fetchPreferences(merchantId, employeeId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/team/${employeeId}/preferences`,
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

export function preferencesQueryOptions(merchantId, employeeId) {
  return queryOptions({
    queryKey: [merchantId, employeeId, "preferences"],
    queryFn: () => fetchPreferences(merchantId, employeeId),
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

async function fetchMe() {
  const response = await fetch("/api/v1/auth/me", {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();

  if (!response.ok) {
    throw { status: response.status, message: result.error };
  } else {
    return result.data;
  }
}

export function meQueryOptions() {
  return queryOptions({
    queryKey: ["me"],
    queryFn: fetchMe,
  });
}

async function fetchActiveTeam(name) {
  const response = await fetch(`/api/v1/public/merchants/${name}/team`, {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });
  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    return result.data;
  }
}

export function activeTeamQueryOptions(name) {
  return queryOptions({
    queryKey: ["active-team", name],
    queryFn: () => fetchActiveTeam(name),
  });
}

async function fetchMerchantServices(name) {
  const response = await fetch(`/api/v1/public/merchants/${name}/services`, {
    method: "GET",
  });
  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    return result.data;
  }
}

export function merchantServicesQueryOptions(name) {
  return queryOptions({
    queryKey: ["merchant-services", name],
    queryFn: () => fetchMerchantServices(name),
  });
}

async function fetchEmployee(merchantId, id) {
  const response = await fetch(`/api/v1/merchants/${merchantId}/team/${id}`, {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    return result.data;
  }
}

export function employeeQueryOptions(merchantId, id) {
  return queryOptions({
    queryKey: [merchantId, "employee", id],
    queryFn: () => fetchEmployee(merchantId, id),
  });
}
