import { Loading, ServerError } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { invalidateLocalStorageAuth, useToast } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import InvitationsTable from "./-components/InvitationsTable";
import NewInvitationDialog from "./-components/NewInvitationDialog";

async function fetchInvitations(merchantId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/team/invitations`,
    {
      method: "GET",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    }
  );

  const result = await response.json();
  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  } else {
    return result.data;
  }
}

function invitationsQueryOptions(merchantId) {
  return queryOptions({
    queryKey: [merchantId, "invitations"],
    queryFn: () => fetchInvitations(merchantId),
  });
}

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/team/_topnav/invitations"
)({
  component: RouteComponent,
  loader: async ({
    context: {
      queryClient,
      authContext: { merchantId },
    },
  }) => {
    await queryClient.ensureQueryData(invitationsQueryOptions(merchantId));
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function RouteComponent() {
  const { showToast } = useToast();
  const { queryClient } = Route.useRouteContext({ from: Route.id });
  const { merchantId } = useAuth();

  const [showNewInvitationDialog, setShowNewInvitationDialog] = useState(false);

  const {
    data: invitations,
    isLoading,
    isError,
    error,
  } = useQuery(invitationsQueryOptions(merchantId));

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return <ServerError error={error.message} />;
  }

  async function revokeHandler(invitation) {
    const response = await fetch(
      `/api/v1/merchants/${merchantId}/team/invitations/${invitation.id}/revoke`,
      {
        method: "POST",
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
        message: "Team member invitation revoked successfully",
        variant: "success",
      });

      await queryClient.invalidateQueries({
        queryKey: [merchantId, "invitations"],
      });
    }
  }

  async function resendHandler(invitation) {
    const response = await fetch(
      `/api/v1/merchants/${merchantId}/team/invitations/${invitation.id}/resend`,
      {
        method: "POST",
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
        message: "Team member invitation resent successfully",
        variant: "success",
      });

      await queryClient.invalidateQueries({
        queryKey: [merchantId, "invitations"],
      });
    }
  }

  return (
    <>
      <p className="pb-6 text-xl">Invitations</p>
      <NewInvitationDialog
        Route={Route}
        isOpen={showNewInvitationDialog}
        onClose={() => setShowNewInvitationDialog(false)}
      />
      <div className="flex min-h-0 flex-1">
        <InvitationsTable
          data={invitations}
          onNewItem={() => setShowNewInvitationDialog(true)}
          onRevoke={revokeHandler}
          onResend={resendHandler}
        />
      </div>
    </>
  );
}
