import { Search01Icon } from "@hugeicons/core-free-icons";
import {
  Button,
  Icon,
  Loading,
  SearchInput,
  ServerError,
  Toggle,
  ToggleGroup,
} from "@reservations/components";
import {
  formatDuration,
  getDisplayPrice,
  merchantServicesQueryOptions,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import ServiceDetails from "./ServiceDetails";

export default function ServiceSection({ router, merchantInfo, merchantName }) {
  const {
    data: categories,
    isLoading,
    isError,
    error,
  } = useQuery({ ...merchantServicesQueryOptions(merchantName) });

  const [activeCategoryId, setActiveCategoryId] = useState("");
  const [searchText, setSearchText] = useState("");
  const [selectedService, setSelectedService] = useState(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);

  const showToggles = categories?.length > 1;
  const currentCategoryId =
    activeCategoryId || (categories?.[0]?.id ?? "uncategorized");

  const displayCategories = useMemo(() => {
    if (!categories) return [];

    let active = categories?.filter(
      (category) => (category?.id ?? "uncategorized") === currentCategoryId
    );

    if (!showToggles && searchText.trim().length > 0) {
      const searchTextLow = searchText.toLowerCase();
      active = active.map((category) => ({
        ...category,
        services: category.services.filter((service) =>
          service.name?.toLowerCase().includes(searchTextLow)
        ),
      }));
    }

    return active;
  }, [categories, currentCategoryId, showToggles, searchText]);

  const selectedCategoryName = categories?.find(
    (c) => c.id === selectedService?.category_id
  )?.name;

  if (isError) {
    return <ServerError error={error.message} />;
  }

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div className="flex h-full w-full flex-col">
      <ServiceDetails
        merchantName={merchantName}
        locationId={merchantInfo.location_id}
        service={selectedService}
        isOpen={isDetailsOpen}
        onClose={() => {
          setSelectedService(null);
          setIsDetailsOpen(false);
        }}
        category={selectedCategoryName}
        router={router}
      />
      <div className="flex w-full flex-col gap-5 pb-5">
        {!showToggles && (
          <SearchInput
            searchText={searchText}
            onChange={(text) => setSearchText(text)}
            placeholder="Search for a service..."
            styles="w-full"
          />
        )}

        {showToggles && (
          <div className="flex w-full pb-1">
            <ToggleGroup
              styles="w-full"
              multiple={false}
              value={currentCategoryId}
              onValueChange={(val) => {
                setActiveCategoryId(val);
              }}
            >
              {categories.map((category) => {
                const catId = category.id ?? "uncategorized";
                return (
                  <Toggle key={catId} value={catId}>
                    {category.name || "Uncategorized"}
                  </Toggle>
                );
              })}
            </ToggleGroup>
          </div>
        )}
      </div>

      <div className="flex-1">
        {displayCategories.length > 0 &&
        displayCategories[0].services.length > 0 ? (
          <div className="flex flex-col gap-8">
            {displayCategories.map((category) => {
              const catId = category.id ?? "uncategorized";

              return (
                <div key={catId} className="flex flex-col gap-3">
                  <div className="flex flex-col gap-4">
                    {category.services.map((service) => (
                      <ServiceItem
                        key={service.id}
                        service={service}
                        router={router}
                        locationId={merchantInfo.location_id}
                        onClick={(clickedService) => {
                          setSelectedService(clickedService);
                          setIsDetailsOpen(true);
                        }}
                      />
                    ))}
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="mt-10 flex flex-col items-center justify-start gap-4">
            {!showToggles && searchText.trim().length > 0 ? (
              <>
                <Icon icon={Search01Icon} styles="size-10 text-fg-disabled!" />
                <p className="text-fg-muted">
                  No services found for "{searchText}"
                </p>
              </>
            ) : (
              <p className="text-fg-muted">
                No services available in this category.
              </p>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

function ServiceItem({ service, router, locationId, onClick }) {
  return (
    <div
      type="button"
      onClick={() => onClick?.(service)}
      className="group border-border bg-bg-raised shadow-card
        hover:border-border-strong flex w-full cursor-pointer items-center
        justify-between gap-4 rounded-xl border p-4 text-left"
    >
      <div className="flex flex-col gap-2.5">
        <p className="text-fg leading-tight font-semibold">{service.name}</p>

        <div className="text-fg-muted flex items-center gap-2 text-sm">
          <span>
            {service.min_duration &&
            service.max_duration &&
            service.min_duration !== service.max_duration
              ? `${formatDuration(service.min_duration)} - ${formatDuration(service.max_duration)}`
              : formatDuration(service.total_duration)}
          </span>
          <span className="bg-fg-muted size-1 rounded-full"></span>
          <span className="text-fg">
            {getDisplayPrice(service.price, service.price_type)}
          </span>
        </div>
      </div>
      <Link
        from={router.fullPath}
        to="book"
        search={{
          locationId: locationId,
          serviceId: service.id,
          type: service.booking_type,
        }}
      >
        <Button
          variant="primary"
          styles="py-1.5 px-3 text-sm"
          name="Reserve"
          buttonText="Reserve"
        />
      </Link>
    </div>
  );
}
