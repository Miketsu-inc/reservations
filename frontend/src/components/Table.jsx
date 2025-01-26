import Button from "@components/Button";
import SearchInput from "@components/SearchInput";
import PlusIcon from "@icons/PlusIcon";
import {
  CellStyleModule,
  ClientSideRowModelModule,
  ColumnApiModule,
  ColumnAutoSizeModule,
  QuickFilterModule,
  themeAlpine,
} from "ag-grid-community";
import { AgGridReact } from "ag-grid-react";
import { useCallback, useRef, useState } from "react";

export default function Table({
  rowData,
  columnDef,
  columnsToAutoSize,
  itemName,
  onNewItem,
}) {
  const tableRef = useRef();
  const [searchText, setSearchText] = useState("");
  const [isLoading, setIsLoading] = useState(true);

  const resetView = useCallback(() => {
    tableRef.current.api.resetColumnState();
    tableRef.current.api.autoSizeColumns(columnsToAutoSize);
  }, []);

  return (
    <div className="h-full w-full">
      <div className="flex flex-col-reverse justify-between gap-2 py-2 sm:flex-row sm:gap-0">
        <SearchInput
          searchText={searchText}
          onChange={(text) => setSearchText(text)}
        />
        <div className="flex flex-row justify-between sm:gap-3">
          <Button
            onClick={resetView}
            styles="p-2 text-sm w-fit"
            buttonText="Reset view"
          />
          <Button
            onClick={onNewItem}
            styles="p-2 text-sm w-fit"
            buttonText={`New ${itemName}`}
          >
            <PlusIcon styles="w-4 h-4 md:w-5 md:h-5 mr-1 text-white" />
          </Button>
        </div>
      </div>
      <div className="h-2/3">
        <div className={`${isLoading ? "invisible" : "visible"} h-full w-full`}>
          <AgGridReact
            ref={tableRef}
            theme={themeAlpine}
            quickFilterText={searchText}
            modules={[
              ClientSideRowModelModule,
              QuickFilterModule,
              CellStyleModule,
              ColumnApiModule,
              ColumnAutoSizeModule,
            ]}
            rowData={rowData}
            columnDefs={columnDef}
            defaultColDef={{ sortable: true, suppressMovable: true }}
            getRowId={(params) => String(params.data.id)}
            onFirstDataRendered={(params) => {
              params.api.autoSizeColumns(columnsToAutoSize);
            }}
            onGridReady={() => setIsLoading(false)}
            suppressColumnVirtualisation={true}
          />
        </div>
      </div>
    </div>
  );
}
