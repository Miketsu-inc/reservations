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
    <div className="border-border overflow-hidden rounded-lg">
      <DeleteModal
        itemName={category.name}
        open={isDeleteModalOpen}
        onOpenChange={setIsDeleteModalOpen}
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
          <p className="text-base font-semibold">{`${category.id ? `${category.name}` : "Uncategorized"}`}</p>
        </div>
        <div className="flex flex-row gap-0.5">
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
            className="text-fg-muted hover:bg-hover hover:text-fg cursor-pointer
              rounded-lg p-2"
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
                transition-transform duration-200 size-4.5`}
            />
          </button>
        </div>
      </div>
      <div
        className={`${isCollapsed ? "grid-rows-[0fr] opacity-0" : "grid-rows-[1fr] opacity-100"}
          transition-[grid, opcaity] grid duration-300 ease-in-out`}
      >
        <div className="overflow-hidden">
          <div className="pt-2 pb-4">
            {category.services.length > 0 ? (
              children
            ) : (
              <div
                className="border-border-strong text-fg-muted rounded-xl border
                  border-dashed py-10 text-sm"
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
          className="text-fg-muted hover:bg-hover hover:text-fg cursor-pointer
            rounded-lg p-2"
        >
          <Icon icon={MoreVerticalIcon} styles="size-4.5 rotate-90" />
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
              className="hover:bg-hover cursor-pointer gap-2.5"
            >
              <Icon icon={Edit03Icon} styles="size-4.5 text-fg-muted" />
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
                  : "hover:bg-hover cursor-pointer"
                } gap-2.5`}
            >
              <Icon
                icon={ArrowLeft02Icon}
                styles="size-4.5 -rotate-90 text-fg-muted"
              />
              <p>Move down</p>
            </button>
          </PopoverClose>
          <PopoverClose asChild>
            <button
              disabled={category.sequence === 1}
              onClick={onMoveUp}
              className={`${category.sequence === 1 ? "opacity-35" : "hover:bg-hover cursor-pointer"}
                gap-2.5`}
            >
              <Icon
                icon={ArrowLeft02Icon}
                styles="size-4.5 rotate-90 text-fg-muted"
              />
              <p>Move up</p>
            </button>
          </PopoverClose>
          <PopoverClose asChild>
            <button
              onClick={onDelete}
              className="hover:bg-hover cursor-pointer gap-2.5"
            >
              <Icon icon={Delete02Icon} styles="size-4.5 text-fg-danger" />
              <p className="text-fg-danger">Delete category</p>
            </button>
          </PopoverClose>
        </div>
      </PopoverContent>
    </Popover>
  );
}
