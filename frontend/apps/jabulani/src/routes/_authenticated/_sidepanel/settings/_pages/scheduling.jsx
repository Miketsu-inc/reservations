import {
  EditIcon,
  PlusIcon,
  ThreeDotsIcon,
  TrashBinIcon,
} from "@reservations/assets";
import {
  Button,
  Card,
  Loading,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
  ServerError,
} from "@reservations/components";
import {
  blockedTimeTypesQueryOptions,
  formatDuration,
  invalidateLocalStorageAuth,
  useToast,
  useWindowSize,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useCallback, useState } from "react";
import BlockedTypesModal from "../-components/BlockedTypesModal";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_pages/scheduling"
)({
  component: BlockedTimesManager,
  loader: async ({ context: { queryClient } }) => {
    await queryClient.ensureQueryData(blockedTimeTypesQueryOptions());
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function BlockedTimesManager() {
  const { queryClient } = Route.useRouteContext();
  const { showToast } = useToast();
  const windowSize = useWindowSize();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingData, setEditingData] = useState(null);
  const {
    data: blockedTypes,
    isLoading,
    isError,
    error,
  } = useQuery(blockedTimeTypesQueryOptions());

  const hasData = blockedTypes && blockedTypes.length > 0;

  const invalidateBlokedTypesQuery = useCallback(async () => {
    await queryClient.invalidateQueries({
      queryKey: ["blocked-time-types"],
    });
  }, [queryClient]);

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error} />;
  }

  function handleOpenModal(timeType) {
    setEditingData(timeType);
    setIsModalOpen(true);
  }

  async function handleDelete(id) {
    try {
      const response = await fetch(
        `/api/v1/merchants/blocked-time-types/${id}`,
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
        invalidateLocalStorageAuth(response.status);
        showToast({ message: result.error.message, variant: "error" });
      } else {
        showToast({
          message: "Blocked time type deleted successfully",
          variant: "success",
        });
        invalidateBlokedTypesQuery();
      }
    } catch (err) {
      showToast({ message: err.message, variant: "error" });
    }
  }

  return (
    <div className="flex flex-col gap-6 md:pl-10">
      <div className="flex items-center justify-between gap-2">
        <div className="flex flex-col gap-1">
          <div className="text-text_color text-xl font-semibold">
            Blocked time types
          </div>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Create and customize blocked time that can be scheduled in the
            calendar.
          </p>
        </div>
        {hasData && (
          <Button
            variant="primary"
            styles="sm:py-2 sm:px-4 w-fit p-2"
            buttonText={windowSize !== "sm" ? "Add" : ""}
            onClick={() => handleOpenModal()}
          >
            <PlusIcon styles="size-6 sm:size-5 sm:mr-2 sm:mb-0.5 text-white" />
          </Button>
        )}
      </div>

      <div className="flex flex-col gap-3">
        {hasData ? (
          blockedTypes.map((type) => (
            <Card key={type.id} styles="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="text-4xl">{type.icon}</div>
                <div>
                  <div className="text-text_color font-medium">{type.name}</div>
                  <div className="text-sm text-gray-500 dark:text-gray-400">
                    {formatDuration(type.duration)}
                  </div>
                </div>
              </div>

              <Popover>
                <PopoverTrigger asChild>
                  <button
                    className="hover:bg-hvr_gray hover:*:stroke-text_color h-fit
                      cursor-pointer rounded-lg p-1"
                  >
                    <ThreeDotsIcon
                      styles="size-6 stroke-4 stroke-gray-400
                        dark:stroke-gray-500"
                    />
                  </button>
                </PopoverTrigger>
                <PopoverContent side="left" styles="w-auto">
                  <div
                    className="itmes-start flex w-32 flex-col *:flex *:w-full
                      *:flex-row *:items-center *:rounded-lg *:p-2"
                  >
                    <PopoverClose asChild>
                      <button
                        className="hover:bg-hvr_gray cursor-pointer gap-3"
                        onClick={() => handleOpenModal(type)}
                      >
                        <EditIcon styles="size-4 ml-1" />
                        <p>Edit </p>
                      </button>
                    </PopoverClose>

                    <PopoverClose asChild>
                      <button
                        onClick={() => handleDelete(type.id)}
                        className="hover:bg-hvr_gray cursor-pointer gap-3"
                      >
                        <TrashBinIcon styles="size-5 mb-0.5" />
                        <p className="text-red-600 dark:text-red-500">Delete</p>
                      </button>
                    </PopoverClose>
                  </div>
                </PopoverContent>
              </Popover>
            </Card>
          ))
        ) : (
          <div
            className="flex flex-col items-center justify-center gap-2 py-8
              text-center"
          >
            <h3 className="text-text_color mt-4 text-lg font-medium">
              No blocked time types
            </h3>
            <p className="max-w-sm text-sm text-gray-500 dark:text-gray-400">
              Create templates for common breaks like Lunch or Meetings to
              schedule them quickly.
            </p>
            <Button
              variant="primary"
              styles="mt-6 sm:py-2 sm:px-4"
              buttonText="Create your first type"
              onClick={() => handleOpenModal()}
            >
              <PlusIcon styles="size-5 mr-2 mb-0.5 text-white" />
            </Button>
          </div>
        )}
      </div>
      <BlockedTypesModal
        key={editingData?.id || "new"}
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
          setEditingData(null);
        }}
        editData={editingData}
        onSubmit={invalidateBlokedTypesQuery}
      />
    </div>
  );
}
