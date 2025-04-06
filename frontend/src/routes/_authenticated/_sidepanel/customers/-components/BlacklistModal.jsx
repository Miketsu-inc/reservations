import Button from "@components/Button";
import Modal from "@components/Modal";

export default function BlacklistModal({ data, isOpen, onClose, onSubmit }) {
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
        <div className="flex justify-center py-3">
          <div className="py-4 text-center">
            <p className="text-gray-700 dark:text-gray-300">
              {data?.is_blacklisted
                ? `You are about to remove this customer from the blacklist.
                They will be able to book appointments by themself from now on.`
                : `You are about to blacklist this customer. They will not be able to
                book appointments from now on. They will see a message asking them
                to contact you for an appointment once trying.`}
              <br />
              You can always revert this action later.
            </p>
          </div>
        </div>
        <div className="flex flex-row items-center justify-end gap-4">
          <Button
            variant="tertiary"
            name="cancel"
            styles="py-2 px-3"
            buttonText="Cancel"
            type="button"
            onClick={onClose}
          />
          <Button
            variant="primary"
            name={data?.is_blacklisted ? "remove" : "blacklist"}
            styles="py-2 px-3"
            buttonText={data?.is_blacklisted ? "Remove" : "Blacklist"}
            type="button"
            onClick={() => {
              onSubmit(data);
              onClose();
            }}
          />
        </div>
      </div>
    </Modal>
  );
}
