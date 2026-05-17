import {
  ArrowLeft01Icon,
  Edit03Icon,
  Settings02Icon,
} from "@hugeicons/core-free-icons";
import {
  Avatar,
  Card,
  Icon,
  Loading,
  ServerError,
} from "@reservations/components";
import { meQueryOptions, useWindowSize } from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";

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
  const windowSize = useWindowSize();
  const isWindowSmall = windowSize === "sm" || windowSize === "md";

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div className="flex justify-center">
      <div className="flex w-full max-w-sm flex-col justify-center gap-4">
        <p className="pt-4 pb-4 text-2xl">Profile</p>
        <Card styles="p-8">
          <div className="text-primary mb-12 flex justify-end">
            <Link from={Route.fullPath} to="edit">
              <span className="flex flex-row items-center gap-2">
                <Icon icon={Edit03Icon} styles="size-5" />
                <p>Edit</p>
              </span>
            </Link>
          </div>
          <div className="flex flex-col items-center gap-6">
            <Avatar
              styles="size-24 text-2xl!"
              initials={`${user.first_name[0]}${user.last_name[0]}`}
            />
            <p className="text-2xl">{`${user.first_name} ${user.last_name}`}</p>
          </div>
          <hr className="text-border_color my-12" />
          <div className="flex flex-col gap-4">
            <div>
              <p>First name</p>
              <p className="text-text_color/70 text-sm">{user.first_name}</p>
            </div>
            <div>
              <p>Last name</p>
              <p className="text-text_color/70 text-sm">{user.last_name}</p>
            </div>
            <div>
              <p>Email</p>
              <p className="text-text_color/70 text-sm">{user.email}</p>
            </div>
            <div>
              <p>Phone number</p>
              <p className="text-text_color/70 text-sm">{user.phone_number}</p>
            </div>
          </div>
        </Card>
        {isWindowSmall && (
          <>
            <Card>
              <Link
                className="flex flex-row items-center justify-between p-2"
                from={Route.fullPath}
                to="/settings"
              >
                <span className="flex flex-row items-center gap-4">
                  <Icon icon={Settings02Icon} styles="size-6" />
                  <p>Settings</p>
                </span>
                <Icon
                  icon={ArrowLeft01Icon}
                  styles="size-6 text-text_color rotate-180"
                />
              </Link>
            </Card>
            <Card>
              <a
                className="flex flex-row items-center justify-between p-2"
                href="http://app.reservations.local:3000"
              >
                <p className="font-bold">For businesses</p>
                <Icon
                  icon={ArrowLeft01Icon}
                  styles="size-6 text-text_color rotate-180"
                />
              </a>
            </Card>
          </>
        )}
      </div>
    </div>
  );
}
