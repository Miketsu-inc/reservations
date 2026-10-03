import { createFileRoute } from "@tanstack/react-router";
import SettingsPageHeader from "./-components/SettingsPageHeader";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_topnav/billing"
)({
  component: BillingPage,
});

function BillingPage() {
  return (
    <div className="flex flex-col gap-6 pb-8">
      <SettingsPageHeader
        title="Billing"
        description="Plans, invoices, and payment methods will be managed here."
      />
    </div>
  );
}
