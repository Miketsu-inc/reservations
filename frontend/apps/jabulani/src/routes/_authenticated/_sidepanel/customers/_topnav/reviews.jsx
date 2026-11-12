import { StarIcon } from "@hugeicons/core-free-icons";
import { Button, Icon, Loading, ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  calendarServicesQueryOptions,
  calendarTeamMembersQueryOptions,
  useToast,
} from "@reservations/lib";
import {
  useInfiniteQuery,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import ReportReviewModal from "./reviews/-components/ReportReviewModal";
import ReviewCard from "./reviews/-components/ReviewCard";
import ReviewFilters from "./reviews/-components/ReviewFilters";
import {
  deleteReviewReply,
  merchantReviewsInfiniteQueryOptions,
  merchantReviewStatsQueryOptions,
  replyToReview,
  reportReview,
} from "./reviews/-components/reviewQueries";
import ReviewSummary from "./reviews/-components/ReviewSummary";

const INITIAL_FILTERS = {
  stars: "all",
  hasReply: "all",
  employeeIds: [],
  serviceIds: [],
  search: "",
  sort: "newest",
};

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/customers/_topnav/reviews"
)({
  component: MerchantReviewsPage,
  loader: ({
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => {
    void queryClient.prefetchQuery(merchantReviewStatsQueryOptions(merchantId));
    void queryClient.prefetchInfiniteQuery(
      merchantReviewsInfiniteQueryOptions(merchantId)
    );
    void queryClient.prefetchQuery(calendarServicesQueryOptions(merchantId));
    void queryClient.prefetchQuery(calendarTeamMembersQueryOptions(merchantId));
  },
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});

function MerchantReviewsPage() {
  const { merchantId } = useAuth();
  const queryClient = useQueryClient();
  const { showToast } = useToast();

  const [filters, setFilters] = useState(INITIAL_FILTERS);
  const [reportingReview, setReportingReview] = useState(null);

  const {
    data,
    isLoading: isReviewsLoading,
    isError: isReviewsError,
    error: reviewsError,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery(
    merchantReviewsInfiniteQueryOptions(merchantId, filters)
  );

  const {
    data: stats,
    isLoading: isStatsLoading,
    isError: isStatsError,
    error: statsError,
  } = useQuery(merchantReviewStatsQueryOptions(merchantId));

  const { data: serviceCategories = [] } = useQuery(
    calendarServicesQueryOptions(merchantId)
  );
  const services = useMemo(
    () => serviceCategories.flatMap((cat) => cat.services || []),
    [serviceCategories]
  );
  const { data: team = [] } = useQuery(
    calendarTeamMembersQueryOptions(merchantId)
  );

  const allReviews = useMemo(
    () => data?.pages.flatMap((page) => page.reviews) ?? [],
    [data]
  );

  const displayedReviews = useMemo(() => {
    const term = filters.search.trim().toLowerCase();
    if (!term) return allReviews;

    return allReviews.filter((review) => {
      const customerName =
        `${review.customer?.first_name} ${review.customer?.last_name}`.toLowerCase();
      return customerName.includes(term);
    });
  }, [allReviews, filters.search]);

  const hasActiveFilters =
    filters.stars !== "all" ||
    filters.hasReply !== "all" ||
    filters.employeeIds.length > 0 ||
    filters.serviceIds.length > 0 ||
    filters.search.trim() !== "";

  const matchingCount = hasActiveFilters
    ? displayedReviews.length
    : (stats?.total_reviews ?? allReviews.length);

  function handleFilterChange(changes) {
    setFilters((prev) => ({ ...prev, ...changes }));
  }

  async function runAction(action, { success, failure }) {
    try {
      await action();
      showToast({ message: success, variant: "success" });
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: [merchantId, "merchant-reviews"],
        }),
        queryClient.invalidateQueries({
          queryKey: [merchantId, "merchant-reviews-stats"],
        }),
      ]);
      return true;
    } catch (err) {
      showToast({ message: err?.message || failure, variant: "error" });
      return false;
    }
  }

  const handleReply = (reviewId, comment) =>
    runAction(() => replyToReview(merchantId, reviewId, comment), {
      success: "Reply saved successfully",
      failure: "Failed to post reply",
    });

  const handleDeleteReply = (reviewId) =>
    runAction(() => deleteReviewReply(merchantId, reviewId), {
      success: "Reply deleted successfully",
      failure: "Failed to delete reply",
    });

  const handleReport = (reviewId, reason) =>
    runAction(() => reportReview(merchantId, reviewId, reason), {
      success: "Review reported for moderation",
      failure: "Failed to report review",
    });

  return (
    <div
      className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-3 py-4 sm:px-5
        sm:py-6"
    >
      <div className="flex flex-col gap-3">
        <div>
          <h1 className="text-xl font-semibold">Reviews</h1>
          <p className="text-fg-muted text-sm">
            Manage your client feedback, ratings, and replies.
          </p>
        </div>

        <ReviewSummary
          stats={stats}
          isLoading={isStatsLoading}
          isError={isStatsError}
          error={statsError}
          selectedRating={filters.stars}
          onSelectRating={(star) => handleFilterChange({ stars: star })}
        />
      </div>

      <ReviewFilters
        filters={filters}
        onChange={handleFilterChange}
        services={services}
        team={team}
        totalCount={matchingCount}
      />

      {isReviewsError ? (
        <ServerError error={reviewsError?.message} />
      ) : isReviewsLoading ? (
        <Loading />
      ) : displayedReviews.length === 0 ? (
        <div
          className="text-fg-muted my-12 flex flex-col items-center gap-3
            text-center"
        >
          <Icon icon={StarIcon} styles="size-12 text-border-strong" />
          <p className="text-fg">
            {hasActiveFilters ? "No matching reviews" : "No reviews yet"}
          </p>
          <p className="max-w-md text-sm">
            {hasActiveFilters
              ? "Try adjusting or clearing your filters to see more client reviews."
              : "When customers complete their bookings and leave feedback, their ratings and reviews will appear here."}
          </p>
          {hasActiveFilters && (
            <Button
              variant="secondary"
              styles="px-3 py-1.5 text-sm"
              buttonText="Clear filters"
              type="button"
              onClick={() => setFilters(INITIAL_FILTERS)}
            />
          )}
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          {displayedReviews.map((review) => (
            <ReviewCard
              key={review.id}
              review={review}
              onReply={handleReply}
              onDeleteReply={handleDeleteReply}
              onReport={setReportingReview}
            />
          ))}

          {hasNextPage && (
            <Button
              variant="secondary"
              styles="w-full py-2.5 text-sm"
              buttonText="Load more reviews"
              type="button"
              isLoading={isFetchingNextPage}
              onClick={() => fetchNextPage()}
            />
          )}
        </div>
      )}

      <ReportReviewModal
        isOpen={Boolean(reportingReview)}
        review={reportingReview}
        onClose={() => setReportingReview(null)}
        onSubmit={handleReport}
      />
    </div>
  );
}
