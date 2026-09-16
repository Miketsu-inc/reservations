import {
  BirthdayCakeIcon,
  Call02Icon,
  CheckmarkCircle02Icon,
  Delete02Icon,
  Edit03Icon,
  Mail01Icon,
  MoreVerticalIcon,
  UnavailableIcon,
  UserSwitchIcon,
} from "@hugeicons/core-free-icons";
import {
  Avatar,
  Card,
  DeleteModal,
  Icon,
  Loading,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
  ServerError,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import {
  customersQueryOptions,
  invalidateLocalStorageAuth,
  useToast,
} from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import BlacklistModal from "../-components/BlacklistModal";
import TransferAppsModal from "../-components/TransferAppsModal";
import BookingHistory from "./-components/BookingHistory";
import CustomerStats from "./-components/CustomerStats";
import ExpandableNote from "./-components/ExpandableNote";

async function fetchCustomerResource(merchantId, customerId, resource = "") {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/customers/${customerId}${resource}`,
    {
      method: "GET",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    }
  );

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  }

  return result.data;
}

function customerProfileQueryOptions(merchantId, customerId) {
  return queryOptions({
    queryKey: [merchantId, "customer-profile", customerId],
    queryFn: () => fetchCustomerResource(merchantId, customerId),
  });
}

function customerStatsQueryOptions(merchantId, customerId) {
  return queryOptions({
    queryKey: [merchantId, "customer-stats", customerId],
    queryFn: () => fetchCustomerResource(merchantId, customerId, "/stats"),
  });
}

function formatBirthday(dateString) {
  return new Date(dateString).toLocaleDateString("en-US", {
    month: "long",
    day: "numeric",
  });
}

function formatVisitDate(dateString) {
  if (!dateString) return null;

  return new Date(dateString).toLocaleDateString([], {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/customers/_topnav/$customerId/"
)({
  validateSearch: (search) => ({
    status: ["upcoming", "completed", "cancelled"].includes(search.status)
      ? search.status
      : "upcoming",
  }),
  component: CustomerDetailsPage,
  loader: async ({
    params,
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => {
    await queryClient.ensureQueryData(
      customerProfileQueryOptions(merchantId, params.customerId)
    );
  },
  pendingComponent: Loading,
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});

function CustomerDetailsPage() {
  const navigate = Route.useNavigate();
  const { status } = Route.useSearch();
  const [showBlacklistModal, setShowBlacklistModal] = useState(false);
  const [showTransferModal, setShowTransferModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();
  const { merchantId } = useAuth();
  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { customerId } = Route.useParams();

  const profileQuery = useQuery(
    customerProfileQueryOptions(merchantId, customerId)
  );
  const statsQuery = useQuery(
    customerStatsQueryOptions(merchantId, customerId)
  );
  const customersQuery = useQuery({
    ...customersQueryOptions(merchantId),
    enabled: showTransferModal,
  });

  if (profileQuery.isLoading) return <Loading />;
  if (profileQuery.isError) {
    return <ServerError error={profileQuery.error?.message} />;
  }

  const customer = profileQuery.data;
  const stats = statsQuery.data;
  const fullName =
    [customer.first_name, customer.last_name].filter(Boolean).join(" ") ||
    "Unnamed customer";
  const initials =
    `${customer.first_name?.charAt(0) ?? ""}${customer.last_name?.charAt(0) ?? ""}` ||
    "?";
  const lastVisited = formatVisitDate(stats?.last_visited);

  async function deleteHandler(id) {
    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/customers/${id}`,
        { method: "DELETE" }
      );

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
        return;
      }

      showToast({
        message: "Customer deleted successfully",
        variant: "success",
      });
      await queryClient.invalidateQueries({
        queryKey: [merchantId, "customers"],
      });
      navigate({ from: Route.fullPath, to: "/customers" });
    } catch (error) {
      setServerError(error.message);
    }
  }

  async function blacklistHandler(data) {
    const options = {
      method: data.method,
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    };
    if (data.method === "PUT") {
      options.body = JSON.stringify({
        id: data.id,
        blacklist_reason: data.blacklistReason,
      });
    }

    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/customers/${data.id}/blacklist`,
        options
      );
      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
        return;
      }

      showToast({
        message:
          data.method === "PUT"
            ? "Customer blacklisted successfully"
            : "Customer removed from blacklist successfully",
        variant: "success",
      });
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: [merchantId, "customer-profile", customerId],
        }),
        queryClient.invalidateQueries({
          queryKey: [merchantId, "customers"],
        }),
      ]);
      setServerError();
    } catch (error) {
      setServerError(error.message);
    }
  }

  async function transferHandler(data) {
    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/customers/transfer`,
        {
          method: "PUT",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify({
            from_customer_id: data.from,
            to_customer_id: data.to,
          }),
        }
      );

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
        return;
      }

      showToast({
        message: "Bookings transferred successfully",
        variant: "success",
      });
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: [merchantId, "customer-stats", customerId],
        }),
        queryClient.invalidateQueries({
          queryKey: [merchantId, "customer-bookings", customerId],
        }),
        queryClient.invalidateQueries({
          queryKey: [merchantId, "customers"],
        }),
      ]);
      setServerError();
    } catch (error) {
      setServerError(error.message);
    }
  }

  function statusChangeHandler(nextStatus) {
    navigate({
      from: Route.fullPath,
      search: (previous) => ({ ...previous, status: nextStatus }),
      replace: true,
    });
  }

  return (
    <main className="flex justify-center">
      <TransferAppsModal
        data={{
          from: customer.id,
          customers: customersQuery.data ?? [customer],
        }}
        isOpen={showTransferModal}
        onClose={() => setShowTransferModal(false)}
        onSubmit={transferHandler}
      />
      <BlacklistModal
        key={customer.id}
        data={customer}
        isOpen={showBlacklistModal}
        onClose={() => setShowBlacklistModal(false)}
        onSubmit={(selectedCustomer) =>
          blacklistHandler({
            method: selectedCustomer.is_blacklisted ? "DELETE" : "PUT",
            id: selectedCustomer.id,
            blacklistReason: selectedCustomer.blacklist_reason,
          })
        }
      />
      <DeleteModal
        itemName={fullName}
        open={showDeleteModal}
        onOpenChange={setShowDeleteModal}
        onDelete={() => deleteHandler(customer.id)}
      />

      <div className="flex w-full max-w-xl flex-col gap-8 px-3 py-4 sm:px-0">
        <ServerError error={serverError} />

        <Card styles="flex flex-col items-start gap-4">
          <div className="flex w-full justify-between gap-4">
            <div className="flex min-w-0 items-center gap-4">
              <Avatar initials={initials} />
              <div className="flex min-w-0 flex-col gap-1">
                <h1 className="text-text_color truncate text-lg font-bold">
                  {fullName}
                </h1>
                <div className="flex flex-wrap items-center gap-2">
                  {customer.is_blacklisted && (
                    <span
                      className="rounded-lg bg-red-400/30 p-1 text-sm
                        text-red-700 dark:bg-red-700/30 dark:text-red-400"
                    >
                      Blacklisted
                    </span>
                  )}
                  {customer.is_dummy && (
                    <span
                      className="bg-hvr_gray text-text_color/70 rounded-lg p-1
                        text-sm"
                    >
                      Added manually
                    </span>
                  )}
                </div>
                <p className="text-text_color/60 text-sm">
                  {lastVisited
                    ? `Last visited: ${lastVisited}`
                    : "No completed visits yet"}
                </p>
              </div>
            </div>
            <CustomerActions
              customer={customer}
              hasBookings={(stats?.times_booked ?? 0) > 0}
              onBlacklist={() => setShowBlacklistModal(true)}
              onDelete={() => setShowDeleteModal(true)}
              onEdit={() =>
                navigate({
                  from: Route.fullPath,
                  to: `/customers/edit/${customer.id}`,
                })
              }
              onTransfer={() => setShowTransferModal(true)}
            />
          </div>

          <div
            className="text-text_color/70 flex w-full flex-col items-start gap-3
              text-sm sm:flex-row sm:flex-wrap sm:items-center sm:gap-6"
          >
            {customer.email && (
              <a
                className="flex min-w-0 items-center gap-2 hover:underline"
                href={`mailto:${customer.email}`}
              >
                <Icon icon={Mail01Icon} styles="size-5 shrink-0" />
                <span className="truncate">{customer.email}</span>
              </a>
            )}
            {customer.phone_number && (
              <a
                className="flex items-center gap-2 hover:underline"
                href={`tel:${customer.phone_number}`}
              >
                <Icon icon={Call02Icon} styles="size-4 shrink-0" />
                {customer.phone_number}
              </a>
            )}
            {customer.birthday && (
              <div className="flex items-center gap-2">
                <Icon icon={BirthdayCakeIcon} styles="size-5 shrink-0" />
                {formatBirthday(customer.birthday)}
              </div>
            )}
          </div>
          <ExpandableNote text={customer.note} />
        </Card>

        <CustomerStats
          stats={stats}
          isLoading={statsQuery.isLoading}
          error={statsQuery.error}
        />

        <BookingHistory
          customerId={customerId}
          merchantId={merchantId}
          status={status}
          counts={stats}
          onStatusChange={statusChangeHandler}
          route={Route}
        />
      </div>
    </main>
  );
}

