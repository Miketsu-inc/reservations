import {
  ClientSideRowModelModule,
  ColumnAutoSizeModule,
  QuickFilterModule,
  RowSelectionModule,
  colorSchemeDarkBlue,
  themeAlpine,
} from "ag-grid-community";
import { AgGridReact } from "ag-grid-react";
import TableActions from "../../-components/TableActions";
import TableColorPicker from "./TableColorPicker";

function currencyFormatter(params) {
  return params.value.toLocaleString();
}

const columns = [
  { field: "name", flex: 1 },
  {
    field: "color",
    cellRenderer: ({ data }) => {
      return (
        <TableColorPicker value={data.color} onChange={(e) => console.log(e)} />
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
    cellRenderer: () => {
      return <TableActions />;
    },
    resizable: false,
  },
];

export default function ServicesTable({ searchText, servicesData }) {
  return (
    <div className="h-full w-full">
      <AgGridReact
        theme={themeAlpine.withPart(colorSchemeDarkBlue)}
        quickFilterText={searchText}
        modules={[
          ClientSideRowModelModule,
          RowSelectionModule,
          QuickFilterModule,
          ColumnAutoSizeModule,
        ]}
        rowData={servicesData}
        autoSizeStrategy={{
          type: "fitCellContents",
          colIds: ["name", "color", "duration", "price", "cost", "actions"],
        }}
        columnDefs={columns}
        defaultColDef={{ sortable: true }}
        rowSelection={{ mode: "singleRow" }}
      />
    </div>
  );
}
