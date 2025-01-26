import WarningIcon from "@icons/WarningIcon";
import Button from "./Button";
import Modal from "./Modal";

export default function DeleteModal({ isOpen, onClose, onDelete, itemName }) {
  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <div className="m-2 md:m-4">
        <div className="flex justify-center py-2">
          <div className="flex rounded-full bg-red-200 p-3 dark:bg-red-600">
            <WarningIcon styles="w-8 h-8 text-red-600 dark:text-red-200" />
          </div>
        </div>
        <div className="my-1 flex justify-center">
          <p className="text-lg font-semibold md:text-xl">Are you sure?</p>
        </div>
        <div className="flex justify-center py-3">
          <div className="w-4/5 py-4 text-center">
            <p className="text-gray-700 dark:text-gray-300">
              You are about to delete
              <span className="font-bold text-text_color"> {itemName}</span>.
              <br />
              This is a permanent action which cannot be reverted!
            </p>
          </div>
        </div>
        <div className="flex flex-row items-center justify-end gap-4">
          <Button
            variant="tertiary"
            name="cancel"
            styles="py-2 px-3"
            buttonText="Cancel"
            onClick={onClose}
          />
          <Button
            variant="danger"
            name="delete"
            styles="py-2 px-3"
            buttonText="Delete"
            onClick={(e) => {
              onDelete(e), onClose();
            }}
          />
        </div>
      </div>
    </Modal>
  );
}
