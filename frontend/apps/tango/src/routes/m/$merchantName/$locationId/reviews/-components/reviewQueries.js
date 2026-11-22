import { reviewFiltersToSearchParams } from "@reservations/lib";
import {
  infiniteQueryOptions,
  keepPreviousData,
  queryOptions,
} from "@tanstack/react-query";

async function fetchPublicReviews(merchantName, locationId, filters) {
  const params = reviewFiltersToSearchParams(filters);

  const response = await fetch(
    `/api/v1/public/merchants/${merchantName}/locations/${locationId}/reviews?${params}`,
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
    throw result.error;
  } else {
    return result.data;
  }
}

export function publicReviewsInfiniteQueryOptions(
  merchantName,
  locationId,
  filters
) {
  return infiniteQueryOptions({
    queryKey: ["public-reviews-infinite", merchantName, locationId, filters],
    queryFn: ({ pageParam }) =>
      fetchPublicReviews(merchantName, locationId, {
        ...filters,
        cursor: pageParam,
      }),
    initialPageParam: "",
    getNextPageParam: (lastPage) =>
      lastPage.has_next_page ? lastPage.next_cursor : undefined,
    placeholderData: keepPreviousData,
  });
}

async function fetchPublicReviewStats(merchantName, locationId) {
  const response = await fetch(
    `/api/v1/public/merchants/${merchantName}/locations/${locationId}/reviews/stats`,
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
    throw result.error;
  } else {
    return result.data;
  }
}

export function publicReviewStatsQueryOptions(merchantName, locationId) {
  return queryOptions({
    queryKey: ["public-review-stats", merchantName, locationId],
    queryFn: () => fetchPublicReviewStats(merchantName, locationId),
  });
}
