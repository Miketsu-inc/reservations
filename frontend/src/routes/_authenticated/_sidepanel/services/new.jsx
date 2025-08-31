import Loading from "@components/Loading";
import ServerError from "@components/ServerError";
import { useToast } from "@lib/hooks";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import ServicePage from "./-components/ServicePage";

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

export const Route = createFileRoute("/_authenticated/_sidepanel/services/new")(
  {
    component: RouteComponent,
    loader: async ({ context: { queryClient } }) => {
      await queryClient.ensureQueryData(serviceFormOptionsQueryOptions());
    },
    errorComponent: ({ error }) => {
      return <ServerError error={error.message} />;
    },
  }
);

function RouteComponent() {
  const router = useRouter();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();

  const {
    data: formOptions,
    isLoading,
    isError,
    error,
  } = useQuery(serviceFormOptionsQueryOptions());

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error} />;
  }

  async function saveServiceHandler(service) {
    try {
      const response = await fetch("/api/v1/merchants/services", {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify(service),
      });

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        showToast({
          message: "Service added successfully",
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
        categories={formOptions.categories}
        products={formOptions.products}
        onSave={saveServiceHandler}
        route={Route}
      />
    </>
  );
}
