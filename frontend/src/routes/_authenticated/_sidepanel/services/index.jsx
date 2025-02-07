import ServerError from "@components/ServerError";
import { useToast } from "@lib/hooks";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import ServiceModal from "./-components/ServiceModal";
import ServicesTable from "./-components/ServicesTable";

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
  const [showServiceModal, setShowServiceModal] = useState(false);
  const [modalData, setModalData] = useState();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();

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

  async function modalHandler(service) {
    let url = "";
    let method = "";

    // means that the service was already added and now should be modified
    if (service.id != null) {
      url = `/api/v1/merchants/services/${service.id}`;
      method = "PUT";
    } else {
      // for correct json sending
      delete service.id;
      url = "/api/v1/merchants/services";
      method = "POST";
    }

    try {
      const response = await fetch(url, {
        method: method,
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify(service),
      });

      if (!response.ok) {
        invalidateLocalSotrageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        setServerError();
        router.invalidate();
        setModalData();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  return (
    <div className="flex h-screen justify-center px-4">
      <ServiceModal
        data={modalData}
        isOpen={showServiceModal}
        onClose={() => setShowServiceModal(false)}
        onSubmit={modalHandler}
      />
      <div className="w-full md:w-3/4">
        <ServerError error={serverError} />
        <p className="text-xl">Services</p>
        <ServicesTable
          servicesData={loaderData}
          onNewItem={() => {
            // the first condition is necessary for it to not cause an error
            // in case of a new item
            if (modalData && modalData.id) {
              setModalData();
            }
            setShowServiceModal(true);
          }}
          onEdit={(service) => {
            setModalData(service);
            setShowServiceModal(true);
          }}
          onDelete={deleteHandler}
        />
      </div>
    </div>
  );
}
