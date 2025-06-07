import ServerError from "@components/ServerError";
import { useToast } from "@lib/hooks";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import ServicePage from "./-components/ServicePage";

async function fetchServiceData(id) {
  const response = await fetch(`/api/v1/merchants/services/${id}`, {
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

async function fetchServicePageFormOptions() {
  const response = await fetch("/api/v1/merchants/services/form-options", {
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

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/services/edit/$id"
)({
  component: RouteComponent,
  loader: async ({ params }) => {
    const service = await fetchServiceData(params.id);
    const formOptions = await fetchServicePageFormOptions();

    return {
      // crumb: "Edit service",
      service: service,
      products: formOptions?.products,
      categories: formOptions?.categories,
    };
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function RouteComponent() {
  const loaderData = Route.useLoaderData();
  const router = useRouter();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();

  async function saveServiceHandler(service) {
    try {
      const response = await fetch(`/api/v1/merchants/services/${service.id}`, {
        method: "PUT",
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
        showToast({
          message: "Service updated successfully",
          variant: "success",
        });
        setServerError();
        router.navigate({
          from: Route.fullPath,
          to: "/services",
        });
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  return (
    <>
      <ServerError error={serverError} />
      <ServicePage
        service={loaderData.service}
        categories={loaderData.categories}
        onSave={saveServiceHandler}
        route={Route}
      />
    </>
  );
}
