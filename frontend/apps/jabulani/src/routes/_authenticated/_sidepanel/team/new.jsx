import { ServerError } from "@reservations/components";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import TeamMemberPage from "./-components/EmployeePage";

export const Route = createFileRoute("/_authenticated/_sidepanel/team/new")({
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = Route.useNavigate();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();

  async function saveHandler(employee) {
    try {
      const response = await fetch("/api/v1/merchants/employees", {
        method: "POST",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify(employee),
      });

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        showToast({
          message: "Team member added successfully",
          variant: "success",
        });
        setServerError();
        navigate({
          from: Route.fullPath,
          to: "/team",
        });
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  return (
    <>
      <ServerError error={serverError} />
      <TeamMemberPage onSave={saveHandler} />
    </>
  );
}
