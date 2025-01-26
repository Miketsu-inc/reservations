import ConfirmModal from "@components/ConfirmModal";
import Loading from "@components/Loading";
import { useWindowSize } from "@lib/hooks";
import { lazy, Suspense, useState } from "react";
import TableActions from "../../-components/TableActions";
import TableColorPicker from "./TableColorPicker";

const Table = lazy(() => import("@components/Table"));

function currencyFormatter(params) {
  return params.value.toLocaleString();
}

export default function ServicesTable({
  servicesData,
  onDelete,
  onEdit,
  onNewItem,
}) {
  const windowSize = useWindowSize();
  const [selected, setSelected] = useState({ id: 0, name: "" });
  const [showModal, setShowModal] = useState(false);

  const isSmallScreen =
    windowSize === "sm" || windowSize === "md" || windowSize === "lg";

  const columnDef = [
    { field: "id", hide: true, sort: "asc" },
    { field: "name", flex: 1, ...(isSmallScreen ? { minWidth: 120 } : {}) },
    {
      field: "color",
      cellRenderer: ({ data }) => {
        return <TableColorPicker key={data.id} value={data.color} />;
      },
      sortable: false,
      minWidth: 90,
      maxWidth: 90,
    },
    {
      field: "description",
      flex: 2,
      ...(isSmallScreen ? { minWidth: 180 } : {}),
    },
    {
      field: "duration",
      headerName: "Duration (min)",
      cellClass: "text-right",
    },
    {
      field: "price",
      headerName: "Price (HUF)",
      valueFormatter: currencyFormatter,
      cellClass: "text-right",
    },
    {
      field: "cost",
      headerName: "Cost (HUF)",
      valueFormatter: currencyFormatter,
      cellClass: "text-right",
    },
    {
      field: "actions",
      headerName: "",
      cellRenderer: (params) => {
        return (
          <TableActions
            key={params.data.id}
            onEdit={() => onEdit(servicesData[params.node.sourceRowIndex])}
            onDelete={() => {
              setSelected({ id: params.data.id, name: params.data.name });
              setShowModal(true);
            }}
          />
        );
      },
      resizable: false,
      sortable: false,
    },
  ];

  return (
    <div className="h-full w-full">
      <ConfirmModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        onSubmit={() => onDelete(selected)}
        headerText="Delete service"
      >
        <div className="py-4">
          <p>
            Are you sure you want to delete
            <span className="font-bold"> {selected.name}</span>?
          </p>
          <p className="text-red-500">
            This is a permanent action and cannot be reverted!
          </p>
        </div>
      </ConfirmModal>
      <Suspense fallback={<Loading />}>
        <Table
          rowData={servicesData}
          columnDef={columnDef}
          columnsToAutoSize={["duration", "price", "cost", "actions"]}
          itemName="service"
          onNewItem={onNewItem}
        />
      </Suspense>
    </div>
  );
}
