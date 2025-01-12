import XIcon from "@icons/XIcon";
import Button from "./Button";
import Modal from "./Modal";

export default function ConfirmModal({
  isOpen,
  onClose,
  onSubmit,
  headerText,
  children,
}) {
  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <div className="m-2">
        <div className="my-1 flex flex-row items-center justify-between">
          <p className="text-lg md:text-xl">{headerText}</p>
          <XIcon
            styles="h-8 w-8 md:h-9 md:w-9 fill-text_color cursor-pointer"
            onClick={onClose}
          />
        </div>
        <hr className="" />
        {children}
        <hr className="py-1" />
        <div className="flex flex-row items-center justify-end gap-4">
          <Button
            name="cancel"
            styles="p-2 bg-transparent border-2 border-primary"
            buttonText="Cancel"
            onClick={onClose}
          />
          <Button
            name="confirm"
            styles="p-2"
            buttonText="Confirm"
            onClick={(e) => {
              onSubmit(e), onClose();
            }}
          />
        </div>
      </div>
    </Modal>
  );
}
