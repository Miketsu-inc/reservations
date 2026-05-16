import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/_navigation/favorites")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="flex h-full justify-center">
      <div className="flex w-full max-w-md flex-col">
        <p className="pt-4 pb-8 text-2xl">Favorites</p>
        <div className="flex h-full flex-col justify-center text-center">
          <p className="text-lg">You do not have any favorites</p>
          <p className="text-text_color/60">
            Try adding some from them to appear here
          </p>
        </div>
      </div>
    </div>
  );
}
