import {
  ArrowLeft01Icon,
  Cancel01Icon,
  StarIcon,
} from "@hugeicons/core-free-icons";
import {
  Button,
  Icon,
  Loading,
  MultiSelect,
  SearchInput,
  Select,
  ServerError,
} from "@reservations/components";
import {
  activeTeamQueryOptions,
  formatShortName,
  merchantInfoQueryOptions,
  merchantServicesQueryOptions,
  publicReviewsInfiniteQueryOptions,
  publicReviewStatsQueryOptions,
  toAvatar,
} from "@reservations/lib";
import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import ReviewFiltersSidebar from "./-components/ReviewFiltersSidebar";
import ReviewItem from "./-components/ReviewItem";
import ReviewStats from "./-components/ReviewStats";

const SORT_OPTIONS = [
  { value: "most_relevant", label: "Most relevant" },
  { value: "newest", label: "Newest" },
  { value: "oldest", label: "Oldest" },
  { value: "highest", label: "Highest rating" },
  { value: "lowest", label: "Lowest rating" },
];

const INITIAL_FILTERS = {
  stars: "all",
  employeeIds: [],
  serviceIds: [],
  sort: "most_relevant",
};

export const Route = createFileRoute("/m/$merchantName/$locationId/reviews/")({
  component: ReviewsPage,
  loader: ({
    params: { merchantName, locationId },
    context: { queryClient },
  }) => {
    void queryClient.prefetchQuery(merchantInfoQueryOptions(merchantName));
    void queryClient.prefetchQuery(
      publicReviewStatsQueryOptions(merchantName, locationId)
    );
    void queryClient.prefetchInfiniteQuery(
      publicReviewsInfiniteQueryOptions(
        merchantName,
        locationId,
        INITIAL_FILTERS
      )
    );
    void queryClient.prefetchQuery(
      activeTeamQueryOptions(merchantName, locationId)
    );
    void queryClient.prefetchQuery(merchantServicesQueryOptions(merchantName));
  },
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});

