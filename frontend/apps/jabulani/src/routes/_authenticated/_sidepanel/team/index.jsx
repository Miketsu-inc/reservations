import { Loading, ServerError } from "@reservations/components";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import EmployeeTable from "./-components/EmployeeTable";

async function fetchEmployees() {
  const response = await fetch("/api/v1/merchants/employees", {
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

function employeesQueryOptions() {
  return queryOptions({
    queryKey: ["employees"],
    queryFn: fetchEmployees,
  });
}

export const Route = createFileRoute("/_authenticated/_sidepanel/team/")({
  component: RouteComponent,
  loader: async ({ context: { queryClient } }) => {
    await queryClient.ensureQueryData(employeesQueryOptions());
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function RouteComponent() {
  const router = useRouter();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();

  const { queryClient } = Route.useRouteContext({ from: Route.id });

  const {
    data: employees,
    isLoading,
    isError,
    error,
  } = useQuery(employeesQueryOptions());

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  function handleRowClick(e) {
    const employeeId = e.data.id;
    const target = e.event.target;
    const colId = target.closest("[col-id]")?.getAttribute("col-id");

    if (colId === "actions") {
      return;
    }

    router.navigate({
      from: Route.fullPath,
      to: `${employeeId}`,
    });
  }

  async function deleteHandler(employee) {
    try {
      const response = await fetch(
        `/api/v1/merchants/employees/${employee.id}`,
        {
          method: "DELETE",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
        }
      );

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        showToast({
          message: "Employee deleted successfully",
          variant: "success",
        });
        await queryClient.invalidateQueries(employeesQueryOptions());
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  return (
    <div className="h-screen px-4 md:px-0">
      <p className="py-6 text-xl font-semibold">Team members</p>
      <ServerError error={serverError} />
      <div className="h-2/3">
        <EmployeeTable
          data={employees}
          onRowClick={handleRowClick}
          oneNewItem={() =>
            router.navigate({ from: Route.fullPath, to: "new" })
          }
          onDelete={deleteHandler}
          onEdit={(employee) =>
            router.navigate({ from: Route.fullPath, to: `edit/${employee.id}` })
          }
        />
      </div>
    </div>
  );
}
