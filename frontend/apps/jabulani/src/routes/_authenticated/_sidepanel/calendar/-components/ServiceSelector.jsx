import { Search01Icon } from "@hugeicons/core-free-icons";
import {
  CloseButton,
  Icon,
  SearchInput,
  Toggle,
  ToggleGroup,
} from "@reservations/components";
import { useWindowSize } from "@reservations/lib";
import { useMemo, useState } from "react";
import { ServiceCard } from "./BookingCards";

export default function ServiceSelector({
  categories,
  onSelect,
  onClose,
  isNested,
}) {
  const [searchText, setSearchText] = useState("");
  const [filterType, setFilterType] = useState("all");

  const { isWindowSmall } = useWindowSize();

  const hasMultipleBookingTypes = useMemo(() => {
    const allServices = categories.flatMap((c) => c.services);
    const hasAppointments = allServices.some(
      (s) => s.booking_type === "appointment"
    );
    const hasClasses = allServices.some((s) => s.booking_type === "class");
    return hasAppointments && hasClasses;
  }, [categories]);

  const filteredCategories = useMemo(() => {
    const searchTextLow = searchText.toLowerCase();
    return (
      categories
        .map((category) => {
          const categoryNameMatches = category.name
            ?.toLowerCase()
            .includes(searchTextLow);

          const matchingServices = category.services.filter((service) => {
            const matchesType =
              filterType === "all" || service.booking_type === filterType;
            if (!matchesType) return false;

            if (categoryNameMatches) return true;

            const nameMatches = service.name
              ?.toLowerCase()
              .includes(searchTextLow);

            return nameMatches;
          });
          return {
            ...category,
            services: matchingServices,
          };
        })
        //remove the empty categories
        .filter((category) => category.services.length > 0)
    );
  }, [categories, searchText, filterType]);

  return (
    <div
      className={`flex h-full w-full flex-col
        ${isWindowSmall || isNested ? "pt-0" : "pt-10"}`}
    >
      {isWindowSmall && !isNested && (
        <div className="flex items-center justify-end px-4 pt-5">
          <CloseButton onClick={onClose} styles="size-8" />
        </div>
      )}
      <div className="border-border flex flex-col gap-5 border-b px-6 pb-5">
        <div className="flex flex-col gap-4">
          <p className="text-xl font-semibold">Select a service</p>
          <SearchInput
            searchText={searchText}
            onChange={(text) => setSearchText(text)}
            placeholder="Search services..."
            styles="w-full"
          />
        </div>
        {hasMultipleBookingTypes && (
          <div className="text-sm">
            <ToggleGroup
              multiple={false}
              value={filterType}
              onValueChange={(type) => setFilterType(type)}
            >
              <Toggle value="all">All Services</Toggle>
              <Toggle value="appointment">1-on-1</Toggle>
              <Toggle value="class">Group</Toggle>
            </ToggleGroup>
          </div>
        )}
      </div>
      <div className="flex-1 overflow-y-auto px-6 py-5 dark:scheme-dark">
        {filteredCategories.length > 0 ? (
          <div className="flex flex-col gap-6">
            {filteredCategories.map((category) => (
              <div
                key={category.id ?? "uncategorized"}
                className="flex flex-col gap-3"
              >
                <div className="flex items-center gap-2">
                  <span className="text-fg-muted text-sm font-semibold">
                    {category.name || "Other"}
                  </span>
                  <div
                    className="bg-active text-fg-muted flex size-5 items-center
                      justify-center rounded-full text-xs"
                  >
                    {category.services.length}
                  </div>
                </div>

                <div className="flex flex-col gap-3">
                  {category.services.map((service) => (
                    <ServiceCard
                      key={service.id}
                      service={service}
                      onClick={() => onSelect(service)}
                    />
                  ))}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="mt-10 flex flex-col items-center justify-start gap-4">
            <Icon icon={Search01Icon} styles="size-10 text-fg-disabled!" />
            <p className="text-fg-muted text-sm">No services found</p>
          </div>
        )}
      </div>
    </div>
  );
}
