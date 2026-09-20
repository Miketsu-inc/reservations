import { PlusSignIcon } from "@hugeicons/core-free-icons";
import { Button, Icon, Loading, ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  customersQueryOptions,
  invalidateLocalStorageAuth,
  useToast,
  useWindowSize,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import BlacklistModal from "./-components/BlacklistModal";
import CustomersTable from "./-components/CustomersTable";
import TransferAppsModal from "./-components/TransferAppsModal";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/customers/_topnav/"
)({
  component: CustomersPage,
  loader: async ({
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => {
    await queryClient.ensureQueryData(customersQueryOptions(merchantId));
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function CustomersPage() {
  const navigate = Route.useNavigate();
  const [showTransferModal, setShowTransferModal] = useState(false);
  const [transferCustomerId, setTransferCustomerId] = useState();
  const [showBlacklistModal, setShowBlacklistModal] = useState(false);
  const [blacklistModalData, setBlacklistModalData] = useState();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();
  const { merchantId } = useAuth();
  const { windowSize } = useWindowSize();

  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { data, isLoading, isError, error } = useQuery(
    customersQueryOptions(merchantId)
  );

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
        `/api/v1/merchants/${merchantId}/customers/${selected.id}`,
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
        await queryClient.invalidateQueries(customersQueryOptions(merchantId));
        setServerError();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  async function blacklistHandler(data) {
    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/customers/${data.id}/blacklist`,
        {
          method: "PUT",
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
        await queryClient.invalidateQueries(customersQueryOptions(merchantId));
        setServerError();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  return (
    <div className="h-full">
      <div className="flex flex-row items-center justify-between">
        <p className="pb-6 text-xl">Customers</p>
        <Link from={Route.fullPath} to="new">
          <Button
            variant="primary"
            styles="p-2 md:px-4 w-fit"
            buttonText={windowSize !== "sm" ? "New Customer" : ""}
          >
            <Icon
              icon={PlusSignIcon}
              styles="size-6 md:size-5 md:mr-2 text-white"
            />
          </Button>
        </Link>
      </div>
      <div className="flex h-full min-h-0 justify-center">
        <TransferAppsModal
          fromCustomerId={transferCustomerId}
          isOpen={showTransferModal}
          onClose={() => {
            setShowTransferModal(false);
            setTransferCustomerId();
          }}
        />
        <BlacklistModal
          key={blacklistModalData?.id || "new"}
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
        <div className="flex min-h-0 w-full flex-1 flex-col gap-5">
          <ServerError error={serverError} />
          <div className="flex min-h-0 w-full flex-1">
            <CustomersTable
              customersData={data}
              onTransfer={(customerId) => {
                setTransferCustomerId(customerId);
                setTimeout(() => setShowTransferModal(true), 0);
              }}
              onEdit={(customer) => {
                navigate({
                  from: Route.fullPath,
                  to: `edit/${customer.id}`,
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
    </div>
  );
}
