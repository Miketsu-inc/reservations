import { Button, Modal, Textarea } from "@reservations/components";
import { useEffect, useState } from "react";

export default function BlacklistModal({ data, isOpen, onClose, onSubmit }) {
  const [reason, setReason] = useState("");

  // without useEffect data is undefined on first render
  useEffect(() => {
    if (isOpen) {
      setReason(data?.blacklist_reason || "");
    }
  }, [isOpen, data]);

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <div className="m-3 sm:w-md md:m-4">
        <p className="pb-6 text-xl">
          {data?.is_blacklisted
            ? "Remove customer from blacklist"
            : "Blacklist customer"}
        </p>
        <div className="flex items-center justify-center gap-2 py-2">
          <p className="text-lg font-semibold">
            {data?.first_name + " " + data?.last_name}
          </p>
        </div>
        <div
          className="text-text_color/80 flex flex-col justify-center gap-4 py-3
            text-center"
        >
          {data?.is_blacklisted ? (
            <p>
              You are about to remove this customer from the blacklist. They
              will be able to create bookings by themselves from now on.
            </p>
          ) : (
            <p>
              You are about to blacklist this customer. They will not be able to
              create bookings from now on. They will see a message asking them
              to contact you for creating a booking once trying.
            </p>
          )}
          <Textarea
            styles="p-2 max-h-20 min-h-20"
            id="blacklist_reason"
            name="blacklist_reason"
            labelText={data?.is_blacklisted ? "Your Reson for blacklist" : ""}
            required={false}
            placeholder="Add your reson here..."
            value={reason}
            inputData={(data) => setReason(data.value)}
          />
          <p>You can always revert this action later.</p>
        </div>
        <div className="flex flex-row items-center justify-end gap-4">
          <Button
            variant="tertiary"
            name="cancel"
            styles="py-2 px-3"
            buttonText="Cancel"
            type="button"
            onClick={() => {
              if (data?.blacklist_reason) {
                setReason(data.blacklist_reason);
              } else {
                setReason("");
              }
              onClose();
            }}
          />
          <Button
            variant="primary"
            name={data?.is_blacklisted ? "remove" : "blacklist"}
            styles="py-2 px-3"
            buttonText={data?.is_blacklisted ? "Remove" : "Blacklist"}
            type="button"
            onClick={() => {
              onSubmit({
                ...data,
                blacklist_reason: reason.trim(),
              });
              onClose();
            }}
          />
        </div>
      </div>
    </Modal>
  );
}
