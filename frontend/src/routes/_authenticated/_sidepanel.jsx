import { createFileRoute } from "@tanstack/react-router";
import DashboardLayout from "../../pages/dashboard/DashboardLayout";

export const Route = createFileRoute("/_authenticated/_sidepanel")({
  component: DashboardLayout,
});
