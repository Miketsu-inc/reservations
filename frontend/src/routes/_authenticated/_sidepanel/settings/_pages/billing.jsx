import { createFileRoute } from "@tanstack/react-router";
import SectionHeader from "../-components/SectionHeader";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_pages/billing"
)({
  component: BillingPage,
  loader: () => ({ crumb: "Billing" }),
});

function BillingPage() {
  return (
    <div className="flex flex-col">
      <SectionHeader title="Billing" styles="" />
    </div>
  );
}
