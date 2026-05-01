import { Loading, ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import EmployeePage from "./-components/EmployeePage";

async function fetchEmployee(merchantId, id) {
  const response = await fetch(`/api/v1/merchants/${merchantId}/team/${id}`, {
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

export function employeeQueryOptions(merchantId, id) {
  return queryOptions({
    queryKey: [merchantId, "employee", id],
    queryFn: () => fetchEmployee(merchantId, id),
  });
}

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/team/edit/$id"
)({
  component: RouteComponent,
  loader: async ({
    context: {
      queryClient,
      authContext: { merchantId },
    },
    params,
  }) => {
    await queryClient.ensureQueryData(
      employeeQueryOptions(merchantId, params.id)
    );
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
  const { merchantId } = useAuth();

  const {
    data: employee,
    isLoading,
    isError,
    error,
  } = useQuery(employeeQueryOptions(merchantId, id));

  async function saveEmployee(employee) {
    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/team/${employee.id}`,
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
