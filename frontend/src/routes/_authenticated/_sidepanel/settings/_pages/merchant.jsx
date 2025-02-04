import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/settings/_pages/merchant"
)({
  component: MerchantPage,
});

function MerchantPage() {
  return <div className="flex w-full flex-col gap-6">General Info</div>;
}
