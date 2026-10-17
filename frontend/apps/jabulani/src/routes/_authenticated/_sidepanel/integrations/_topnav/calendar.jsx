import { GoogleIcon } from "@reservations/assets";
import { Button, Card } from "@reservations/components";
import { useAuth } from "@reservations/jabulani/lib";
import { JABULANI_URL } from "@reservations/lib";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/_sidepanel/integrations/_topnav/calendar"
)({
  component: RouteComponent,
});

function RouteComponent() {
  const { merchantId } = useAuth();

  return (
    <div className="flex h-full flex-col px-4 pt-4">
      <h1 className="pb-4 text-xl font-semibold">Calendar</h1>
      <div className="flex flex-col gap-4">
        <Card styles="flex flex-row justify-between items-center gap-2">
          <div className="flex flex-col gap-0.5">
            <span className="text-sm font-semibold">Two-way calendar sync</span>
            <p className="text-fg-muted text-sm">
              Events will be automatically synced between your Google calendar
              and our system
            </p>
          </div>
          <a
            href={`${JABULANI_URL}/api/v1/merchants/${merchantId}/integrations/google/calendar`}
          >
            <Button
              variant="tertiary"
              styles="py-2 pr-3.5 pl-3 text-sm"
              buttonText="Sync"
              onClick={() => {}}
            >
              <GoogleIcon styles="size-4 fill-fg mr-2" />
            </Button>
          </a>
        </Card>
      </div>
    </div>
  );
}
