import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/_sidepanel/settings/")({
  beforeLoad: () => {
    throw redirect({
      to: "/settings/profile",
    });
  },
});
