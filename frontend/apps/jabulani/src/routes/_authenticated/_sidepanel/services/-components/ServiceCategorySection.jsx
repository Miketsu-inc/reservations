import {
  ArrowLeft01Icon,
  ArrowLeft02Icon,
  Delete02Icon,
  Edit03Icon,
  MoreVerticalIcon,
} from "@hugeicons/core-free-icons";
import {
  DeleteModal,
  Icon,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { useToast } from "@reservations/lib";
import { useState } from "react";
import EditServiceCategoryModal from "./EditServiceCategoryModal";

export default function ServiceCategorySection({
  children,
  category,
  categoryCount,
  refresh,
  onMove,
}) {
  const [isCollapsed, setIsCollapsed] = useState(
    localStorage.getItem(`category_${category.id}_collapsed`) === "true"
  );
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const { showToast } = useToast();
  const { merchantId } = useAuth();

  async function deleteHandler() {
    const response = await fetch(
      `/api/v1/merchants/${merchantId}/service-categories/${category.id}`,
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
    <div className="border-border_color overflow-hidden rounded-lg">
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
      <div className={"flex flex-row items-center justify-between gap-2"}>
        <div className="flex flex-row items-center">
          <p className="text-lg font-semibold">{`${category.id ? `${category.name}` : "Uncategorized"}`}</p>
        </div>
        <div className="flex flex-row gap-3">
          {category.id !== null && (
            <CategoryOptions
              category={category}
              categoryCount={categoryCount}
              onEdit={() => setIsEditModalOpen(true)}
              onDelete={() => setIsDeleteModalOpen(true)}
              onMoveUp={() => onMove(category.id, "forward")}
              onMoveDown={() => onMove(category.id, "backward")}
            />
          )}
          <button
            className="hover:bg-hvr_gray hover:*:stroke-text_color
              cursor-pointer rounded-lg p-1"
            onClick={() => {
              localStorage.setItem(
                `category_${category.id}_collapsed`,
                !isCollapsed
              );
              setIsCollapsed(!isCollapsed);
            }}
          >
            <Icon
              icon={ArrowLeft01Icon}
              styles={`${isCollapsed ? "-rotate-90" : "rotate-90"}
                transition-transform duration-200 size-6 text-gray-400
                dark:text-gray-500`}
            />
          </button>
        </div>
      </div>
      <div
        className={`${isCollapsed ? "grid-rows-[0fr] opacity-0" : "grid-rows-[1fr] opacity-100"}
          transition-[grid, opcaity] grid duration-300 ease-in-out`}
      >
        <div className="overflow-hidden">
          <div className="py-4">
            {category.services.length > 0 ? (
              children
            ) : (
              <div
                className="bg-layer_bg border-text_color rounded-lg border
                  border-dashed py-18"
              >
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

function CategoryOptions({
  category,
  categoryCount,
  onEdit,
  onMoveDown,
  onMoveUp,
  onDelete,
}) {
  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          className="hover:bg-hvr_gray hover:*:stroke-text_color cursor-pointer
            rounded-lg p-1"
        >
          <Icon
            icon={MoreVerticalIcon}
            styles="size-6 rotate-90 text-gray-400 dark:text-gray-500"
          />
        </button>
      </PopoverTrigger>
      <PopoverContent side="left">
        <div
          className="flex flex-col items-start *:flex *:w-full *:flex-row
            *:items-center *:rounded-lg *:p-2"
        >
          <PopoverClose asChild>
            <button
              onClick={onEdit}
              className="hover:bg-hvr_gray cursor-pointer gap-4"
            >
              <Icon icon={Edit03Icon} styles="size-6" />
              <p>Edit category</p>
            </button>
          </PopoverClose>
          <PopoverClose asChild>
            <button
              disabled={category.sequence === categoryCount}
              onClick={onMoveDown}
              className={`${
                category.sequence === categoryCount
                  ? "opacity-35"
                  : "hover:bg-hvr_gray cursor-pointer"
                } gap-4`}
            >
              <Icon
                icon={ArrowLeft02Icon}
                styles="size-6 -rotate-90 stroke-current"
              />
              <p>Move down</p>
            </button>
          </PopoverClose>
          <PopoverClose asChild>
            <button
              disabled={category.sequence === 1}
              onClick={onMoveUp}
              className={`${category.sequence === 1 ? "opacity-35" : "hover:bg-hvr_gray cursor-pointer"}
                gap-4`}
            >
              <Icon
                icon={ArrowLeft02Icon}
                styles="size-6 rotate-90 stroke-current"
              />
              <p>Move up</p>
            </button>
          </PopoverClose>
          <PopoverClose asChild>
            <button
              onClick={onDelete}
              className="hover:bg-hvr_gray cursor-pointer gap-4"
            >
              <Icon
                icon={Delete02Icon}
                styles="size-6 ml-0.5 mb-0.5 text-red-600 dark:text-red-500"
              />
              <p className="text-red-600 dark:text-red-500">Delete category</p>
            </button>
          </PopoverClose>
        </div>
      </PopoverContent>
    </Popover>
  );
}
