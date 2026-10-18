import {
  ArrowLeft01Icon,
  Edit03Icon,
  Settings02Icon,
} from "@hugeicons/core-free-icons";
import {
  AvatarUpload,
  Card,
  Icon,
  Loading,
  ServerError,
} from "@reservations/components";
import { JABULANI_URL, meQueryOptions, useWindowSize } from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, useRouteContext } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/_navigation/profile/")({
  loader: async ({ context: { queryClient } }) => {
    await queryClient.ensureQueryData(meQueryOptions());
  },
  component: RouteComponent,
  errorComponent: ({ error }) => {
    return <ServerError error={error.message} />;
  },
});

function RouteComponent() {
  const { data: user, isLoading } = useQuery(meQueryOptions());
  const { isWindowSmall } = useWindowSize();
  const { queryClient } = useRouteContext({ from: Route.id });

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div className="flex justify-center">
      <div className="flex w-full max-w-sm flex-col justify-center gap-4">
        <h1 className="pt-4 pb-2 text-xl font-semibold">Profile</h1>
        <Card styles="p-6">
          <div className="text-fg-primary -mt-1 -mr-1 flex justify-end text-sm">
            <Link from={Route.fullPath} to="edit">
              <span
                className="hover:bg-primary-subtle flex flex-row items-center
                  gap-1.5 rounded-lg px-2 py-1"
              >
                <Icon icon={Edit03Icon} styles="size-4" />
                <p>Edit</p>
              </span>
            </Link>
          </div>
          <div className="flex flex-col items-center gap-4">
            <AvatarUpload
              styles="size-20 text-2xl!"
              img={user.avatar_url}
              initials={`${user.first_name[0]}${user.last_name[0]}`}
              alt={`${user.first_name} ${user.last_name}`}
              uploadUrl="/api/v1/media/avatar"
              deleteUrl="/api/v1/media/avatar"
              canRemove={!!user.avatar_url}
              onChange={() => queryClient.invalidateQueries(meQueryOptions())}
            />
            <p className="text-xl font-semibold">{`${user.first_name} ${user.last_name}`}</p>
          </div>
          <hr className="border-border-subtle my-6" />
          <div className="flex flex-col gap-4">
            <div>
              <p className="text-fg-muted text-xs tracking-wide uppercase">
                First name
              </p>
              <p className="text-fg mt-0.5 text-sm">{user.first_name}</p>
            </div>
            <div>
              <p className="text-fg-muted text-xs tracking-wide uppercase">
                Last name
              </p>
              <p className="text-fg mt-0.5 text-sm">{user.last_name}</p>
            </div>
            <div>
              <p className="text-fg-muted text-xs tracking-wide uppercase">
                Email
              </p>
              <p className="text-fg mt-0.5 text-sm">{user.email}</p>
            </div>
            <div>
              <p className="text-fg-muted text-xs tracking-wide uppercase">
                Phone number
              </p>
              <p className="text-fg mt-0.5 text-sm">{user.phone_number}</p>
            </div>
          </div>
        </Card>
        {isWindowSmall && (
          <>
            <Card>
              <Link
                className="flex flex-row items-center justify-between p-2.5
                  text-sm"
                from={Route.fullPath}
                to="/settings"
              >
                <span className="flex flex-row items-center gap-3">
                  <Icon icon={Settings02Icon} styles="size-5 text-fg-muted" />
                  <p>Settings</p>
                </span>
                <Icon
                  icon={ArrowLeft01Icon}
                  styles="size-5 text-fg-muted rotate-180"
                />
              </Link>
            </Card>
            <Card>
              <a
                className="flex flex-row items-center justify-between p-2.5
                  text-sm"
                href={JABULANI_URL}
              >
                <p>For businesses</p>
                <Icon
                  icon={ArrowLeft01Icon}
                  styles="size-5 text-fg-muted rotate-180"
                />
              </a>
            </Card>
          </>
        )}
      </div>
    </div>
  );
}
