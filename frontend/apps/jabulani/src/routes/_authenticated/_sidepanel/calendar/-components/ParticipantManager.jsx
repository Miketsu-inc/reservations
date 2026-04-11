import {
  BanIcon,
  CalendarIcon,
  CustomersIcon,
  PersonIcon,
  PlusIcon,
  SunIcon,
  ThreeDotsIcon,
  TickIcon,
  TrashBinIcon,
} from "@reservations/assets";
import {
  Avatar,
  Drawer,
  DrawerContent,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
} from "@reservations/components";
import { useState } from "react";
import CustomerProfile from "./CustomerProfile";
import CustomerSelector from "./CustomerSelector";
import NestedSidePanel from "./NestedSidePanel";

const statusMap = {
  completed: {
    bgColor: "bg-green-600",
    icon: <TickIcon styles="size-4 text-white" />,
  },
  "no-show": {
    bgColor: "bg-red-500",
    icon: <BanIcon styles="size-3.5 text-white" />,
  },
  booked: {
    bgColor: "bg-gray-500",
    icon: <CalendarIcon styles="size-3.5 text-white" />,
  },
  confirmed: {
    bgColor: "bg-accent",
    icon: <SunIcon styles="size-3.5 text-white" />,
  },
};

export default function ParticipantManager({
  customers,
  participants = [],
  onRemove,
  onAdd,
  maxParticipants,
  disabled,
  onStatusChange,
  isWindowSmall,
}) {
  const [nestedPageState, setNestedPageState] = useState({
    isOpen: false,
    active: "add",
  });

  const [activeProfile, setActiveProfile] = useState(null);

  const handleOpenAdd = () => {
    setNestedPageState({ isOpen: true, active: "add" });
  };

  const handleViewProfile = (customer) => {
    setActiveProfile(customer);
    setNestedPageState({ isOpen: true, active: "view" });
  };

  const ActiveNestedContent = (
    <>
      {nestedPageState.active === "add" && (
        <CustomerSelector
          key={nestedPageState.isOpen}
          onSave={(newParticipants) => {
            onAdd(newParticipants);
            setNestedPageState((prev) => ({ ...prev, isOpen: false }));
          }}
          customers={customers}
          isGroupMode={true}
          selected={participants}
          styles="pt-0!"
          isInDrawer={isWindowSmall}
        />
      )}
      {nestedPageState.active === "view" && activeProfile && (
        <CustomerProfile
          customer={activeProfile}
          disabled={disabled}
          onRemove={() => {
            onRemove(activeProfile);
            setNestedPageState((prev) => ({ ...prev, isOpen: false }));
          }}
          styles="pt-0!"
        />
      )}
    </>
  );

  return (
    <div className="bg-layer_bg relative flex h-full w-full flex-col">
      {isWindowSmall ? (
        <Drawer
          open={nestedPageState.isOpen}
          onOpenChange={(open) =>
            setNestedPageState((prev) => ({ ...prev, isOpen: open }))
          }
          styles="p-0!"
        >
          <DrawerContent
            styles="h-full"
            popUpStyles={`${nestedPageState.active === "view" && activeProfile ? "" : "h-[calc(80vh+3rem)]! overflow-y-hidden!"}`}
          >
            {ActiveNestedContent}
          </DrawerContent>
        </Drawer>
      ) : (
        <NestedSidePanel
          isOpen={nestedPageState.isOpen}
          onBack={() =>
            setNestedPageState((prev) => ({ ...prev, isOpen: false }))
          }
          styles="size-6"
        >
          {ActiveNestedContent}
        </NestedSidePanel>
      )}
      <div
        className={`${isWindowSmall ? "px-5 pt-4" : "px-4 pt-14"} flex h-full
          flex-col pb-4`}
      >
        <div
          className={`border-border_color flex flex-col ${
            disabled ? "gap-5" : "gap-8"
          } border-b pb-4`}
        >
          <div className="flex items-center justify-between gap-2">
            <p className="text-text_color text-xl font-semibold">
              Participants
            </p>
            <span
              className="bg-primary/10 text-primary rounded-md px-2.5 py-0.5
                text-sm font-semibold"
            >
              {participants.length} / {maxParticipants}
            </span>
          </div>
          {!disabled ? (
            <button
              className="flex w-full items-center gap-4 rounded-lg px-3 py-2
                hover:bg-gray-200/40 dark:hover:bg-gray-700/20"
              onClick={handleOpenAdd}
            >
              <div
                className="bg-primary/20 text-primary flex size-14 shrink-0
                  items-center justify-center rounded-full"
              >
                <PlusIcon styles="size-7" />
              </div>
              <div className="flex flex-col items-start">
                <span className="text-text_color font-medium">
                  Add Participants
                </span>
              </div>
            </button>
          ) : (
            <span className="pb-2 text-sm text-gray-500 dark:text-gray-400">
              View all the participants for this booking
            </span>
          )}
        </div>

        <div className="no-scrollbar flex-1 overflow-y-auto pt-6">
          {participants.length === 0 ? (
            <div
              className="flex h-full flex-col items-center justify-start gap-4
                pt-5 text-center opacity-60"
            >
              <div
                className="flex size-18 items-center justify-center rounded-full
                  bg-gray-400/20"
              >
                <CustomersIcon styles="size-12 text-gray-500 dark:text-gray-400" />
              </div>
              <div className="flex flex-col gap-1">
                <p className="text-lg font-medium">No participants yet</p>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Add customers to this group booking.
                </p>
              </div>
            </div>
          ) : (
            <div className="flex flex-col gap-2 pb-20">
              {participants.map((participant) => (
                <ParticipantItem
                  key={participant.customer_id}
                  customer={participant}
                  onView={() => handleViewProfile(participant)}
                  onRemove={() => onRemove(participant)}
                  disabled={disabled}
                  onStatusChange={onStatusChange}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function ParticipantItem({
  customer,
  onView,
  onRemove,
  disabled,
  onStatusChange,
}) {
  const status = statusMap[customer?.status];

  return (
    <div
      className="group relative flex w-full items-center justify-between gap-2
        rounded-xl px-3 py-2 hover:bg-gray-200/40 dark:hover:bg-gray-700/20"
    >
      <div className="flex min-w-0 flex-1 items-center gap-4">
        <Avatar
          styles="size-14! text-[16px]! shrink-0 rounded-full!"
          img={customer?.avatar_url}
          initials={
            customer?.first_name && customer?.last_name
              ? `${customer.first_name[0]}${customer.last_name[0]}`
              : "?"
          }
        />
        <div className="flex min-w-0 flex-col">
          <span className="text-text_color truncate font-semibold">
            {customer.first_name} {customer.last_name}
          </span>

          {customer.phone_number && (
            <span className="text-sm text-gray-400 dark:text-gray-500">
              {customer.phone_number}
            </span>
          )}
        </div>
      </div>
      {customer.participant_id && (
        <div
          className={`ring-layer_bg absolute top-12 left-13 inline-flex size-5
          items-center justify-center gap-2 rounded-full text-xs ring-2
          ${status.bgColor}`}
          title={customer.status}
        >
          {status.icon}
        </div>
      )}

      <Popover>
        <PopoverTrigger asChild>
          <button className="cursor-pointer rounded-lg p-0.5">
            <ThreeDotsIcon styles="size-6 stroke-4 stroke-gray-500 rotate-90" />
          </button>
        </PopoverTrigger>
        <PopoverContent side="left" styles="w-auto">
          <div className="flex flex-col gap-2">
            {customer.participant_id && (
              <div
                className="border-border_color flex flex-col gap-1 border-b pt-1
                  pb-2"
              >
                <span className="pl-2 text-sm text-gray-500 dark:text-gray-400">
                  Set Status
                </span>
                <div
                  className="itmes-start flex flex-col *:flex *:w-full
                    *:flex-row *:items-center *:rounded-lg *:p-2"
                >
                  {customer.status !== "completed" && (
                    <PopoverClose asChild>
                      <button
                        className="hover:bg-hvr_gray cursor-pointer gap-3
                          pl-1.5!"
                        onClick={() =>
                          onStatusChange(
                            customer.participant_id,
                            "completed",
                            customer.status
                          )
                        }
                      >
                        <TickIcon styles="size-6 fill-text_color" />
                        Completed
                      </button>
                    </PopoverClose>
                  )}
                  {customer.status !== "confirmed" && (
                    <PopoverClose asChild>
                      <button
                        className="hover:bg-hvr_gray cursor-pointer gap-3"
                        onClick={() =>
                          onStatusChange(
                            customer.participant_id,
                            "confirmed",
                            customer.status
                          )
                        }
                      >
                        <SunIcon styles="size-5 text-text_color" />
                        Confirmed
                      </button>
                    </PopoverClose>
                  )}
                  {customer.status !== "no-show" && (
                    <PopoverClose asChild>
                      <button
                        className="hover:bg-hvr_gray cursor-pointer gap-3"
                        onClick={() =>
                          onStatusChange(
                            customer.participant_id,
                            "no-show",
                            customer.status
                          )
                        }
                      >
                        <BanIcon styles="size-5" />
                        No Show
                      </button>
                    </PopoverClose>
                  )}
                </div>
              </div>
            )}

            <div
              className="itmes-start flex flex-col *:flex *:w-full *:flex-row
                *:items-center *:rounded-lg *:p-2"
            >
              <PopoverClose asChild>
                <button
                  className="hover:bg-hvr_gray cursor-pointer gap-3"
                  onClick={() => {
                    // let the Popover fully close before the drawer opens
                    setTimeout(() => {
                      onView();
                    }, 50);
                  }}
                >
                  <PersonIcon styles="size-5 fill-text_color" />
                  <p>View Customer</p>
                </button>
              </PopoverClose>
              {!disabled && (
                <PopoverClose asChild>
                  <button
                    onClick={onRemove}
                    className="hover:bg-hvr_gray cursor-pointer gap-3"
                  >
                    <TrashBinIcon styles="size-5 mb-0.5" />
                    <p className="text-red-600 dark:text-red-500">
                      Remove customer
                    </p>
                  </button>
                </PopoverClose>
              )}
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}
