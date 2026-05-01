import { Loading, ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  invalidateLocalStorageAuth,
  serviceFormOptionsQueryOptions,
  useToast,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import ServicePage from "./-components/ServicePage";

export const Route = createFileRoute("/_authenticated/_sidepanel/services/new")(
  {
    component: RouteComponent,
    loader: async ({
      context: {
        queryClient,
        useContext: { merchantId },
      },
    }) => {
      await queryClient.ensureQueryData(
        serviceFormOptionsQueryOptions(merchantId)
      );
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
  const { merchantId } = useAuth();

  const {
    data: formOptions,
    isLoading,
    isError,
    error,
  } = useQuery(serviceFormOptionsQueryOptions(merchantId));

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error} />;
  }

  async function saveServiceHandler(service) {
    try {
      const response = await fetch(`/api/v1/merchants/${merchantId}/services`, {
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
