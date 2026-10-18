import { meQueryOptions } from "@reservations/lib";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/_navigation/home")({
  component: RouteComponent,
});

function RouteComponent() {
  const { data: user } = useQuery(meQueryOptions());

  return (
    <div className="flex justify-center">
      <div className="flex w-full max-w-md flex-col justify-center">
        <h1 className="pt-4 pb-6 text-xl font-semibold">Home</h1>
        <div>
          <p className="text-fg-muted">Welcome back {user.first_name} 👋</p>
        </div>
      </div>
    </div>
  );
}
