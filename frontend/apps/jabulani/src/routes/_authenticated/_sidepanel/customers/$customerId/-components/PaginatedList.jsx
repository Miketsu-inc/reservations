import { BackArrowIcon } from "@reservations/assets";
import { Button, Card } from "@reservations/components";
import { useState } from "react";

export default function PaginatedList({
  data = [],
  itemsPerPage = 6,
  title,
  renderItem,
  emptyMessage = "No items found",
  showItemCount = true,
  styles,
}) {
  const [currentPage, setCurrentPage] = useState(1);

  const totalPages = Math.ceil(data.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const currentItems = data.slice(startIndex, startIndex + itemsPerPage);

  function handlePageChange(page) {
    setCurrentPage(page);
  }

  if (data.length === 0) {
    return (
      <Card styles={`p-0! ${styles}`}>
        <div className="border-border_color border-b p-4">
          <h3 className="text-text_color text-xl font-semibold">{title}</h3>
        </div>
        <div className="p-12 text-center">
          <p className="text-gray-600 dark:text-gray-400">{emptyMessage}</p>
        </div>
      </Card>
    );
  }

  return (
    <Card styles={`p-0! ${styles}`}>
      <div className="border-border_color border-b p-4">
        <div className="flex items-center justify-between">
          <h3 className="text-text_color text-xl font-semibold">{title}</h3>
          {showItemCount && (
            <div className="text-sm text-gray-600 dark:text-gray-300">
              Page {currentPage} of {totalPages}
            </div>
          )}
        </div>
      </div>

      <div className="divide-border_color divide-y">
        {currentItems.map((item, index) => (
          <div key={item.id || index} className="">
            {renderItem(item, startIndex + index)}
          </div>
        ))}
      </div>

      {totalPages > 1 && (
        <div
          className="border-border_color flex w-full justify-center border-t
            px-6 py-4"
        >
          <div className="flex items-center gap-2">
            <Button
              type="button"
              name="previousButton"
              onClick={() => handlePageChange(currentPage - 1)}
              disabled={currentPage === 1}
              variant="tertiary"
              styles="p-2 w-fit rounded-sm"
            >
              <BackArrowIcon styles="size-5 stroke-text_color" />
            </Button>

            <div className="flex gap-1">
              <Pagination
                totalPages={totalPages}
                currentPage={currentPage}
                handlePageChange={handlePageChange}
              />
            </div>

            <Button
              type="button"
              name="nextButton"
              onClick={() => handlePageChange(currentPage + 1)}
              disabled={currentPage === totalPages}
              variant="tertiary"
              styles="p-2 w-fit rounded-sm"
            >
              <span className="flex items-center justify-center gap-1">
                <BackArrowIcon styles="size-5 stroke-text_color rotate-180" />
              </span>
            </Button>
          </div>
        </div>
      )}
    </Card>
  );
}

function Pagination({ totalPages, currentPage, handlePageChange }) {
  const pages = [];

  if (totalPages <= 5) {
    for (let i = 1; i <= totalPages; i++) {
      pages.push(i);
    }
  } else {
    if (currentPage <= 3) {
      pages.push(1, 2, 3, "...", totalPages);
    } else if (currentPage >= totalPages - 2) {
      pages.push(1, "...", totalPages - 2, totalPages - 1, totalPages);
    } else {
      const startPage = currentPage - 1;
      pages.push(startPage, startPage + 1, startPage + 2, "...", totalPages);
    }
  }

  return pages.map((page, index) =>
    page === "..." ? (
      <span
        key={`ellipsis-${index}`}
        className="px-2 py-2 text-sm text-gray-500 dark:text-gray-400"
      >
        ...
      </span>
    ) : (
      <button
        key={page}
        onClick={() => handlePageChange(page)}
        className={`rounded-sm px-3 py-2 text-sm font-medium ${
          currentPage === page
            ? "bg-primary/80 px-3.5 text-white"
            : `border-2 border-gray-300 bg-white text-gray-700 hover:bg-gray-50
              dark:border-gray-800 dark:bg-transparent dark:text-gray-400
              dark:hover:bg-gray-800`
          }`}
      >
        {page}
      </button>
    )
  );
}
