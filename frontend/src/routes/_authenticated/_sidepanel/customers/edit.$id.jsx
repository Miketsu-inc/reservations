import Loading from "@components/Loading";
import ServerError from "@components/ServerError";
import { useToast } from "@lib/hooks";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import CustomerPage from "./-components/CustomerPage";

async function fetchCustomerData(id) {
  const response = await fetch(`/api/v1/merchants/customers/${id}`, {
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

function customerQueryOptions(id) {
  return queryOptions({
    queryKey: ["customer", id],
    queryFn: () => fetchCustomerData(id),
  });
}

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/customers/edit/$id"
)({
  component: RouteComponent,
  loader: ({ context: { queryClient }, params }) => {
    return queryClient.ensureQueryData(customerQueryOptions(params.id));
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function RouteComponent() {
  const router = useRouter();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();
  const { id } = Route.useParams();
  const { data, isLoading, isError, error } = useQuery(
    customerQueryOptions(id)
  );

  async function saveCustomerHandler(customer) {
    try {
      const response = await fetch(
        `/api/v1/merchants/customers/${customer.id}`,
        {
          method: "PUT",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify(customer),
        }
      );

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        showToast({
          message: "Customer modifyed successfully",
          variant: "success",
        });
        router.navigate({
          from: Route.fullPath,
          to: router.history.back(),
        });
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  return (
    <>
      <ServerError error={serverError} />
      <CustomerPage customer={data} onSave={saveCustomerHandler} />
    </>
  );
}
