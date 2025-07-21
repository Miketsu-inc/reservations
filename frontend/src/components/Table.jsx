import Button from "@components/Button";
import SearchInput from "@components/SearchInput";
import ExportIcon from "@icons/ExportIcon";
import PlusIcon from "@icons/PlusIcon";
import { useWindowSize } from "@lib/hooks";
import {
  CellStyleModule,
  ClientSideRowModelModule,
  ColumnApiModule,
  ColumnAutoSizeModule,
  CsvExportModule,
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
  exportName = "export",
  onRowClick,
  columnsToExport,
  noRowsOverlayComponent,
  showControls = true,
}) {
  const tableRef = useRef();
  const [searchText, setSearchText] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const windowSize = useWindowSize();

  // you should only autosize columns which does not have a flex field
  // as with that the autosize will get applied instead of flex
  // making the table potentially not fill it's grid
  const resetView = useCallback(() => {
    tableRef.current.api.resetColumnState();
    tableRef.current.api.autoSizeColumns(columnsToAutoSize || []);
  }, [columnsToAutoSize]);

  const onBtnExport = useCallback(() => {
    tableRef.current.api.exportDataAsCsv({
      fileName: `${exportName}.csv`,
      columnKeys: columnsToExport,
    });
  }, [exportName, columnsToExport]);

  return (
    <div className="md:bg-layer_bg md:border-border_color flex h-full w-full flex-1 flex-col md:rounded-lg md:border md:px-4 md:py-4 md:shadow-sm">
      <div className="flex flex-col-reverse justify-between gap-2 pb-2 sm:flex-row sm:gap-0">
        <div className="flex items-center justify-center gap-3 pt-2 md:pt-0">
          <div className="w-full md:w-auto">
            <SearchInput
              searchText={searchText}
              onChange={(text) => setSearchText(text)}
            />
          </div>
          <Button
            variant="tertiary"
            styles="p-2 text-sm w-fit text-nowrap"
            buttonText={windowSize != "sm" ? "Export" : ""}
            onClick={onBtnExport}
          >
            <ExportIcon styles="text-text_color md:mr-2 md:mb-0.5 size-5" />
          </Button>
        </div>
        {showControls && (
          <div className="flex flex-row justify-between sm:gap-3">
            <Button
              variant="primary"
              onClick={resetView}
              styles="p-2 text-sm w-fit"
              buttonText="Reset view"
            />
            <Button
              variant="primary"
              onClick={onNewItem}
              styles="p-2 text-sm w-fit"
              buttonText={`New ${itemName}`}
            >
              <PlusIcon styles="size-4 md:size-5 mr-1 text-white" />
            </Button>
          </div>
        )}
      </div>
      <div
        className={`${isLoading ? "invisible" : "visible"} h-full w-full ${rowData?.length === 0 ? "ag-header-hidden" : ""} `}
      >
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
            CsvExportModule,
          ]}
          rowData={rowData}
          columnDefs={columnDef}
          defaultColDef={{ sortable: true, suppressMovable: true }}
          getRowId={(params) => String(params.data.id)}
          onFirstDataRendered={(params) => {
            params.api.autoSizeColumns(columnsToAutoSize || []);
          }}
          onGridReady={() => setIsLoading(false)}
          // suppressColumnVirtualisation is needed for autosizing to work on mobile
          // if disabled only columns in view will get autosized
          suppressColumnVirtualisation={true}
          onRowClicked={onRowClick}
          noRowsOverlayComponent={noRowsOverlayComponent}
        />
      </div>
    </div>
  );
}
