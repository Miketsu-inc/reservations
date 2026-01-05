import { EditIcon, TrashBinIcon } from "@reservations/assets";
import { DeleteModal, Loading } from "@reservations/components";
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
  const windowSize = useWindowSize();
  const isSmallScreen =
    windowSize !== "xl" || windowSize != +"2xl" || windowSize !== "3xl";

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
      ...(isSmallScreen ? { minWidth: 160 } : {}),
      valueGetter: (params) =>
        `${params.data.first_name} ${params.data.last_name}`.trim(),
    },
    {
      field: "role",
      headerName: "Role",
      flex: 1,
      ...(isSmallScreen ? { minWidth: 120 } : {}),
    },
    {
      field: "email",
      headerName: "Email",
      flex: 1,
      ...(isSmallScreen ? { minWidth: 120 } : {}),
    },
    {
      field: "phone_number",
      headerName: "Phone number",
      flex: 1,
      ...(isSmallScreen ? { minWidth: 120 } : {}),
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
              <EditIcon styles="size-4 mx-1" />
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
                <TrashBinIcon styles="size-5 mx-1" />
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
        />
      </Suspense>
    </div>
  );
}
