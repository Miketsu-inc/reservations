import { Avatar, Button, Loading, ServerError } from "@reservations/components";
import { meQueryOptions, useToast } from "@reservations/lib";
import { queryOptions, useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, redirect } from "@tanstack/react-router";

async function fetchInvitation(token) {
  const response = await fetch(`/api/v1/invitations/${token}`, {
    method: "GET",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
  });

  const result = await response.json();
  if (!response.ok) {
    throw result.error;
  } else {
    return result.data;
  }
}

function invitationQueryOptions(token) {
  return queryOptions({
    queryKey: [token, "get-invitation"],
    queryFn: () => fetchInvitation(token),
  });
}

export const Route = createFileRoute("/invitations")({
  component: RouteComponent,
  loaderDeps: ({ search }) => search,
  beforeLoad: async ({ context: { queryClient } }) => {
    try {
      await queryClient.ensureQueryData(meQueryOptions());
    } catch (error) {
      if (error.status === 401) {
        throw redirect({
          from: Route.fullPath,
          to: "/login",
          search: { redirect: location.href },
        });
      }

      throw error;
    }
  },
  loader: async ({ context: { queryClient }, deps: search }) => {
    await queryClient.ensureQueryData(invitationQueryOptions(search.token));
  },
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function RouteComponent() {
  const { token } = Route.useSearch();
  const { showToast } = useToast();
  const navigate = Route.useNavigate();

  const { data: invitation, isLoading } = useQuery(
    invitationQueryOptions(token)
  );

  if (isLoading) {
    return <Loading />;
  }

  async function acceptHandler() {
    const response = await fetch(`/api/v1/invitations/${token}/accept`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    });

    if (!response.ok) {
      const result = await response.json();
      showToast({ message: result.error.message, variant: "error" });
    }
  }

  async function declineHandler() {
    const response = await fetch(`/api/v1/invitations/${token}/decline`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
    });

    if (!response.ok) {
      const result = await response.json();
      showToast({ message: result.error.message, variant: "error" });
    } else {
      navigate({
        from: Route.fullPath,
        to: "/",
      });
    }
  }

  if (!token) {
    return (
      <div className="flex h-screen items-center justify-center px-4">
        <div className="flex w-full max-w-xl flex-col">
          <div className="pt-4 pb-12">
            <p className="text-2xl">Invalid token query parameter</p>
            <Link
              from={Route.fullPath}
              to="/login"
              className="text-text_color/60"
            >
              Back to login
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-screen items-center justify-center px-4">
      <div className="flex w-full max-w-xl flex-col gap-6">
        <div>
          <p className="text-2xl">Team Invitation</p>
          <p className="text-text_color/60">
            {`You have been invited to ${invitation?.merchant_name} by ${invitation?.invitor_name}`}
          </p>
          <p className="text-text_color/60">
            Accept the invitation to join the team
          </p>
          <div className="flex flex-row justify-center py-4">
            <Avatar styles={"size-24"} />
          </div>
        </div>
        <div className="flex h-full flex-row items-center justify-end gap-4">
          <Button
            styles="py-2 px-4"
            variant="tertiary"
            type="button"
            buttonText="Decline"
            onClick={acceptHandler}
          />
          <Button
            styles="py-2 px-4"
            type="button"
            buttonText="Accept"
            onClick={declineHandler}
          />
        </div>
      </div>
    </div>
  );
}
