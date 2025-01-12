import ConfirmModal from "@components/ConfirmModal";
import {
  CellStyleModule,
  ClientSideRowModelModule,
  QuickFilterModule,
  themeAlpine,
} from "ag-grid-community";
import { AgGridReact } from "ag-grid-react";
import { useState } from "react";
import TableActions from "../../-components/TableActions";
import TableColorPicker from "./TableColorPicker";

function currencyFormatter(params) {
  return params.value.toLocaleString();
}

export default function ServicesTable({
  onDelete,
  onEdit,
  searchText,
  servicesData,
}) {
  const [selected, setSelected] = useState({ id: 0, name: "" });
  const [showModal, setShowModal] = useState(false);

  const columnDef = [
    { field: "id", hide: true, sort: "asc" },
    { field: "name", flex: 1, minWidth: 70 },
    {
      field: "color",
      cellRenderer: ({ data }) => {
        return <TableColorPicker key={data.id} value={data.color} />;
      },
      sortable: false,
      minWidth: 90,
      maxWidth: 90,
    },
    { field: "description", flex: 2, minWidth: 125 },
    {
      field: "duration",
      headerName: "Duration (min)",
      cellClass: "text-right",
      maxWidth: 150,
    },
    {
      field: "price",
      headerName: "Price (HUF)",
      valueFormatter: currencyFormatter,
      cellClass: "text-right",
      maxWidth: 150,
    },
    {
      field: "cost",
      headerName: "Cost (HUF)",
      valueFormatter: currencyFormatter,
      cellClass: "text-right",
      maxWidth: 150,
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
      maxWidth: 100,
    },
  ];

  return (
    <div className="h-full w-full">
      <ConfirmModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        onSubmit={() => onDelete(selected)}
        headerText="Confirmation"
      >
        <div className="py-4">
          <p>
            Are you sure you want to delete this service?
            <span className="font-bold"> {selected.name}</span>
          </p>
          <p className="text-red-500">
            This is a permanent action and cannot be reverted!
          </p>
        </div>
      </ConfirmModal>
      <AgGridReact
        theme={themeAlpine}
        quickFilterText={searchText}
        modules={[ClientSideRowModelModule, QuickFilterModule, CellStyleModule]}
        rowData={servicesData}
        columnDefs={columnDef}
        defaultColDef={{ sortable: true }}
        getRowId={(params) => String(params.data.id)}
      />
    </div>
  );
}
