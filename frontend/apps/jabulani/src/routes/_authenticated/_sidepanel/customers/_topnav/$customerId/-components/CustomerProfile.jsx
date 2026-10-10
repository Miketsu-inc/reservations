import {
  BirthdayCakeIcon,
  Call02Icon,
  CheckmarkCircle02Icon,
  Delete02Icon,
  Edit03Icon,
  Mail01Icon,
  MoreVerticalIcon,
  UnavailableIcon,
  UserAdd01Icon,
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
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import BlacklistModal from "../../-components/BlacklistModal";
import TransferAppsModal from "../../-components/TransferAppsModal";
import ExpandableNote from "./ExpandableNote";
import {
  customerProfileQueryOptions,
  customerStatsQueryOptions,
} from "./queries";

export default function CustomerProfile({ customerId }) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [showBlacklistModal, setShowBlacklistModal] = useState(false);
  const [showTransferModal, setShowTransferModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();
  const { merchantId } = useAuth();
  const profileQuery = useQuery(
    customerProfileQueryOptions(merchantId, customerId)
  );
  const { data: stats } = useQuery(
    customerStatsQueryOptions(merchantId, customerId)
  );

  if (profileQuery.isLoading) {
    return (
      <section>
        <h2 className="mb-4 text-xl">Profile</h2>
        <Loading />
      </section>
    );
  }

  if (profileQuery.isError) {
    return (
      <section>
        <h2 className="mb-4 text-xl">Profile</h2>
        <ServerError error={profileQuery.error?.message} />
      </section>
    );
  }

  const customer = profileQuery.data;
  const fullName =
    [customer.first_name, customer.last_name].filter(Boolean).join(" ") ||
    "Unnamed customer";
  const initials =
    `${customer.first_name?.charAt(0) ?? ""}${customer.last_name?.charAt(0) ?? ""}` ||
    "?";
  const hasContactDetails = Boolean(
    customer.email || customer.phone_number || customer.birthday
  );
  const hasBookings =
    (stats?.times_booked ?? 0) > 0 ||
    (stats?.times_confirmed ?? 0) > 0 ||
    (stats?.times_completed ?? 0) > 0 ||
    (stats?.times_cancelled ?? 0) > 0 ||
    (stats?.times_no_show ?? 0) > 0;

  async function deleteHandler() {
    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/customers/${customer.id}`,
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
      await queryClient.invalidateQueries(customersQueryOptions(merchantId));
      navigate({ to: "/customers" });
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
        queryClient.invalidateQueries(
          customerProfileQueryOptions(merchantId, customerId)
        ),
        queryClient.invalidateQueries(customersQueryOptions(merchantId)),
      ]);
      setServerError();
    } catch (error) {
      setServerError(error.message);
    }
  }

  return (
    <section>
      <TransferAppsModal
        fromCustomerId={customer.id}
        isOpen={showTransferModal}
        onClose={() => setShowTransferModal(false)}
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
        onDelete={deleteHandler}
      />

      <ServerError error={serverError} />
      <h2 className="mb-4 text-xl">Profile</h2>
      <Card styles="flex h-auto! flex-col gap-4 p-5! sm:p-6!">
        <div className="flex w-full justify-between gap-4">
          <div className="flex min-w-0 items-center gap-4">
            <Avatar styles="size-14! text-lg!" initials={initials} />
            <div className="flex min-w-0 flex-col gap-1">
              <h1 className="text-text_color truncate text-2xl font-bold">
                {fullName}
              </h1>
              <div className="flex flex-wrap items-center gap-2">
                {customer.is_blacklisted && (
                  <span
                    className="rounded-lg bg-red-400/30 p-1 text-sm text-red-700
                      dark:bg-red-700/30 dark:text-red-400"
                  >
                    Blacklisted
                  </span>
                )}
                {customer.is_dummy && (
                  <span
                    className="border-border_color bg-hvr_gray/50
                      text-text_color/70 inline-flex items-center gap-1.5
                      rounded-full border px-2 py-1 text-xs font-medium"
                  >
                    <Icon icon={UserAdd01Icon} styles="size-3.5 shrink-0" />
                    Added manually
                  </span>
                )}
              </div>
            </div>
          </div>
          <CustomerActions
            customer={customer}
            hasBookings={hasBookings}
            onBlacklist={() => setShowBlacklistModal(true)}
            onDelete={() => setShowDeleteModal(true)}
            onEdit={() => navigate({ to: `/customers/edit/${customer.id}` })}
            onTransfer={() => setShowTransferModal(true)}
          />
        </div>

        {hasContactDetails && (
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
        )}
        <ExpandableNote text={customer.note} />
      </Card>
    </section>
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
          className="hover:bg-hvr_gray flex size-9 cursor-pointer items-center
            justify-center rounded-lg"
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

function formatBirthday(dateString) {
  return new Date(dateString).toLocaleDateString("en-US", {
    month: "long",
    day: "numeric",
  });
}
