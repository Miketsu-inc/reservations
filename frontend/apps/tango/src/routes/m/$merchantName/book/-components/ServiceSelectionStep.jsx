import {
  Clock01Icon,
  PlusSignIcon,
  Search01Icon,
  Tick02Icon,
  UserGroupIcon,
} from "@hugeicons/core-free-icons";
import {
  Avatar,
  Icon,
  SearchInput,
  ServerError,
  Toggle,
  ToggleGroup,
} from "@reservations/components";
import {
  formatDuration,
  getDisplayPrice,
  useActiveSection,
} from "@reservations/lib";
import merchantServicesQueryOptions from "@reservations/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { StepContentSkeleton } from "./StepContentSkeleton";

// implement fetching services by employee and location later
export default function ServiceSelectionStep({
  merchantName,
  _locationId,
  _employeeId,
  onServiceSelect,
  employee,
}) {
  const {
    data: categories,
    isLoading,
    isError,
    error,
  } = useQuery({ ...merchantServicesQueryOptions(merchantName) });

  const [searchText, setSearchText] = useState("");
  const [selectedService, setSelectedService] = useState(null);
  const showToggles = categories?.length > 1;

  const displayCategories = useMemo(() => {
    if (!categories) return [];
    if (!searchText.trim()) return categories;

    const searchTextLow = searchText.toLowerCase();

    return categories
      .map((category) => ({
        ...category,
        services: category.services.filter((service) =>
          service.name?.toLowerCase().includes(searchTextLow)
        ),
      }))
      .filter((category) => category.services.length > 0);
  }, [categories, searchText]);

  const categoryIds = useMemo(
    () => displayCategories.map((c) => String(c.id ?? "uncategorized")),
    [displayCategories]
  );

  const activeCategoryId = useActiveSection(categoryIds);
  const currentCategoryId =
    activeCategoryId || categoryIds[0] || "uncategorized";

  function scrollToCategory(id) {
    const element = document.getElementById(id);
    if (element) {
      element.scrollIntoView({ behavior: "smooth", block: "start" });
    }
  }

  function handleServiceSelect(service) {
    if (selectedService?.id === service.id) {
      setSelectedService(null);
      onServiceSelect(null);
    } else {
      setSelectedService(service);
      onServiceSelect(service);
    }
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  if (isLoading) {
    return <StepContentSkeleton />;
  }

  return (
    <div className="flex h-full w-full flex-col gap-6">
      <h1 className="text-3xl font-bold">Select a Service</h1>
      {employee?.first_name && (
        <div
          className="bg-layer_bg border-border_color flex w-fit items-center
            gap-2 rounded-full border py-1.5 pr-3 pl-2"
        >
          <Avatar
            styles="size-8! text-[12px]! shrink-0 rounded-full!"
            img={employee?.avatar_url}
            initials={`${employee.first_name[0]}${employee.last_name[0]}`}
          />
          <span className="text-sm font-medium">
            {employee.first_name} {employee.last_name}
          </span>
        </div>
      )}
      <div className="flex flex-1 flex-col">
        <div
          className="bg-bg_color sticky top-14.75 z-10 flex w-full flex-col
            gap-5 py-4 lg:top-17"
        >
          {!showToggles && (
            <SearchInput
              searchText={searchText}
              onChange={(text) => setSearchText(text)}
              placeholder="Search for a service..."
              styles="w-full p-3"
            />
          )}
          {showToggles && (
            <div className="flex w-full pb-1">
              <ToggleGroup
                styles="w-full"
                value={currentCategoryId}
                onValueChange={(val) => {
                  scrollToCategory(val);
                }}
              >
                {categories.map((category) => {
                  const catId = String(category.id ?? "uncategorized");
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
          {displayCategories.length > 0 ? (
            <div className="flex flex-col gap-10">
              {displayCategories.map((category) => {
                // if not conveted to string there would be  a mismatch in type when reading the id for intesection observer
                const catId = String(category.id ?? "uncategorized");

                return (
                  <div
                    key={catId}
                    id={catId}
                    className="flex scroll-mt-60 flex-col gap-4"
                  >
                    <div className="text-text_color font-semibold">
                      {category.name || "Uncategorized"}
                    </div>

                    <ul className="flex flex-col gap-4">
                      {category.services.map((service) => (
                        <ServiceItem
                          key={service.id}
                          service={service}
                          isSelected={selectedService?.id === service.id}
                          onSelect={handleServiceSelect}
                        />
                      ))}
                    </ul>
                  </div>
                );
              })}
            </div>
          ) : (
            <div
              className="mt-10 flex flex-col items-center justify-start gap-4"
            >
              {searchText.trim().length > 0 ? (
                <>
                  <Icon icon={Search01Icon} styles="size-10 text-gray-500" />
                  <p className="text-gray-400 dark:text-gray-500">
                    No services found for "{searchText}"
                  </p>
                </>
              ) : (
                <p className="text-gray-400 dark:text-gray-500">
                  No services available.
                </p>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function ServiceItem({ service, isSelected, onSelect }) {
  const isGroup = service.booking_type !== "appointment";

  return (
    <li
      role="radio"
      aria-checked={isSelected}
      onClick={() => onSelect(service)}
      className={`bg-layer_bg text-text_color border-border_color flex w-full
        cursor-pointer items-center justify-between gap-4 rounded-md border p-4
        transition-all duration-200 hover:bg-gray-50 dark:hover:bg-gray-200/5
        ${isSelected ? "ring-primary ring-1" : ""} `}
    >
      <div className="flex flex-col gap-2">
        <p className="text-[17px] font-medium">{service.name}</p>
        <div className="text-text_color/80 flex items-center gap-5 text-sm">
          <div className="flex items-center gap-1.5">
            <Icon icon={Clock01Icon} styles="size-4" />
            <span>{formatDuration(service.total_duration)}</span>
          </div>

          {isGroup && (
            <div className="text-text_color/80 flex items-center gap-1.5">
              <Icon icon={UserGroupIcon} styles="size-5" />
              <span>Max {service.max_participants}</span>
            </div>
          )}
        </div>
        {service.description && (
          <div
            className="line-clamp-2 text-sm text-gray-500 dark:text-gray-300/80"
          >
            {service.description}
          </div>
        )}
        <span className="text-[17px]">
          {getDisplayPrice(service.price, service.price_type)}
        </span>
      </div>
      <div
        className={`flex size-8 shrink-0 items-center justify-center
          rounded-full border transition-colors ${
            isSelected ? "border-primary bg-primary" : "border-gray-400"
          } `}
      >
        <Icon
          icon={isSelected ? Tick02Icon : PlusSignIcon}
          styles={`
            ${isSelected ? "text-white size-6" : "text-gray-400 size-5"}`}
        />
      </div>
    </li>
  );
}
