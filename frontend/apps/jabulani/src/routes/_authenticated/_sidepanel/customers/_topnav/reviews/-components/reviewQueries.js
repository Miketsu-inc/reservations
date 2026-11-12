import { invalidateLocalStorageAuth } from "@reservations/lib";
import {
  infiniteQueryOptions,
  keepPreviousData,
  queryOptions,
} from "@tanstack/react-query";

export const PAGE_SIZE = 10;

export async function fetchMerchantReviews(
  merchantId,
  {
    stars,
    employeeIds,
    serviceIds,
    hasReply,
    sort = "newest",
    cursor,
    limit = PAGE_SIZE,
  } = {}
) {
  const params = new URLSearchParams();
  if (stars && stars !== "all") params.set("stars", String(stars));

  if (employeeIds && employeeIds.length > 0) {
    params.set("employee_ids", employeeIds.join(","));
  }

  if (serviceIds && serviceIds.length > 0) {
    params.set("service_ids", serviceIds.join(","));
  }

  if (hasReply !== undefined && hasReply !== null && hasReply !== "all") {
    params.set(
      "has_reply",
      hasReply === "replied" || hasReply === true ? "true" : "false"
    );
  }
  if (sort) params.set("sort", sort);
  if (limit) params.set("limit", String(limit));
  if (cursor) params.set("cursor", cursor);

  const response = await fetch(
    `/api/v1/merchants/${merchantId}/reviews?${params.toString()}`,
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
  }

  return result.data;
}

export function merchantReviewsInfiniteQueryOptions(merchantId, filters = {}) {
  const employeeIdsKey = filters.employeeIds?.slice().sort().join(",") || "all";
  const serviceIdsKey = filters.serviceIds?.slice().sort().join(",") || "all";

  return infiniteQueryOptions({
    queryKey: [
      merchantId,
      "merchant-reviews",
      filters.stars ?? "all",
      employeeIdsKey,
      serviceIdsKey,
      filters.hasReply ?? "all",
      filters.sort ?? "newest",
    ],
    queryFn: ({ pageParam }) =>
      fetchMerchantReviews(merchantId, {
        ...filters,
        cursor: pageParam,
      }),
    initialPageParam: "",
    getNextPageParam: (lastPage) =>
      lastPage.has_next_page ? lastPage.next_cursor : undefined,
    placeholderData: keepPreviousData,
  });
}

export async function fetchMerchantReviewStats(merchantId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/reviews/stats`,
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
  }

  return result.data;
}

export function merchantReviewStatsQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "merchant-reviews-stats"],
    queryFn: () => fetchMerchantReviewStats(merchantId),
  });
}

export async function replyToReview(merchantId, reviewId, comment) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/reviews/${reviewId}/reply`,
    {
      method: "POST",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify({ comment }),
    }
  );

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    const result = await response.json();
    throw result.error;
  }
}

export async function deleteReviewReply(merchantId, reviewId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/reviews/${reviewId}/reply`,
    {
      method: "DELETE",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    }
  );

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    const result = await response.json();
    throw result.error;
  }
}

export async function reportReview(merchantId, reviewId, reason) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/reviews/${reviewId}/report`,
    {
      method: "POST",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify({ reason }),
    }
  );

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    const result = await response.json();
    throw result.error;
  }
}
