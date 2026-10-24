import {
  Call02Icon,
  Delete02Icon,
  Edit03Icon,
  Mail01Icon,
  MoreVerticalIcon,
} from "@hugeicons/core-free-icons";
import {
  AvatarUpload,
  Card,
  DeleteModal,
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
  calendarTeamMembersQueryOptions,
  employeeQueryOptions,
  invalidateLocalStorageAuth,
  meQueryOptions,
  useToast,
} from "@reservations/lib";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { useState } from "react";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/team/_topnav/members/$id"
)({
  component: RouteComponent,
  loader: async ({
    context: {
      queryClient,
      authContext: { merchantId },
    },
    params,
  }) => {
    await queryClient.ensureQueryData(
      employeeQueryOptions(merchantId, params.id)
    );
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function RouteComponent() {
  const { id } = Route.useParams({ from: Route.id });
  const router = useRouter();
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [serverError, setServerError] = useState();
  const { showToast } = useToast();
  const { merchantId, employeeId } = useAuth();
  const queryClient = useQueryClient();

  const {
    data: employee,
    isLoading,
    isError,
    error,
  } = useQuery(employeeQueryOptions(merchantId, id));

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  async function refreshAvatar() {
    await Promise.all([
      queryClient.invalidateQueries(employeeQueryOptions(merchantId, id)),
      queryClient.invalidateQueries(
        calendarTeamMembersQueryOptions(merchantId)
      ),
    ]);

    if (employee.id === employeeId) {
      await queryClient.invalidateQueries(meQueryOptions());
      await router.invalidate();
    }
  }

  async function deleteHandler() {
    try {
      const response = await fetch(
        `/api/v1/merchants/${merchantId}/team/${employee.id}`,
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
    <div className="flex justify-center">
      <ServerError error={serverError} />
      <DeleteModal
        itemName={`${employee.first_name} ${employee.last_name}`}
        open={showDeleteModal}
        onOpenChange={setShowDeleteModal}
        onDelete={deleteHandler}
      />
      <div
        className="flex w-full flex-col gap-5 px-3 lg:w-2/3 lg:px-0 2xl:w-1/2"
      >
        <Card styles="flex flex-col items-start gap-4">
          <div className="flex w-full justify-between">
            <div className="flex items-center gap-4">
              <AvatarUpload
                img={employee.avatar_url}
                initials={`${employee.first_name.charAt(0)}${employee.last_name.charAt(0)}`}
                alt={`${employee.first_name} ${employee.last_name}`}
                url={`/api/v1/merchants/${merchantId}/team/${employee.id}/avatar`}
                canRemove={employee.has_own_avatar}
                onChange={refreshAvatar}
              />

              <div className="flex flex-col gap-1">
                <div className="flex flex-col gap-2 md:flex-row md:gap-4">
                  <h2 className="text-fg text-lg font-bold">
                    {employee.first_name} {employee.last_name}
                  </h2>
                </div>
                <div
                  className="text-fg-muted flex flex-row items-center gap-2
                    text-sm"
                >
                  <p>{employee.role}</p>
                  {employee.is_active ? (
                    <p
                      className="bg-success-subtle text-fg-success rounded-lg
                        p-1 text-sm"
                    >
                      active
                    </p>
                  ) : (
                    <p
                      className="bg-danger-subtle text-fg-danger rounded-lg p-1
                        text-sm"
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
                    className="hover:bg-hover hover:*:stroke-fg h-fit
                      cursor-pointer rounded-lg p-1"
                  >
                    <Icon
                      icon={MoreVerticalIcon}
                      styles="size-6 rotate-90 text-fg-muted"
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
                        className="hover:bg-hover cursor-pointer gap-3"
                        onClick={() => {
                          router.navigate({
                            from: Route.fullPath,
                            to: `/team/members/edit/${employee.id}`,
                          });
                        }}
                      >
                        <Icon icon={Edit03Icon} styles="size-5" />
                        <p>Edit team member</p>
                      </button>
                    </PopoverClose>
                    {employee.role !== "owner" && (
                      <PopoverClose asChild>
                        <button
                          onClick={() => setShowDeleteModal(true)}
                          className="hover:bg-hover cursor-pointer gap-3"
                        >
                          <Icon
                            icon={Delete02Icon}
                            styles="text-fg-danger size-5 mb-0.5"
                          />
                          <p className="text-fg-danger">Delete team member</p>
                        </button>
                      </PopoverClose>
                    )}
                  </div>
                </PopoverContent>
              </Popover>
            </div>
          </div>

          <div
            className="text-fg-muted flex flex-col items-start gap-3 text-sm
              sm:flex-row sm:items-center sm:gap-6"
          >
            {employee.email && (
              <div className="flex items-center gap-2">
                <Icon icon={Mail01Icon} styles="size-5 text-fg-muted" />
                {employee.email}
              </div>
            )}
            <div className="flex items-center gap-6 sm:justify-start">
              {employee.phone_number && (
                <div className="flex items-center gap-2">
                  <Icon
                    icon={Call02Icon}
                    styles="size-4 mb-0.5 text-fg-muted"
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