function ReviewsPage() {
  const { merchantName, locationId } = Route.useParams();

  const [isScrolled, setIsScrolled] = useState(false);
  const [filters, setFilters] = useState(INITIAL_FILTERS);
  const [search, setSearch] = useState("");

  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 50);
    };
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  const {
    data: merchantInfo,
    isLoading: isMerchantInfoLoading,
    isError: isMerchantInfoError,
    error: merchantInfoError,
  } = useQuery(merchantInfoQueryOptions(merchantName));

  const {
    data: stats,
    isLoading: isStatsLoading,
    isError: isStatsError,
    error: statsError,
  } = useQuery(publicReviewStatsQueryOptions(merchantName, locationId));

  const {
    data,
    isLoading: isReviewsLoading,
    isError: isReviewsError,
    error: reviewsError,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery(
    publicReviewsInfiniteQueryOptions(merchantName, locationId, filters)
  );

  const { data: team = [] } = useQuery(
    activeTeamQueryOptions(merchantName, locationId)
  );

  const { data: serviceCategories = [] } = useQuery(
    merchantServicesQueryOptions(merchantName)
  );

  const services = serviceCategories.flatMap((category) => category.services);

  if (isMerchantInfoError || isStatsError || isReviewsError) {
    return (
      <ServerError
        error={(merchantInfoError ?? statsError ?? reviewsError).message}
      />
    );
  }

  if (isMerchantInfoLoading || isStatsLoading || isReviewsLoading) {
    return <Loading />;
  }

  const reviews = data.pages.flatMap((page) => page.reviews);

  const term = search.trim().toLowerCase();
  const displayedReviews = term
    ? reviews.filter((review) =>
        [
          review.comment,
          review.service_name,
          formatShortName(review.customer),
          review.employee && formatShortName(review.employee),
        ].some((text) => text?.toLowerCase().includes(term))
      )
    : reviews;

  function handleFilterChange(changes) {
    setFilters((prev) => ({ ...prev, ...changes }));
  }

  function clearFilters() {
    setFilters(INITIAL_FILTERS);
    setSearch("");
  }

  const activeFilters = [
    filters.stars !== "all" && {
      id: "stars",
      label: `${filters.stars} ${filters.stars === 1 ? "star" : "stars"}`,
      onClear: () => handleFilterChange({ stars: "all" }),
    },
    ...filters.serviceIds.map((id) => ({
      id: `service-${id}`,
      label: services.find((service) => service.id === id)?.name ?? "Service",
      onClear: () =>
        handleFilterChange({
          serviceIds: filters.serviceIds.filter(
            (serviceId) => serviceId !== id
          ),
        }),
    })),
    ...filters.employeeIds.map((id) => ({
      id: `employee-${id}`,
      label: formatShortName(
        team.find((member) => member.id === id),
        "Team member"
      ),
      onClear: () =>
        handleFilterChange({
          employeeIds: filters.employeeIds.filter(
            (employeeId) => employeeId !== id
          ),
        }),
    })),
  ].filter(Boolean);

  const hasActiveFilters = activeFilters.length > 0 || term !== "";

  return (
    <div className="bg-bg text-fg min-h-screen w-full">
      <div
        className={`bg-bg sticky top-0 z-50 flex items-center justify-between
          py-3 transition-colors duration-200 ${
            isScrolled ? "border-border border-b shadow-xs" : "border-none"
          }`}
      >
        <div
          className="mx-auto flex w-full max-w-7xl items-center gap-4 px-4
            sm:px-6 lg:px-12"
        >
          <Link
            to="/m/$merchantName"
            params={{ merchantName }}
            className="hover:bg-hover md:bg-bg-raised border-border
              md:shadow-card flex h-fit w-fit cursor-pointer rounded-full p-1
              transition-colors md:border md:p-2.5"
            aria-label={`Back to ${merchantInfo.merchant_name}`}
          >
            <Icon icon={ArrowLeft01Icon} styles="text-fg size-5" />
          </Link>

          <div
            className={`text-xl font-semibold transition-all duration-200 ${
              isScrolled
                ? "translate-y-0 opacity-100"
                : "-translate-y-4 opacity-0"
              }`}
          >
            Reviews
          </div>
        </div>
      </div>

      <div
        className="mx-auto flex w-full max-w-7xl flex-col gap-8 px-4 pt-3 pb-12
          sm:px-6 lg:px-12"
      >
        <div className="flex flex-col gap-2">
          <h1 className="text-3xl font-bold tracking-tight lg:text-4xl">
            Reviews
          </h1>
          <p className="text-fg-muted text-sm">From verified visits</p>
        </div>

        <div className="grid grid-cols-1 gap-8 lg:grid-cols-12 lg:gap-12">
          <div className="lg:col-span-3">
            <ReviewStats
              stats={stats}
              selectedStar={filters.stars}
              onSelectStar={(stars) => handleFilterChange({ stars })}
            />
          </div>

          <div className="flex flex-col gap-10 lg:col-span-6">
            <div className="flex flex-col gap-3">
              <div className="flex flex-col items-center gap-3 sm:flex-row">
                <div className="w-full flex-1">
                  <SearchInput
                    placeholder="Search reviews"
                    searchText={search}
                    onChange={setSearch}
                  />
                </div>

                <div className="w-full shrink-0 sm:w-44">
                  <Select
                    options={SORT_OPTIONS}
                    value={filters.sort}
                    onSelect={(option) =>
                      handleFilterChange({ sort: option.value })
                    }
                    placeholder="Sort by"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3 lg:hidden">
                <MultiSelect
                  options={services.map((service) => ({
                    value: service.id,
                    label: service.name,
                  }))}
                  values={filters.serviceIds}
                  onSelect={(serviceIds) => handleFilterChange({ serviceIds })}
                  placeholder="Select services"
                  displayText="service"
                />

                <MultiSelect
                  options={team.map((member) => {
                    const { img, initials } = toAvatar(member);
                    return {
                      value: member.id,
                      label: formatShortName(member),
                      img,
                      initials,
                    };
                  })}
                  values={filters.employeeIds}
                  onSelect={(employeeIds) =>
                    handleFilterChange({ employeeIds })
                  }
                  placeholder="Select team"
                  displayText="team member"
                />
              </div>

              {activeFilters.length > 0 && (
                <div className="flex flex-wrap items-center gap-2">
                  {activeFilters.map((filter) => (
                    <span
                      key={filter.id}
                      className="bg-fg text-bg inline-flex items-center gap-1.5
                        rounded-full px-3 py-1 text-xs"
                    >
                      <span>{filter.label}</span>
                      <button
                        type="button"
                        onClick={filter.onClear}
                        aria-label={`Remove filter ${filter.label}`}
                      >
                        <Icon icon={Cancel01Icon} styles="size-3.5" />
                      </button>
                    </span>
                  ))}
                  <button
                    type="button"
                    onClick={clearFilters}
                    className="text-fg-muted hover:text-fg cursor-pointer
                      text-xs"
                  >
                    Clear all
                  </button>
                </div>
              )}
            </div>

            {displayedReviews.length === 0 ? (
              <div
                className="text-fg-muted flex flex-col items-center gap-3 py-2
                  text-center"
              >
                <Icon icon={StarIcon} styles="size-12 text-border-strong" />
                <p className="text-fg">
                  {hasActiveFilters ? "No matching reviews" : "No reviews yet"}
                </p>
                <p className="max-w-md text-sm">
                  {hasActiveFilters
                    ? "Try adjusting or clearing your filters to see more reviews."
                    : "This business has no reviews yet."}
                </p>
                {hasActiveFilters && (
                  <Button
                    variant="secondary"
                    styles="px-3 py-1.5 text-sm"
                    buttonText="Clear filters"
                    type="button"
                    onClick={clearFilters}
                  />
                )}
              </div>
            ) : (
              <div className="flex flex-col gap-10">
                <div className="divide-border/40 flex flex-col divide-y">
                  {displayedReviews.map((review) => (
                    <ReviewItem
                      key={review.id}
                      review={review}
                      merchantInfo={merchantInfo}
                    />
                  ))}
                </div>

                {hasNextPage && (
                  <div className="flex justify-center">
                    <Button
                      variant="secondary"
                      buttonText="Load more reviews"
                      type="button"
                      isLoading={isFetchingNextPage}
                      onClick={() => fetchNextPage()}
                      styles="px-6 py-2.5 text-sm"
                    />
                  </div>
                )}
              </div>
            )}
          </div>

          <div className="hidden lg:col-span-3 lg:block">
            <ReviewFiltersSidebar
              services={services}
              serviceCounts={stats.service_counts}
              team={team}
              filters={filters}
              onChange={handleFilterChange}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
