import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/_sidepanel/team/")({
  beforeLoad: () => {
    throw redirect({
      to: "/team/members",
    });
  },
});
