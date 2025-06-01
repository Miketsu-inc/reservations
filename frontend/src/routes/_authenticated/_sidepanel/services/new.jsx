import ServerError from "@components/ServerError";
import { useToast } from "@lib/hooks";
import { invalidateLocalSotrageAuth } from "@lib/lib";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import ServicePage from "./-components/ServicePage";

export const Route = createFileRoute("/_authenticated/_sidepanel/services/new")(
  {
    component: RouteComponent,
    // loader: () => ({ crumb: "new service" }),
  }
);

function RouteComponent() {
  const router = useRouter();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();

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
        invalidateLocalSotrageAuth(response.status);
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
      <ServicePage onSave={saveServiceHandler} />
    </>
  );
}
