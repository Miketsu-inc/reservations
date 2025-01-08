import ConfirmModal from "@components/ConfirmModal";
import {
  CellStyleModule,
  ClientSideRowModelModule,
  colorSchemeDarkBlue,
  ColumnAutoSizeModule,
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

const defaultSelected = {
  id: 0,
  rowId: "",
};

export default function ServicesTable({ onDelete, searchText, servicesData }) {
  const [selected, setSelected] = useState(defaultSelected);
  const [showModal, setShowModal] = useState(false);

  const columns = [
    { field: "name", flex: 1 },
    {
      field: "color",
      cellRenderer: ({ data }) => {
        return (
          <TableColorPicker
            value={data.color}
            onChange={(e) => console.log(e)}
          />
        );
      },
    },
    { field: "description", flex: 2 },
    { field: "duration", cellClass: "text-right" },
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
      cellRenderer: ({ data, node }) => {
        return (
          <TableActions
            data={{ id: data.ID, rowId: node.id }}
            onEdit={(data) => console.log(data)}
            onDelete={(data) => {
              setSelected(data), setShowModal(true);
            }}
          />
        );
      },
      resizable: false,
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
          <p>Are you sure you want to delete this service?</p>
          <p className="text-red-500">
            This is a permanent action and cannot be reverted!
          </p>
        </div>
      </ConfirmModal>
      <AgGridReact
        theme={themeAlpine.withPart(colorSchemeDarkBlue)}
        quickFilterText={searchText}
        modules={[
          ClientSideRowModelModule,
          QuickFilterModule,
          ColumnAutoSizeModule,
          CellStyleModule,
        ]}
        rowData={servicesData}
        autoSizeStrategy={{
          type: "fitCellContents",
          colIds: ["name", "color", "duration", "price", "cost", "actions"],
        }}
        editType={"fullRow"}
        columnDefs={columns}
        defaultColDef={{ sortable: true }}
      />
    </div>
  );
}