function CustomerActions({
  customer,
  hasBookings,
  onBlacklist,
  onDelete,
  onEdit,
  onTransfer,
}) {
  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          aria-label="Customer actions"
          className="hover:bg-hvr_gray cursor-pointer rounded-lg p-2"
          type="button"
        >
          <Icon
            icon={MoreVerticalIcon}
            styles="size-5 rotate-90 text-text_color/60"
          />
        </button>
      </PopoverTrigger>
      <PopoverContent side="left" styles="w-auto">
        <div
          className="flex flex-col *:flex *:items-center *:gap-3 *:rounded-lg
            *:p-2"
        >
          {!customer.is_dummy && (
            <PopoverClose asChild>
              <button
                className="hover:bg-hvr_gray cursor-pointer"
                onClick={onBlacklist}
                type="button"
              >
                <Icon
                  icon={
                    customer.is_blacklisted
                      ? CheckmarkCircle02Icon
                      : UnavailableIcon
                  }
                  styles="size-5"
                />
                {customer.is_blacklisted
                  ? "Remove from blacklist"
                  : "Blacklist customer"}
              </button>
            </PopoverClose>
          )}
          <PopoverClose asChild>
            <button
              className="hover:bg-hvr_gray cursor-pointer"
              onClick={onEdit}
              type="button"
            >
              <Icon icon={Edit03Icon} styles="size-5" />
              Edit customer
            </button>
          </PopoverClose>
          {customer.is_dummy && hasBookings && (
            <PopoverClose asChild>
              <button
                className="hover:bg-hvr_gray cursor-pointer"
                onClick={onTransfer}
                type="button"
              >
                <Icon icon={UserSwitchIcon} styles="size-5" />
                Transfer bookings
              </button>
            </PopoverClose>
          )}
          {customer.is_dummy && (
            <PopoverClose asChild>
              <button
                className="hover:bg-hvr_gray cursor-pointer text-red-600
                  dark:text-red-500"
                onClick={onDelete}
                type="button"
              >
                <Icon icon={Delete02Icon} styles="size-5" />
                Delete customer
              </button>
            </PopoverClose>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}
