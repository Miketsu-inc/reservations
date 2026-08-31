import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/integrations/"
)({
  beforeLoad: () => {
    throw redirect({
      to: "/integrations/calendar",
    });
  },
});
