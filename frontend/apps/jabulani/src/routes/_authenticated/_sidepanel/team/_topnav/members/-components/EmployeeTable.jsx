import { Delete02Icon, Edit03Icon } from "@hugeicons/core-free-icons";
import { Avatar, DeleteModal, Icon, Loading } from "@reservations/components";
import { useWindowSize } from "@reservations/lib";
import { lazy, Suspense, useState } from "react";

const Table = lazy(() => import("@reservations/components/Table"));

export default function EmployeeTable({
  data,
  onRowClick,
  oneNewItem,
  onEdit,
  onDelete,
}) {
  const { isWindowSmall } = useWindowSize();

  const [selected, setSelected] = useState({
    id: 0,
    first_name: "",
    last_name: "",
  });
  const [showDeleteModal, setShowDeleteModal] = useState(false);

  const columnDef = [
    { field: "id", hide: true },
    {
      field: "name",
      headerName: "Name",
      flex: 1,
      resizable: false,
      ...(isWindowSmall ? { minWidth: 180 } : {}),
      valueGetter: (params) =>
        `${params.data.first_name} ${params.data.last_name}`.trim(),
      cellRenderer: (params) => {
        return (
          <div className="flex h-full flex-row items-center gap-3">
            <Avatar
              styles="size-9! text-xs! shrink-0"
              initials={`${params.data.first_name[0]}${params.data.last_name[0]}`}
            />
            <div className="flex flex-col justify-center">
              <p className="text-sm">{`${params.data.first_name} ${params.data.last_name}`}</p>
              <p className="text-fg-muted text-xs">{params.data.email}</p>
            </div>
          </div>
        );
      },
    },
    {
      field: "role",
      headerName: "Role",
      flex: 1,
      resizable: false,
      ...(isWindowSmall ? { minWidth: 80 } : {}),
      cellStyle: {
        display: "flex",
        alignItems: "center",
      },
    },
    {
      field: "phone_number",
      headerName: "Phone number",
      flex: 1,
      resizable: false,
      hide: isWindowSmall,
      ...(isWindowSmall ? { minWidth: 140 } : {}),
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
        return (
          <div
            key={params.data.id}
            className="flex h-full flex-row items-center justify-center"
          >
            <button
              className="text-fg-muted hover:bg-hover hover:text-fg
                cursor-pointer rounded-lg p-2"
              onClick={() => onEdit(data[params.node.sourceRowIndex])}
            >
              <Icon icon={Edit03Icon} styles="size-4.5" />
            </button>
            {params.data.role !== "owner" && (
              <button
                className="text-fg-muted hover:bg-hover hover:text-fg
                  cursor-pointer rounded-lg p-2"
                onClick={() => {
                  setSelected({
                    id: params.data.id,
                    first_name: params.data.first_name,
                    last_name: params.data.last_name,
                  });
                  setShowDeleteModal(true);
                }}
              >
                <Icon icon={Delete02Icon} styles="text-fg-danger size-4.5" />
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
      <DeleteModal
        itemName={`${selected.first_name} ${selected.last_name}`}
        open={showDeleteModal}
        onOpenChange={setShowDeleteModal}
        onDelete={() => onDelete(selected)}
      />
      <Suspense fallback={<Loading />}>
        <Table
          rowData={data}
          columnDef={columnDef}
          itemName="team member"
          onNewItem={oneNewItem}
          onRowClick={onRowClick}
          exportName="team_members"
          columnsToExport={["name, role, email, phone_number"]}
          rowHeight={64}
        />
      </Suspense>
    </div>
  );
}
