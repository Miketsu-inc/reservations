import {
  EditIcon,
  EnvelopeIcon,
  PhoneIcon,
  ThreeDotsIcon,
  TrashBinIcon,
} from "@reservations/assets";
import {
  Avatar,
  Card,
  DeleteModal,
  Loading,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
  ServerError,
} from "@reservations/components";
import {
  invalidateLocalStorageAuth,
  useToast,
  useWindowSize,
} from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";
import { employeeQueryOptions } from "./edit.$id";

export const Route = createFileRoute("/_authenticated/_sidepanel/team/$id")({
  component: RouteComponent,
  loader: async ({ context: { queryClient }, params }) => {
    await queryClient.ensureQueryData(employeeQueryOptions(params.id));
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function RouteComponent() {
  const { id } = Route.useParams({ from: Route.id });
  const router = useRouter();
  const windowSize = useWindowSize();
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();

  const {
    data: employee,
    isLoading,
    isError,
    error,
  } = useQuery(employeeQueryOptions(id));

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  async function deleteHandler() {
    try {
      const response = await fetch(
        `/api/v1/merchants/employees/${employee.id}`,
        {
          method: "DELETE",
          headers: {
            Accept: "application/json",
            "content-type": "application/json",
          },
        }
      );

      if (!response.ok) {
        invalidateLocalStorageAuth(response.status);
        const result = await response.json();
        setServerError(result.error.message);
      } else {
        showToast({
          message: "Team member deleted successfully",
          variant: "success",
        });
        router.navigate({
          from: Route.fullPath,
          to: "/team",
        });
        setServerError();
      }
    } catch (err) {
      setServerError(err.message);
    }
  }

  return (
    <div className="flex justify-center py-5">
      <ServerError error={serverError} />
      <DeleteModal
        itemName={`${employee.first_name} ${employee.last_name}`}
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onDelete={deleteHandler}
      />
      <div
        className="flex w-full flex-col gap-5 px-3 sm:px-0 lg:w-2/3 2xl:w-1/2"
      >
        <Card styles="flex flex-col items-start gap-4">
          <div className="flex w-full justify-between">
            <div className="flex items-center gap-4">
              <Avatar
                initials={`${employee.first_name.charAt(0)}${employee.last_name.charAt(0)}`}
              />

              <div className="flex flex-col gap-1">
                <div
                  className={`flex flex-col gap-2
                    ${windowSize !== "sm" ? "sm:flex-row sm:gap-4" : ""}`}
                >
                  <h2 className="text-text_color text-lg font-bold">
                    {employee.first_name} {employee.last_name}
                  </h2>
                </div>
                <div
                  className="text-text_color/70 flex flex-row items-center gap-2
                    text-sm"
                >
                  <p>{employee.role}</p>
                  {employee.is_active ? (
                    <p
                      className="rounded-lg bg-green-400/30 p-1 text-sm
                        text-green-700 dark:bg-green-700/30 dark:text-green-400"
                    >
                      active
                    </p>
                  ) : (
                    <p
                      className="rounded-lg bg-red-400/30 p-1 text-sm
                        text-red-700 dark:bg-red-700/30 dark:text-red-400"
                    >
                      inactive
                    </p>
                  )}
                </div>
              </div>
            </div>
            <div className="flex flex-col items-start">
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
                    className="itmes-start flex w-auto flex-col *:flex *:w-full
                      *:flex-row *:items-center *:rounded-lg *:p-2"
                  >
                    <PopoverClose asChild>
                      <button
                        className="hover:bg-hvr_gray cursor-pointer gap-3"
                        onClick={() => {
                          router.navigate({
                            from: Route.fullPath,
                            to: `/team/edit/${employee.id}`,
                          });
                        }}
                      >
                        <EditIcon styles="size-4 ml-1" />
                        <p>Edit team member</p>
                      </button>
                    </PopoverClose>
                    {employee.role !== "owner" && (
                      <PopoverClose asChild>
                        <button
                          onClick={() => setShowDeleteModal(true)}
                          className="hover:bg-hvr_gray cursor-pointer gap-3"
                        >
                          <TrashBinIcon styles="size-5 mb-0.5" />
                          <p className="text-red-600 dark:text-red-500">
                            Delete team member
                          </p>
                        </button>
                      </PopoverClose>
                    )}
                  </div>
                </PopoverContent>
              </Popover>
            </div>
          </div>

          <div
            className="text-text_color/70 flex flex-col items-start gap-3
              text-sm sm:flex-row sm:items-center sm:gap-6"
          >
            {employee.email && (
              <div className="flex items-center gap-2">
                <EnvelopeIcon styles="size-5 text-text_color/70" />
                {employee.email}
              </div>
            )}
            <div className="flex items-center gap-6 sm:justify-start">
              {employee.phone_number && (
                <div className="flex items-center gap-2">
                  <PhoneIcon
                    styles="size-4 mb-0.5 fill-text_color/70
                      stroke-text_color/10"
                  />
                  {employee.phone_number}
                </div>
              )}
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}
