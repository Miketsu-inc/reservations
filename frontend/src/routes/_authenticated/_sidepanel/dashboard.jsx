import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/_sidepanel/dashboard")({
  component: DashboardPage,
  loader: () => ({
    crumb: "Dashboard",
  }),
});

function DashboardPage() {
  return <p className="text-text_color">Dashboard page</p>;
}
