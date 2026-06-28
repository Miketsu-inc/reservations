import { Search01Icon } from "@hugeicons/core-free-icons";
import {
  CloseButton,
  Icon,
  SearchInput,
  Toggle,
  ToggleGroup,
} from "@reservations/components";
import { useMemo, useState } from "react";
import { ServiceCard } from "./BookingCards";

export default function ServiceSelector({
  categories,
  onSelect,
  isWindowSmall,
  onClose,
  isNested,
}) {
  const [searchText, setSearchText] = useState("");
  const [filterType, setFilterType] = useState("all");

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
      <div
        className="border-border_color flex flex-col gap-7 border-b px-6 pb-5"
      >
        <div className="flex flex-col gap-4">
          <p className="text-2xl font-semibold">Select a service</p>
          <SearchInput
            searchText={searchText}
            onChange={(text) => setSearchText(text)}
            placeholder="Search client..."
            styles="w-full p-3"
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
          <div className="flex flex-col gap-8">
            {filteredCategories.map((category) => (
              <div
                key={category.id ?? "uncategorized"}
                className="flex flex-col gap-3"
              >
                <div className="flex items-center gap-3">
                  <span className="text-text_color text-lg font-medium">
                    {category.name || "Other"}
                  </span>
                  <div
                    className="flex size-6 items-center justify-center
                      rounded-full bg-gray-200 font-medium dark:bg-gray-400/20"
                  >
                    {category.services.length}
                  </div>
                </div>

                <div className="flex flex-col gap-4">
                  {category.services.map((service) => (
                    <ServiceCard
                      key={service.id}
                      service={service}
                      onClick={() => onSelect(service)}
                      styles="py-2.5!"
                    />
                  ))}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="mt-10 flex flex-col items-center justify-start gap-4">
            <Icon icon={Search01Icon} styles="size-15 text-gray-500!" />
            <p className="text-gray-40 dark:text-gray-500">No services found</p>
          </div>
        )}
      </div>
    </div>
  );
}
