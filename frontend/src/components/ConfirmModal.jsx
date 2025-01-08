import XIcon from "@icons/XIcon";
import { useClickOutside } from "@lib/hooks";
import { useEffect, useRef } from "react";
import Button from "./Button";

export default function ConfirmModal({
  isOpen,
  onClose,
  onSubmit,
  headerText,
  children,
}) {
  const modalRef = useRef();
  useClickOutside(modalRef, onClose);

  useEffect(() => {
    isOpen ? modalRef.current.showModal() : modalRef.current.close();
  }, [isOpen]);

  return (
    <dialog
      className="w-fit rounded-lg bg-layer_bg text-text_color shadow-md shadow-layer_bg
        transition-all backdrop:bg-black backdrop:bg-opacity-35"
      ref={modalRef}
    >
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
    </dialog>
  );
}
