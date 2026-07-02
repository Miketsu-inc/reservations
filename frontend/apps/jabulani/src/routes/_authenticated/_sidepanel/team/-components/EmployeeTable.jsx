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
          <div className="flex h-full flex-row items-center gap-2">
            <Avatar
              styles="size-12! text-sm shrink-0"
              initials={`${params.data.first_name[0]}${params.data.last_name[0]}`}
            />
            <div className="flex flex-col justify-center">
              <p className="text-base">{`${params.data.first_name} ${params.data.last_name}`}</p>
              <p className="text-text_color/60 text-sm">{params.data.email}</p>
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
              className="cursor-pointer"
              onClick={() => onEdit(data[params.node.sourceRowIndex])}
            >
              <Icon icon={Edit03Icon} styles="size-5 mx-1" />
            </button>
            {params.data.role !== "owner" && (
              <button
                className="cursor-pointer"
                onClick={() => {
                  setSelected({
                    id: params.data.id,
                    first_name: params.data.first_name,
                    last_name: params.data.last_name,
                  });
                  setShowDeleteModal(true);
                }}
              >
                <Icon
                  icon={Delete02Icon}
                  styles="text-red-600 dark:text-red-500 size-5 mx-1"
                />
              </button>
            )}
          </div>
        );
      },
      resizable: false,
      sortable: false,
      minWidth: 120,
      maxWidth: 120,
    },
  ];

  return (
    <div className="h-full w-full">
      <DeleteModal
        itemName={`${selected.first_name} ${selected.last_name}`}
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
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
          rowHeight={80}
        />
      </Suspense>
    </div>
  );
}
