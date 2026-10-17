import { Dialog as DialogPrimitive } from "@base-ui/react";
import React from "react";

export function Modal({ ...props }) {
  return <DialogPrimitive.Root {...props} />;
}

export function ModalTrigger({ asChild = false, children, ...props }) {
  return (
    <DialogPrimitive.Trigger
      render={
        asChild &&
        (React.isValidElement(children) || typeof children === "function")
          ? children
          : undefined
      }
      {...props}
    >
      {asChild ? undefined : children}
    </DialogPrimitive.Trigger>
  );
}

export function ModalContent({ styles = "", ...props }) {
  return (
    <DialogPrimitive.Portal>
      <DialogPrimitive.Backdrop
        className="bg-bg-backdrop fixed inset-0 transition-opacity duration-150
          data-ending-style:opacity-0 data-starting-style:opacity-0"
      />
      <DialogPrimitive.Viewport
        className="fixed inset-0 z-60 flex w-full items-center justify-center
          p-4"
      >
        <DialogPrimitive.Popup
          className={`${styles} bg-bg text-fg border-border shadow-modal w-full
            rounded-2xl border transition-all duration-150 focus:outline-none
            data-ending-style:scale-[0.98] data-ending-style:opacity-0
            data-starting-style:scale-[0.98] data-starting-style:opacity-0
            sm:w-fit`}
          {...props}
        />
      </DialogPrimitive.Viewport>
    </DialogPrimitive.Portal>
  );
}

export function ModalClose({ asChild = false, children, ...props }) {
  return (
    <DialogPrimitive.Close
      render={
        asChild &&
        (React.isValidElement(children) || typeof children === "function")
          ? children
          : undefined
      }
      {...props}
    >
      {asChild ? undefined : children}
    </DialogPrimitive.Close>
  );
}
