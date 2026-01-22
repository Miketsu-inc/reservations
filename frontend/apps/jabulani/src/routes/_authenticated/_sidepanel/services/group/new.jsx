import { Loading, ServerError } from "@reservations/components";
import {
  invalidateLocalStorageAuth,
  serviceFormOptionsQueryOptions,
  useToast,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import GroupServicePage from "./-components/GroupServicePage";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/services/group/new"
)({
  component: RouteComponent,
  loader: async ({ context: { queryClient } }) => {
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

  async function saveHandler(service) {
    try {
      const response = await fetch("/api/v1/merchants/group-services", {
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
          message: "Group service added successfully",
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
      <GroupServicePage
        categories={formOptions.categories}
        products={formOptions.products}
        onSave={saveHandler}
        route={Route}
      />
    </>
  );
}
