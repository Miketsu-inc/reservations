import { GoogleIcon } from "@reservations/assets";
import { Button, Card } from "@reservations/components";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/integrations/"
)({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="flex h-full flex-col px-4 py-2 md:px-0 md:py-0">
      <p className="py-6 text-2xl">Integrations</p>
      <p className="pb-6 text-xl">Calendar</p>
      <div className="flex flex-col gap-4">
        <Card styles="flex flex-row justify-between items-center gap-2">
          <div className="flex flex-col gap-2">
            <span className="text-lg">Two-way calendar sync</span>
            <p className="text-sm">
              Events will be automatically synced between your Google calendar
              and our system
            </p>
          </div>
          <a href="http://app.reservations.local:3000/api/v1/merchants/integrations/calendar/google">
            <Button styles="py-2 px-4" buttonText="Sync" onClick={() => {}}>
              <GoogleIcon styles="size-5 fill-text_color mr-3" />
            </Button>
          </a>
        </Card>
      </div>
    </div>
  );
}
