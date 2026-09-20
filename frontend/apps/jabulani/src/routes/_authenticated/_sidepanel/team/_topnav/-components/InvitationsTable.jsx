import { MailRemoveIcon, RecoveryMailIcon } from "@hugeicons/core-free-icons";
import { Icon, Loading } from "@reservations/components";
import { useWindowSize } from "@reservations/lib";
import { lazy, Suspense } from "react";

const Table = lazy(() => import("@reservations/components/Table"));

const GRAY_STYLES =
  "bg-gray-600/20 text-gray-600 dark:bg-gray-500/15 dark:text-gray-400";

const INVITATIONS_STYLES = {
  pending:
    "bg-amber-600/20 text-amber-600 dark:bg-amber-600/15 dark:text-amber-400",
  accepted:
    "bg-green-600/20 text-green-600 dark:bg-green-500/15 dark:text-green-400",
  declined: "bg-red-600/20 text-red-600 dark:bg-red-500/15 dark:text-red-400",
  revoked: GRAY_STYLES,
  expired: GRAY_STYLES,
};

function getStatusSince(data) {
  switch (data.status) {
    case "pending":
      return data.invited_at;
    case "accepted":
      return data.accepted_at;
    case "declined":
      return data.declined_at;
    case "revoked":
      return data.revoked_at;
    case "expired":
      return data.expires_at;
    default:
      return null;
  }
}

export default function InvitationsTable({
  data,
  onNewItem,
  onRevoke,
  onResend,
}) {
  const { isWindowSmall } = useWindowSize();

  const columnDef = [
    { field: "id", hide: true },
    {
      field: "email",
      headerName: "Email",
      flex: 1,
      resizable: false,
      ...(isWindowSmall ? { minWidth: 180 } : {}),
      cellRenderer: (params) => {
        return (
          <div className="flex h-full flex-col justify-center leading-5">
            <p className="">{params.data.email}</p>
            <p className="text-text_color/60 text-sm">{params.data.role}</p>
          </div>
        );
      },
    },
    {
      field: "status",
      headerName: "Status",
      flex: 1,
      resizable: false,
      ...(isWindowSmall ? { minWidth: 80 } : {}),
      cellRenderer: (params) => {
        return (
          <div className="flex h-full items-center">
            <span
              className={`${INVITATIONS_STYLES[params.data.status]} w-fit
                rounded-full px-2 py-1 text-sm`}
            >
              <p>{params.data.status}</p>
            </span>
          </div>
        );
      },
    },
    {
      field: "status_since",
      headerName: "Status since",
      flex: 1,
      resizable: false,
      hide: isWindowSmall,
      ...(isWindowSmall ? { minWidth: 140 } : {}),
      valueGetter: (params) => {
        const date = getStatusSince(params.data);
        return date ? new Date(date) : null;
      },
      valueFormatter: (params) => {
        if (!params.value) return "";
        return new Intl.DateTimeFormat(undefined, {
          dateStyle: "medium",
          timeStyle: "short",
        }).format(params.value);
      },
      cellStyle: {
        display: "flex",
        alignItems: "center",
      },
    },
    {
      field: "actions",
      flex: 1,
      headerName: "",
      cellRenderer: (params) => {
        const canResend =
          params.data.status === "expired" || params.data.status === "revoked";
        const canRevoke = params.data.status === "pending";

        return (
          <div
            key={params.data.id}
            className="flex h-full flex-row items-center justify-center"
          >
            {canResend && (
              <button
                className="hover:bg-hvr_gray flex cursor-pointer flex-row
                  items-center gap-2 rounded-lg px-2"
                onClick={() => onResend(data[params.node.sourceRowIndex])}
              >
                <Icon icon={RecoveryMailIcon} styles="size-6" />
                <p>Resend</p>
              </button>
            )}
            {canRevoke && (
              <button
                className="hover:bg-hvr_gray flex cursor-pointer flex-row
                  items-center gap-2 rounded-lg px-2"
                onClick={() => onRevoke(data[params.node.sourceRowIndex])}
              >
                <Icon icon={MailRemoveIcon} styles="size-6" />
                <p>Revoke</p>
              </button>
            )}
          </div>
        );
      },
      resizable: false,
      sortable: false,
      minWidth: 160,
      maxWidth: 160,
    },
  ];

  return (
    <div className="flex h-full min-h-0 w-full min-w-0 flex-col">
      <Suspense fallback={<Loading />}>
        <Table
          rowData={data}
          columnDef={columnDef}
          itemName="invitation"
          onNewItem={onNewItem}
          exportName="invitations"
          rowHeight={70}
        />
      </Suspense>
    </div>
  );
}
