import { Search01Icon } from "@hugeicons/core-free-icons";
import {
  Button,
  Icon,
  SearchInput,
  Toggle,
  ToggleGroup,
} from "@reservations/components";
import { formatDuration, getDisplayPrice } from "@reservations/lib";
import { Link } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import ServiceDetails from "./ServiceDetails";

export default function ServiceSection({
  categories,
  isWindowSmall,
  router,
  merchantInfo,
}) {
  const [activeCategoryId, setActiveCategoryId] = useState(
    categories.length > 0 ? (categories[0].id ?? "uncategorized") : ""
  );
  const [searchText, setSearchText] = useState("");
  const [selectedService, setSelectedService] = useState(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);

  const shouldShowToggles = categories.length > 1;

  const isOnlyUncategorized =
    categories.length === 1 && categories[0].id == null;

  const displayCategories = useMemo(() => {
    let active = categories.filter(
      (category) => (category.id ?? "uncategorized") === activeCategoryId
    );

    if (isOnlyUncategorized && searchText.trim().length > 0) {
      const searchTextLow = searchText.toLowerCase();
      active = active.map((category) => ({
        ...category,
        services: category.services.filter((service) =>
          service.name?.toLowerCase().includes(searchTextLow)
        ),
      }));
    }

    return active;
  }, [categories, activeCategoryId, isOnlyUncategorized, searchText]);

  return (
    <div className="flex h-full w-full flex-col">
      <ServiceDetails
        merchantName={merchantInfo.merchant_name}
        locationId={merchantInfo.location_id}
        service={selectedService}
        isOpen={isDetailsOpen}
        onClose={() => {
          setSelectedService(null);
          setIsDetailsOpen(false);
        }}
        isWindowSmall={isWindowSmall}
        category={categories[0].name}
      />
      <div className="flex flex-col gap-5 pb-5">
        {isOnlyUncategorized && (
          <SearchInput
            searchText={searchText}
            onChange={(text) => setSearchText(text)}
            placeholder="Search for a service..."
            styles="w-full p-3"
          />
        )}

        {shouldShowToggles && (
          <div className="hide-scrollbar flex w-full overflow-x-auto pb-1">
            <ToggleGroup
              multiple={false}
              value={activeCategoryId}
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

      <div className="flex-1 overflow-y-auto dark:scheme-dark">
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
            {isOnlyUncategorized && searchText.trim().length > 0 ? (
              <>
                <Icon icon={Search01Icon} styles="size-15 text-gray-500!" />
                <p className="text-gray-400 dark:text-gray-500">
                  No services found for "{searchText}"
                </p>
              </>
            ) : (
              <p className="text-gray-400 dark:text-gray-500">
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
      className="group border-border_color bg-layer_bg hover:bg-gray-30 flex
        w-full items-center justify-between gap-4 rounded-md border p-4
        text-left shadow-sm transition-all duration-200
        dark:hover:bg-gray-200/5"
    >
      <div className="flex flex-col gap-2.5">
        <p className="text-text_color text-[17px] leading-tight font-semibold">
          {service.name}
        </p>

        <div className="flex items-center gap-2">
          <span>{formatDuration(service.total_duration)}</span>
          <span className="size-1 rounded-full bg-gray-500 dark:bg-gray-400"></span>
          <span className="text-text_color">
            {getDisplayPrice(service.price, service.price_type)}
          </span>
        </div>
      </div>
      <Link
        from={router.fullPath}
        to="booking"
        search={{
          locationId: locationId,
          serviceId: service.id,
        }}
      >
        <Button
          variant="primary"
          styles="py-1.5 px-2"
          name="Reserve"
          buttonText="Reserve"
        />
      </Link>
    </div>
  );
}
