import {
  Delete02Icon,
  Edit03Icon,
  MoreVerticalIcon,
  PlusSignIcon,
} from "@hugeicons/core-free-icons";
import {
  Button,
  Card,
  Icon,
  Loading,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
  ServerError,
} from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
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
import BlockedTypesModal from "./-components/BlockedTypesModal";
import SettingsPageHeader from "./-components/SettingsPageHeader";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_topnav/blocked-time-types"
)({
  component: BlockedTimeTypesPage,
  loader: async ({
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => {
    await queryClient.ensureQueryData(blockedTimeTypesQueryOptions(merchantId));
  },
  errorComponent: ({ error }) => <ServerError error={error.message} />,
});

function BlockedTimeTypesPage() {
  const { queryClient } = Route.useRouteContext();
  const { merchantId } = useAuth();
  const { showToast } = useToast();
  const { isWindowSmall } = useWindowSize();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingData, setEditingData] = useState(null);
  const {
    data: blockedTypes,
    isLoading,
    isError,
    error,
  } = useQuery(blockedTimeTypesQueryOptions(merchantId));

  const hasData = blockedTypes?.length > 0;
  const refreshBlockedTypes = useCallback(async () => {
    await queryClient.invalidateQueries(
      blockedTimeTypesQueryOptions(merchantId)
    );
  }, [merchantId, queryClient]);

  function openModal(timeType = null) {
    setEditingData(timeType);
    setIsModalOpen(true);
  }

  async function handleDelete(id) {
    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/blocked-time-types/${id}`,
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
        return;
      }

      showToast({
        message: "Blocked time type deleted successfully",
        variant: "success",
      });
      refreshBlockedTypes();
    } catch (err) {
      showToast({ message: err.message, variant: "error" });
    }
  }

  if (isLoading) return <Loading />;

  if (isError) return <ServerError error={error.message} />;

  return (
    <div className="flex flex-col gap-6 pb-8">
      <SettingsPageHeader
        title="Blocked time types"
        description="Create reusable templates for breaks, meetings, and other unavailable time."
        action={
          hasData ? (
            <Button
              variant="primary"
              styles="px-3 py-2 sm:px-4"
              buttonText={isWindowSmall ? "" : "New type"}
              onClick={() => openModal()}
            >
              <Icon icon={PlusSignIcon} styles="size-5 sm:mr-2" />
            </Button>
          ) : null
        }
      />

      {hasData ? (
        <div className="grid gap-3 sm:grid-cols-2">
          {blockedTypes.map((type) => (
            <Card
              key={type.id}
              styles="flex items-center justify-between gap-4 p-4!"
            >
              <div className="flex min-w-0 items-center gap-3">
                <div className="text-3xl" aria-hidden="true">
                  {type.icon}
                </div>
                <div className="min-w-0">
                  <div className="text-text_color truncate font-medium">
                    {type.name}
                  </div>
                  <div className="text-text_color/60 text-sm">
                    {formatDuration(type.duration)}
                  </div>
                </div>
              </div>

              <Popover>
                <PopoverTrigger asChild>
                  <button
                    type="button"
                    aria-label={`Actions for ${type.name}`}
                    className="hover:bg-hvr_gray shrink-0 rounded-lg p-2"
                  >
                    <Icon
                      icon={MoreVerticalIcon}
                      styles="size-5 rotate-90 text-gray-500"
                    />
                  </button>
                </PopoverTrigger>
                <PopoverContent side="left" styles="w-36 p-1">
                  <div className="flex flex-col text-sm">
                    <PopoverClose asChild>
                      <button
                        type="button"
                        className="hover:bg-hvr_gray flex items-center gap-2
                          rounded-lg p-2"
                        onClick={() => openModal(type)}
                      >
                        <Icon icon={Edit03Icon} styles="size-4" />
                        Edit
                      </button>
                    </PopoverClose>
                    <PopoverClose asChild>
                      <button
                        type="button"
                        className="hover:bg-hvr_gray flex items-center gap-2
                          rounded-lg p-2 text-red-600 dark:text-red-400"
                        onClick={() => handleDelete(type.id)}
                      >
                        <Icon icon={Delete02Icon} styles="size-4" />
                        Delete
                      </button>
                    </PopoverClose>
                  </div>
                </PopoverContent>
              </Popover>
            </Card>
          ))}
        </div>
      ) : (
        <div
          className="border-border_color bg-layer_bg flex flex-col items-center
            rounded-xl border px-4 py-12 text-center shadow-sm"
        >
          <div className="bg-primary/10 text-primary mb-4 rounded-full p-3">
            <Icon icon={PlusSignIcon} styles="size-6" />
          </div>
          <h2 className="text-text_color text-lg font-semibold">
            No blocked time types yet
          </h2>
          <p className="text-text_color/65 mt-1 max-w-sm text-sm leading-6">
            Add templates such as lunch, a meeting, or personal time so they can
            be placed on the calendar quickly.
          </p>
          <Button
            variant="primary"
            styles="mt-6 px-4 py-2"
            buttonText="Create your first type"
            onClick={() => openModal()}
          />
        </div>
      )}

      <BlockedTypesModal
        key={editingData?.id || "new"}
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
          setEditingData(null);
        }}
        editData={editingData}
        onSubmit={refreshBlockedTypes}
      />
    </div>
  );
}
