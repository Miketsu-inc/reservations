import Button from "@components/Button";
import DeleteModal from "@components/DeleteModal";
import { Popover, PopoverContent, PopoverTrigger } from "@components/Popover";
import SearchInput from "@components/SearchInput";
import ServerError from "@components/ServerError";
import PlusIcon from "@icons/PlusIcon";
import { useToast, useWindowSize } from "@lib/hooks";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { createFileRoute, Link, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import AddServiceCategoryModal from "./-components/AddServiceCategoryModal";
import ServiceCard from "./-components/ServiceCard";
import ServiceCategoryCard from "./-components/ServiceCategoryCard";

async function fetchServices() {
  const response = await fetch(`/api/v1/merchants/services`, {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalSotrageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

export const Route = createFileRoute("/_authenticated/_sidepanel/services/")({
  component: ServicesPage,
  loader: async () => {
    const services = await fetchServices();

    return {
      crumb: "Services",
      services: services,
    };
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function ServicesPage() {
  const router = useRouter();
  const loaderData = Route.useLoaderData();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();

  const windowSize = useWindowSize();
  const [selected, setSelected] = useState({ id: 0, name: "" });
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [showAddCategoryModal, setShowAddCategoryModal] = useState(false);
  const [searchText, setSearchText] = useState("");

  const isWindowSmall = windowSize === "sm" || windowSize === "md";

  const filteredServicesGroupedByCategories = loaderData.services.map(
    (category) => ({
      ...category,
      services: category.services.filter((service) =>
        service.name.toLowerCase().includes(searchText.toLowerCase())
      ),
    })
  );

  async function deleteHandler(selected) {
    try {
      const response = await fetch(
        `/api/v1/merchants/services/${selected.id}`,
        {
          method: "DELETE",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
        }
      );

      if (!response.ok) {
        invalidateLocalSotrageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        router.invalidate();
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

  return (
    <div className="flex h-full flex-col px-4 py-2 md:px-0 md:py-0">
      <DeleteModal
        itemName={selected.name}
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onDelete={() => deleteHandler(selected)}
      />
      <AddServiceCategoryModal
        isOpen={showAddCategoryModal}
        onClose={() => setShowAddCategoryModal(false)}
        onAdded={router.invalidate}
      />
      <div className="flex w-full flex-col gap-8 py-6">
        <p className="text-xl">Services</p>
        <ServerError error={serverError} />
        <div className="flex flex-row items-center justify-between">
          <SearchInput
            searchText={searchText}
            onChange={(text) => setSearchText(text)}
          />
          {isWindowSmall ? (
            <Popover>
              <PopoverTrigger asChild>
                <Button styles="p-2" variant="primary">
                  <PlusIcon styles="size-6" />
                </Button>
              </PopoverTrigger>
              <PopoverContent align="end">
                <div className="*:hover:bg-hvr_gray flex flex-col items-start *:w-full *:rounded-lg *:p-2">
                  <Link from={Route.fullPath} to="/services/new">
                    New service
                  </Link>
                  <button
                    onClick={() => setShowAddCategoryModal(true)}
                    className="cursor-pointer text-left"
                  >
                    New category
                  </button>
                </div>
              </PopoverContent>
            </Popover>
          ) : (
            <div className="flex flex-row items-center gap-4">
              <Button
                styles="py-2 px-4"
                variant="secondary"
                buttonText="New category"
                onClick={() => setShowAddCategoryModal(true)}
              >
                <PlusIcon styles="size-5 mr-1" />
              </Button>
              <Link from={Route.fullPath} to="/services/new">
                <Button
                  styles="py-2 px-4"
                  variant="primary"
                  buttonText="New service"
                >
                  <PlusIcon styles="size-5 mr-1" />
                </Button>
              </Link>
            </div>
          )}
        </div>
      </div>
      <div className="py-6">
        <ul className="flex flex-wrap gap-4">
          {filteredServicesGroupedByCategories.map((category) => (
            <li className="w-full" key={category.id}>
              <ServiceCategoryCard
                category={category}
                refresh={router.invalidate}
              >
                <ul className="flex flex-wrap gap-4">
                  {category.services.map((service) => (
                    <li className="w-full md:w-fit" key={service.id}>
                      <ServiceCard
                        isWindowSmall={isWindowSmall}
                        service={service}
                        onDelete={() => {
                          setSelected({ name: service.name, id: service.id });
                          setShowDeleteModal(true);
                        }}
                        onEdit={() =>
                          router.navigate({
                            from: Route.fullPath,
                            to: `/services/edit/${service.id}`,
                          })
                        }
                        refresh={router.invalidate}
                      />
                    </li>
                  ))}
                </ul>
              </ServiceCategoryCard>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
