import {
  CustomersIcon,
  PersonIcon,
  PlusIcon,
  ThreeDotsIcon,
  TrashBinIcon,
} from "@reservations/assets";
import {
  Avatar,
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
} from "@reservations/components";
import { useState } from "react";
import CustomerProfile from "./CustomerProfile";
import CustomerSelector from "./CustomerSelector";
import NestedSidePanel from "./NestedSidePanel";

export default function ParticipantManager({
  customers,
  participants = [],
  onRemove,
  onAdd,
  maxParticipants,
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

  return (
    <div className="bg-layer_bg relative flex h-full w-full flex-col">
      <NestedSidePanel
        isOpen={nestedPageState.isOpen}
        onBack={() =>
          setNestedPageState((prev) => ({ ...prev, isOpen: false }))
        }
        styles="size-6"
      >
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
          />
        )}
        {nestedPageState.active === "view" && activeProfile && (
          <CustomerProfile
            customer={activeProfile}
            onRemove={() => {
              onRemove(activeProfile);
              setNestedPageState((prev) => ({ ...prev, isOpen: false }));
            }}
            styles="pt-0!"
          />
        )}
      </NestedSidePanel>

      <div className="flex h-full flex-col pt-16 pb-4">
        <div
          className="border-border_color flex flex-col gap-8 border-b px-4 pb-4"
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

          <button
            className="hover:bg-hvr_gray/20 flex w-full items-center gap-4
              rounded-lg px-3 py-2"
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
        </div>

        <div className="no-scrollbar flex-1 overflow-y-auto px-4 pt-6">
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
                  key={participant.id}
                  customer={participant}
                  onView={() => handleViewProfile(participant)}
                  onRemove={() => onRemove(participant)}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function ParticipantItem({ customer, onView, onRemove }) {
  return (
    <div
      className="group relative flex w-full items-center justify-between
        rounded-xl px-3 py-2 hover:bg-gray-200/40 dark:hover:bg-gray-700/20"
    >
      <div className="flex flex-1 items-center gap-4">
        <Avatar
          styles="size-14! text-[16px]! shrink-0 rounded-full!"
          img={customer?.avatar_url}
          initials={
            customer?.first_name && customer?.last_name
              ? `${customer.first_name[0]}${customer.last_name[0]}`
              : "?"
          }
        />
        <div className="flex flex-col">
          <span className="text-text_color font-semibold">
            {customer.first_name} {customer.last_name}
          </span>
          {customer.phone_number && (
            <span className="text-sm text-gray-400 dark:text-gray-500">
              {customer.phone_number}
            </span>
          )}
        </div>
      </div>

      <Popover>
        <PopoverTrigger asChild>
          <button
            className="group-hover:bg-layer_bg absolute top-5 right-3 h-fit
              cursor-pointer rounded-lg p-1"
          >
            <ThreeDotsIcon styles="size-6 stroke-4 stroke-gray-500 rotate-90" />
          </button>
        </PopoverTrigger>
        <PopoverContent side="left" styles="w-auto">
          <div
            className="itmes-start flex flex-col *:flex *:w-full *:flex-row
              *:items-center *:rounded-lg *:p-2"
          >
            <PopoverClose asChild>
              <button
                className="hover:bg-hvr_gray cursor-pointer gap-3"
                onClick={onView}
              >
                <PersonIcon styles="size-5 fill-text_color" />
                <p>View Customer</p>
              </button>
            </PopoverClose>
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
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}
