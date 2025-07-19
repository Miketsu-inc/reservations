import ServerError from "@components/ServerError";
import { useToast } from "@lib/hooks";
import { invalidateLocalStorageAuth } from "@lib/lib";
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

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/customers/edit/$id"
)({
  component: RouteComponent,
  loader: async ({ params }) => {
    const customer = await fetchCustomerData(params.id);

    return {
      crumb: "Edit Customer",
      customer: customer,
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

  return (
    <>
      <ServerError error={serverError} />
      <CustomerPage
        customer={loaderData.customer}
        onSave={saveCustomerHandler}
      />
    </>
  );
}
