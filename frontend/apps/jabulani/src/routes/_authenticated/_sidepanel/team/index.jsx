import { Loading, ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import EmployeeTable from "./-components/EmployeeTable";

async function fetchEmployees(merchantId) {
  const response = await fetch(`/api/v1/merchants/${merchantId}/team`, {
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

function employeesQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "employees"],
    queryFn: () => fetchEmployees(merchantId),
  });
}

export const Route = createFileRoute("/_authenticated/_sidepanel/team/")({
  component: RouteComponent,
  loader: async ({
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => {
    await queryClient.ensureQueryData(employeesQueryOptions(merchantId));
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
  const { merchantId } = useAuth();

  const {
    data: employees,
    isLoading,
    isError,
    error,
  } = useQuery(employeesQueryOptions(merchantId));

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
        `/api/v1/merchants/${merchantId}/team/${employee.id}`,
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
        await queryClient.invalidateQueries({
          queryKey: [merchantId, "employees"],
        });
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  return (
    <div className="h-screen px-4 py-2 md:px-0 md:py-0">
      <p className="pb-6 text-xl">Team members</p>
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
