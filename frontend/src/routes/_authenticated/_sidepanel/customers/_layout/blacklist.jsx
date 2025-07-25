import ServerError from "@components/ServerError";
import PersonIcon from "@icons/PersonIcon";
import { useToast } from "@lib/hooks";
import { invalidateLocalStorageAuth } from "@lib/lib";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import BlacklistModal from "../-components/BlacklistModal";
import CustomersTable from "../-components/CustomersTable";

async function fetchCustomers() {
  const response = await fetch(`/api/v1/merchants/customers/blacklist`, {
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
  "/_authenticated/_sidepanel/customers/_layout/blacklist"
)({
  component: BlacklistPage,
  loader: async () => {
    const customers = await fetchCustomers();
    return {
      crumb: "Blacklisted",
      customers: customers,
    };
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function BlacklistPage() {
  const router = useRouter();
  const loaderData = Route.useLoaderData();
  const navigate = Route.useNavigate();
  const [showBlacklistModal, setShowBlacklistModal] = useState(false);
  const [blacklistModalData, setBlacklistModalData] = useState();
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();

  function handleRowClick(e) {
    const customerId = e.data.id;
    const target = e.event.target;
    const colId = target.closest("[col-id]")?.getAttribute("col-id");

    if (colId === "actions") {
      return;
    }

    navigate({
      from: Route.fullPath,
      to: `/customers/${customerId}`,
    });
  }

  async function blacklistHandler(id) {
    try {
      const response = await fetch(
        `/api/v1/merchants/customers/blacklist/${id}`,
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
          message: "Customer removed from blacklist successfully",
          variant: "success",
        });
        router.invalidate();
        setServerError();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  return (
    <div className="flex h-screen justify-center">
      <BlacklistModal
        data={blacklistModalData}
        isOpen={showBlacklistModal}
        onClose={() => setShowBlacklistModal(false)}
        onSubmit={(customer) => {
          blacklistHandler(customer.id);
        }}
      />
      <div className="flex w-full gap-5">
        <ServerError error={serverError} />
        <div className="h-2/3 w-full">
          <CustomersTable
            customersData={loaderData.customers}
            onBlackList={(customer) => {
              setBlacklistModalData(customer);
              setTimeout(() => setShowBlacklistModal(true), 0);
            }}
            onEdit={(customer) => {
              router.navigate({
                from: Route.fullPath,
                to: `/customers/edit/${customer.id}`,
              });
            }}
            onRowClick={handleRowClick}
            noRowsOverlayComponent={NoRowsComponent}
          />
        </div>
      </div>
    </div>
  );
}

function NoRowsComponent() {
  return (
    <div className="mb-16 flex flex-col items-center gap-4 px-2 text-gray-600 dark:text-gray-300">
      <PersonIcon styles="fill-current size-16" />
      <div className="flex flex-col items-center gap-2">
        <p className="text-text_color text-base font-medium">
          No blacklisted customers
        </p>
        <p className="text-sm md:w-2/3">
          You can blacklist clients who violate your policies. Blacklisted
          customers will not be able to book appointments through your booking
          page.
        </p>
      </div>
    </div>
  );
}
