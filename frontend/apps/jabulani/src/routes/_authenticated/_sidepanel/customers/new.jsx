import { ServerError } from "@reservations/components";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import CustomerPage from "./-components/CustomerPage";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/customers/new"
)({
  component: RouteComponent,
});

function RouteComponent() {
  const router = useRouter();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();

  async function saveCustomerHandler(customer) {
    try {
      const response = await fetch("/api/v1/merchants/customers", {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify(customer),
      });

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        showToast({
          message: "Customer added successfully",
          variant: "success",
        });
        setServerError();
        router.navigate({
          from: Route.fullPath,
          to: "/customers",
        });
      }
    } catch (err) {
      setServerError(err.message);
    }
  }
  return (
    <>
      <ServerError error={serverError} />
      <CustomerPage onSave={saveCustomerHandler} />
    </>
  );
}
