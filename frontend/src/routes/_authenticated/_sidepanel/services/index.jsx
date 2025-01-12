import Button from "@components/Button";
import Loading from "@components/Loading";
import SearchInput from "@components/SearchInput";
import ServerError from "@components/ServerError";
import PlusIcon from "@icons/PlusIcon";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { Suspense, useState } from "react";
import NewServiceModal from "./-components/NewServiceModal";
import ServicesTable from "./-components/ServicesTable";

async function fetchServices() {
  const response = await fetch(`/api/v1/merchants/services`, {
    method: "GET",
  });

  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    return result.data;
  }
}

export const Route = createFileRoute("/_authenticated/_sidepanel/services/")({
  component: ServicesPage,
  loader: () => fetchServices(),
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function ServicesPage() {
  const router = useRouter();
  const loaderData = Route.useLoaderData();
  const [showModal, setShowModal] = useState(false);
  const [searchText, setSearchText] = useState();

  return (
    <div className="flex h-screen justify-center px-4">
      <NewServiceModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        onSuccess={() => router.invalidate()}
      />
      <div className="w-full md:w-3/4">
        <p className="text-xl">Services</p>
        <div className="flex flex-row justify-between py-2">
          <SearchInput
            searchText={searchText}
            onChange={(text) => setSearchText(text)}
          />
          <Button
            onClick={() => setShowModal(true)}
            styles="p-2 text-sm"
            buttonText="New Service"
          >
            <PlusIcon styles="w-4 h-4 md:w-5 md:h-5 mr-1 text-white" />
          </Button>
        </div>
        <div className="h-2/3">
          <Suspense fallback={<Loading />}>
            <ServicesTable searchText={searchText} servicesData={loaderData} />
          </Suspense>
        </div>
      </div>
    </div>
  );
}
