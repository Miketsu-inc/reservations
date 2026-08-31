import { createFileRoute } from "@tanstack/react-router";
import SectionHeader from "./-components/SectionHeader";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_topnav/billing"
)({
  component: BillingPage,
});

function BillingPage() {
  return (
    <div className="flex flex-col">
      <SectionHeader title="Billing" />
    </div>
  );
}
