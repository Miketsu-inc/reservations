import { ArrowDown01Icon } from "@hugeicons/core-free-icons";
import {
  Button,
  DeleteModal,
  Icon,
  Loading,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
  SearchInput,
  ServerError,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import AddServiceCategoryModal from "./-components/AddServiceCategoryModal";
import ServiceCard from "./-components/ServiceCard";
import ServiceCategorySection from "./-components/ServiceCategorySection";

async function fetchServices(merchantId) {
  const response = await fetch(`/api/v1/merchants/${merchantId}/services`, {
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

function servicesQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "services"],
    queryFn: () => fetchServices(merchantId),
  });
}

function reorderArray(items, itemId, direction) {
  const currentIndex = items.findIndex((item) => item.id === itemId);
  const itemIds = items.map((item) => item.id);

  // remove id from item ids
  itemIds.splice(currentIndex, 1);

  if (direction === "forward") {
    if (currentIndex === 0) return null;
    // put id one index before it's original spot
    itemIds.splice(currentIndex - 1, 0, itemId);
  } else if (direction === "backward") {
    if (currentIndex === items.length - 1) return null;
    // put id on index after it's original spot
    itemIds.splice(currentIndex + 1, 0, itemId);
  } else {
    console.error("wrong direction ", direction);
    return null;
  }

  return itemIds;
}

export const Route = createFileRoute("/_authenticated/_sidepanel/services/")({
  component: ServicesPage,
  loader: async ({
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => {
    await queryClient.ensureQueryData(servicesQueryOptions(merchantId));
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function ServicesPage() {
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();
  const { merchantId } = useAuth();

  const [selected, setSelected] = useState({ id: 0, name: "" });
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [showAddCategoryModal, setShowAddCategoryModal] = useState(false);
  const [searchText, setSearchText] = useState("");

  const { queryClient } = Route.useRouteContext({ from: Route.id });

  const {
    data: services,
    isLoading,
    isError,
    error,
  } = useQuery(servicesQueryOptions(merchantId));

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  async function invalidateServicesQuery() {
    await queryClient.invalidateQueries({
      queryKey: [merchantId, "services"],
    });
  }

  const filteredServicesGroupedByCategories = services.map((category) => ({
    ...category,
    services: category.services.filter((service) =>
      service.name.toLowerCase().includes(searchText.toLowerCase())
    ),
  }));

  async function deleteHandler(selected) {
    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/services/${selected.id}`,
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
        setServerError(result.error.message);
      } else {
        invalidateServicesQuery();
        setServerError();
        showToast({
          message: "Service deleted successfully",
          variant: "success",
        });
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  async function moveCategoryHandler(id, direction) {
    const categories = services.filter((category) => category.id !== null);
    if (categories.length === 1) return;

    const categoryIds = reorderArray(categories, id, direction);
    if (!categoryIds) return;

    const response = await fetch(
      `/api/v1/merchants/${merchantId}/service-categories/reorder`,
      {
        method: "PUT",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          categories: categoryIds,
        }),
      }
    );

    if (!response.ok) {
      invalidateLocalStorageAuth(response.status);
      const result = await response.json();
      setServerError(result.error.message);
    } else {
      invalidateServicesQuery();
      setServerError();
      showToast({
        message: "Service categories reordered successfully",
        variant: "success",
      });
    }
  }

  async function moveServiceHandler(categoryId, id, direction) {
    const servicesInCategory = services.find(
      (category) => category.id === categoryId
    ).services;
    if (servicesInCategory.length === 1) return;

    const serviceIds = reorderArray(servicesInCategory, id, direction);
    if (!serviceIds) return;

    const response = await fetch(
      `/api/v1/merchants/${merchantId}/services/reorder`,
      {
        method: "PUT",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          category_id: categoryId,
          services: serviceIds,
        }),
      }
    );

    if (!response.ok) {
      invalidateLocalStorageAuth(response.status);
      const result = await response.json();
      setServerError(result.error.message);
    } else {
      invalidateServicesQuery();
      setServerError();
      showToast({
        message: "Services reordered successfully",
        variant: "success",
      });
    }
  }

  return (
    <div className="flex justify-center">
      <div className="w-full max-w-4xl">
        <DeleteModal
          itemName={selected.name}
          isOpen={showDeleteModal}
          onClose={() => setShowDeleteModal(false)}
          onDelete={() => deleteHandler(selected)}
        />
        <AddServiceCategoryModal
          isOpen={showAddCategoryModal}
          onClose={() => setShowAddCategoryModal(false)}
          onAdded={invalidateServicesQuery}
        />
        <p className="pb-12 text-2xl">Services</p>
        <div className="flex flex-col gap-8">
          <div className="flex w-full flex-col">
            <ServerError error={serverError} />
            <div className="flex flex-row items-center justify-between">
              <SearchInput
                searchText={searchText}
                onChange={(text) => setSearchText(text)}
              />
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    styles="py-2 px-4"
                    childSide="right"
                    variant="primary"
                    buttonText="New"
                  >
                    <Icon icon={ArrowDown01Icon} styles="size-5 ml-2" />
                  </Button>
                </PopoverTrigger>
                <PopoverContent align="end">
                  <div
                    className="*:hover:bg-hvr_gray flex flex-col items-start
                      *:w-full *:rounded-lg *:p-2"
                  >
                    <Link from={Route.fullPath} to="/services/new">
                      1-on-1 service
                    </Link>
                    <Link from={Route.fullPath} to="/services/group/new">
                      Group service
                    </Link>
                    <PopoverClose asChild>
                      <button
                        onClick={() => setShowAddCategoryModal(true)}
                        className="cursor-pointer text-left"
                      >
                        Category
                      </button>
                    </PopoverClose>
                  </div>
                </PopoverContent>
              </Popover>
            </div>
          </div>
          <ul className="flex flex-col gap-4">
            {filteredServicesGroupedByCategories.map((category) => (
              <li className="w-full" key={category.id}>
                <ServiceCategorySection
                  category={category}
                  // -1 due to uncategorized. Which should always be the last
                  categoryCount={services.length - 1}
                  refresh={invalidateServicesQuery}
                  onMove={moveCategoryHandler}
                >
                  <ul className="flex flex-col gap-4">
                    {category.services.map((service) => (
                      <li className="w-full" key={service.id}>
                        <ServiceCard
                          service={service}
                          serviceCount={category.services.length}
                          onDelete={() => {
                            setSelected({ name: service.name, id: service.id });
                            setShowDeleteModal(true);
                          }}
                          refresh={invalidateServicesQuery}
                          onMove={async (id, direction) =>
                            await moveServiceHandler(category.id, id, direction)
                          }
                        />
                      </li>
                    ))}
                  </ul>
                </ServiceCategorySection>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  );
}
