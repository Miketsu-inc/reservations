import BackArrowIcon from "@icons/BackArrowIcon";
import ThreeDotsIcon from "@icons/ThreeDotsIcon";
import { useState } from "react";

export default function ServiceCategoryCard({ children, category }) {
  const [isCollapsed, setIsCollapsed] = useState(
    localStorage.getItem(`category_${category.id}_collapsed`) === "true"
  );

  return (
    <div
      className="border-border_color overflow-hidden rounded-lg border bg-zinc-200/50
        dark:bg-zinc-950/50"
    >
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
          <button className="hover:bg-hvr_gray hover:*:stroke-text_color cursor-pointer rounded-lg p-1">
            <ThreeDotsIcon styles="size-6 stroke-4 stroke-gray-400 dark:stroke-gray-500" />
          </button>
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
