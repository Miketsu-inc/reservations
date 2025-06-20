import DeleteModal from "@components/DeleteModal";
import { Popover, PopoverContent, PopoverTrigger } from "@components/Popover";
import ArrowIcon from "@icons/ArrowIcon";
import BackArrowIcon from "@icons/BackArrowIcon";
import EditIcon from "@icons/EditIcon";
import ThreeDotsIcon from "@icons/ThreeDotsIcon";
import TrashBinIcon from "@icons/TrashBinIcon";
import { useToast } from "@lib/hooks";
import { PopoverClose } from "@radix-ui/react-popover";
import { useState } from "react";
import EditServiceCategoryModal from "./EditServiceCategoryModal";

export default function ServiceCategoryCard({
  children,
  category,
  categoryCount,
  refresh,
  onMoveUp,
  onMoveDown,
}) {
  const [isCollapsed, setIsCollapsed] = useState(
    localStorage.getItem(`category_${category.id}_collapsed`) === "true"
  );
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const { showToast } = useToast();

  async function deleteHandler() {
    const response = await fetch(
      `/api/v1/merchants/services/categories/${category.id}`,
      {
        method: "DELETE",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
      }
    );

    if (!response.ok) {
      const result = await response.json();
      showToast({ message: result.error.message, variant: "error" });
    } else {
      showToast({
        message: "Category deleted successfully",
        variant: "success",
      });

      refresh();
    }
  }

  return (
    <div
      className="border-border_color overflow-hidden rounded-lg border bg-zinc-200/50
        dark:bg-zinc-950/50"
    >
      <DeleteModal
        itemName={category.name}
        isOpen={isDeleteModalOpen}
        onClose={() => setIsDeleteModalOpen(false)}
        onDelete={deleteHandler}
      />
      <EditServiceCategoryModal
        category={category}
        isOpen={isEditModalOpen}
        onClose={() => setIsEditModalOpen(false)}
        onModified={refresh}
      />
      <div
        className={`${!isCollapsed ? "border-border_color border-b" : ""} flex flex-row items-center
          justify-between gap-2 p-4`}
      >
        <div className="flex flex-row items-center gap-4">
          <div className="flex size-12 shrink-0 overflow-hidden rounded-lg">
            <img
              className="size-full object-cover"
              src="https://dummyimage.com/70x70/d156c3/000000.jpg"
              alt="service photo"
            />
          </div>
          <p className="text-lg font-semibold">{`${category.id ? `${category.name}` : "Uncategorized"}`}</p>
        </div>
        <div className="flex flex-row gap-3">
          {category.id !== null && (
            <Popover>
              <PopoverTrigger asChild>
                <button className="hover:bg-hvr_gray hover:*:stroke-text_color cursor-pointer rounded-lg p-1">
                  <ThreeDotsIcon styles="size-6 stroke-4 stroke-gray-400 dark:stroke-gray-500" />
                </button>
              </PopoverTrigger>
              <PopoverContent side="left">
                <div
                  className="flex flex-col items-start *:flex *:w-full *:flex-row *:items-center *:rounded-lg
                    *:p-2"
                >
                  <PopoverClose asChild>
                    <button
                      onClick={() => setIsEditModalOpen(true)}
                      className="hover:bg-hvr_gray cursor-pointer gap-5"
                    >
                      <EditIcon styles="size-4 ml-1" />
                      <p>Edit category</p>
                    </button>
                  </PopoverClose>
                  <PopoverClose asChild>
                    <button
                      disabled={category.sequence === categoryCount}
                      onClick={() => onMoveDown(category.id)}
                      className={`${category.sequence === categoryCount ? "opacity-35" : "hover:bg-hvr_gray cursor-pointer"}
                      gap-4`}
                    >
                      <ArrowIcon styles="size-6 stroke-current" />
                      <p>Move down</p>
                    </button>
                  </PopoverClose>
                  <PopoverClose asChild>
                    <button
                      disabled={category.sequence === 1}
                      onClick={() => onMoveUp(category.id)}
                      className={`${category.sequence === 1 ? "opacity-35" : "hover:bg-hvr_gray cursor-pointer"}
                      gap-4`}
                    >
                      <ArrowIcon styles="size-6 rotate-180 stroke-current" />
                      <p>Move up</p>
                    </button>
                  </PopoverClose>
                  <PopoverClose asChild>
                    <button
                      onClick={() => setIsDeleteModalOpen(true)}
                      className="hover:bg-hvr_gray cursor-pointer gap-4"
                    >
                      <TrashBinIcon styles="size-5 ml-0.5 mb-0.5" />
                      <p className="text-red-600 dark:text-red-500">
                        Delete category
                      </p>
                    </button>
                  </PopoverClose>
                </div>
              </PopoverContent>
            </Popover>
          )}
          <button
            className="hover:bg-hvr_gray hover:*:stroke-text_color cursor-pointer rounded-lg p-1"
            onClick={() => {
              localStorage.setItem(
                `category_${category.id}_collapsed`,
                !isCollapsed
              );
              setIsCollapsed(!isCollapsed);
            }}
          >
            <BackArrowIcon
              styles={`${isCollapsed ? "-rotate-90" : "rotate-90"} transition-transform duration-200
                size-6 stroke-4 stroke-gray-400 dark:stroke-gray-500`}
            />
          </button>
        </div>
      </div>
      <div
        className={`${isCollapsed ? "grid-rows-[0fr] opacity-0" : "grid-rows-[1fr] opacity-100"}
          transition-[grid, opcaity] grid duration-300 ease-in-out`}
      >
        <div className="overflow-hidden">
          <div className="p-4">
            {category.services.length > 0 ? (
              children
            ) : (
              <div className="bg-layer_bg border-text_color rounded-lg border border-dashed py-18">
                <p className="text-center">
                  Drop a service here to add it to the category
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
