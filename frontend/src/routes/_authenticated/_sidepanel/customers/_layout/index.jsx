import Loading from "@components/Loading";
import ServerError from "@components/ServerError";
import { useToast } from "@lib/hooks";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import BlacklistModal from "../-components/BlacklistModal";
import CustomersTable from "../-components/CustomersTable";
import TransferAppsModal from "../-components/TransferAppsModal";

async function fetchCustomers() {
  const response = await fetch(`/api/v1/merchants/customers`, {
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

function customersQueryOptions() {
  return queryOptions({
    queryKey: ["customers"],
    queryFn: fetchCustomers,
  });
}

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/customers/_layout/"
)({
  component: CustomersPage,
  loader: ({ context: { queryClient } }) => {
    return queryClient.ensureQueryData(customersQueryOptions());
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function CustomersPage() {
  const navigate = Route.useNavigate();
  const [showTransferModal, setShowTransferModal] = useState(false);
  const [transferModalData, setTransferModalData] = useState();
  const [showBlacklistModal, setShowBlacklistModal] = useState(false);
  const [blacklistModalData, setBlacklistModalData] = useState();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();

  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { data, isLoading, isError, error } = useQuery(customersQueryOptions());

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  function handleRowClick(e) {
    const customerId = e.data.id;
    const target = e.event.target;
    const colId = target.closest("[col-id]")?.getAttribute("col-id");

    if (colId === "actions") {
      return;
    }

    navigate({
      from: Route.fullPath,
      to: `${customerId}`,
    });
  }

  async function deleteHandler(selected) {
    try {
      const response = await fetch(
        `/api/v1/merchants/customers/${selected.id}`,
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
          message: "Customer deleted successfully",
          variant: "success",
        });
        await queryClient.invalidateQueries({
          queryKey: ["customers"],
        });
        setServerError();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  async function transferHandler(data) {
    try {
      const response = await fetch(
        `/api/v1/merchants/customers/transfer?from=${data.from}&to=${data.to}`,
        {
          method: "PUT",
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
          message: "Appointments transferred successfully",
          variant: "success",
        });
        await queryClient.invalidateQueries({
          queryKey: ["customers"],
        });
        setServerError();
        setTransferModalData();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  async function blacklistHandler(data) {
    try {
      const response = await fetch(
        `/api/v1/merchants/customers/blacklist/${data.id}`,
        {
          method: "POST",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify({
            id: data.id,
            blacklist_reason: data.blacklist_reason,
          }),
        }
      );

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        showToast({
          message: "Customer blacklisted successfully",
          variant: "success",
        });
        await queryClient.invalidateQueries({
          queryKey: ["customers"],
        });
        setServerError();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  return (
    <div className="flex h-screen justify-center">
      <TransferAppsModal
        data={transferModalData}
        isOpen={showTransferModal}
        onClose={() => setShowTransferModal(false)}
        onSubmit={transferHandler}
      />
      <BlacklistModal
        data={blacklistModalData}
        isOpen={showBlacklistModal}
        onClose={() => setShowBlacklistModal(false)}
        // both adding to and removing from blacklist goes through the same modal and handler
        // so the customer.is_blacklisted field determines the action
        onSubmit={(customer) =>
          blacklistHandler({
            id: customer.id,
            blacklist_reason: customer.blacklist_reason,
          })
        }
      />
      <div className="flex w-full flex-col gap-5">
        <ServerError error={serverError} />
        <div className="h-2/3 w-full">
          <CustomersTable
            customersData={data}
            onTransfer={(index) => {
              setTransferModalData({
                fromIndex: index,
                customers: data,
              });
              setTimeout(() => setShowTransferModal(true), 0);
            }}
            onEdit={(customer) => {
              navigate({
                from: Route.fullPath,
                to: `/customers/edit/${customer.id}`,
              });
            }}
            onDelete={deleteHandler}
            onBlackList={(customer) => {
              setBlacklistModalData(customer);
              setTimeout(() => setShowBlacklistModal(true), 0);
            }}
            onRowClick={handleRowClick}
          />
        </div>
      </div>
    </div>
  );
}
