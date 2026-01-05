import { Loading, ServerError } from "@reservations/components";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import EmployeePage from "./-components/EmployeePage";

async function fetchEmployee(id) {
  const response = await fetch(`/api/v1/merchants/employees/${id}`, {
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

export function employeeQueryOptions(id) {
  return queryOptions({
    queryKey: ["employee", id],
    queryFn: () => fetchEmployee(id),
  });
}

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/team/edit/$id"
)({
  component: RouteComponent,
  loader: async ({ context: { queryClient }, params }) => {
    await queryClient.ensureQueryData(employeeQueryOptions(params.id));
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function RouteComponent() {
  const { id } = Route.useParams({ from: Route.id });
  const router = useRouter();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();

  const {
    data: employee,
    isLoading,
    isError,
    error,
  } = useQuery(employeeQueryOptions(id));

  async function saveEmployee(employee) {
    try {
      const response = await fetch(
        `/api/v1/merchants/employees/${employee.id}`,
        {
          method: "PUT",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify(employee),
        }
      );

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        showToast({
          message: "Employee modified successfully",
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
      <EmployeePage employee={employee} onSave={saveEmployee} />
    </>
  );
}
