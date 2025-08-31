import Loading from "@components/Loading";
import ServerError from "@components/ServerError";
import { useToast } from "@lib/hooks";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { queryOptions, useQueries } from "@tanstack/react-query";
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
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

function serviceQueryOptions(id) {
  return queryOptions({
    queryKey: ["service", id],
    queryFn: () => fetchServiceData(id),
  });
}

async function fetchServiceFormOptions() {
  const response = await fetch("/api/v1/merchants/services/form-options", {
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

function serviceFormOptionsQueryOptions() {
  return queryOptions({
    queryKey: ["service-from-options"],
    queryFn: fetchServiceFormOptions,
  });
}

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/services/edit/$id"
)({
  component: RouteComponent,
  loader: async ({ params, context: { queryClient } }) => {
    await queryClient.ensureQueryData(serviceQueryOptions(params.id));
    await queryClient.ensureQueryData(serviceFormOptionsQueryOptions());
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function RouteComponent() {
  const router = useRouter();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();

  const { id } = Route.useParams({ from: Route.id });

  const queryResults = useQueries({
    queries: [serviceQueryOptions(id), serviceFormOptionsQueryOptions()],
  });

  if (queryResults.some((r) => r.isLoading)) {
    return <Loading />;
  }

  if (queryResults.some((r) => r.isError)) {
    const error = queryResults.find((r) => r.error);
    return <ServerError error={error} />;
  }

  async function saveServiceHandler(service) {
    try {
      const { used_products, ...serviceData } = service;

      await updateServiceData(service.id, serviceData);
      await updateUsedProducts(service.id, used_products);

      showToast({
        message: "Service updated successfully",
        variant: "success",
      });
      setServerError();
      router.navigate({
        from: Route.fullPath,
        to: "/services",
      });
    } catch (err) {
      setServerError(err.message);
    }
  }

  async function updateServiceData(serviceId, serviceData) {
    const response = await fetch(`/api/v1/merchants/services/${serviceId}`, {
      method: "PUT",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify(serviceData),
    });

    if (!response.ok) {
      invalidateLocalStorageAuth(response.status);
      const result = await response.json();
      throw new Error(result.error.message);
    }
  }

  async function updateUsedProducts(serviceId, usedProducts) {
    const response = await fetch(
      `/api/v1/merchants/services/${serviceId}/products`,
      {
        method: "PUT",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          service_id: serviceId,
          used_products: usedProducts,
        }),
      }
    );

    if (!response.ok) {
      invalidateLocalStorageAuth(response.status);
      const result = await response.json();
      throw new Error(result.error.message);
    }
  }

  return (
    <>
      <ServerError error={serverError} />
      <ServicePage
        service={queryResults[0].data}
        categories={queryResults[1].data.categories}
        products={queryResults[1].data.products}
        onSave={saveServiceHandler}
        route={Route}
      />
    </>
  );
}
