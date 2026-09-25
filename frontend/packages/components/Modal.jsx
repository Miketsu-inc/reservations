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
      <DialogPrimitive.Backdrop className="fixed inset-0 bg-black/45" />
      <DialogPrimitive.Viewport
        className="fixed inset-0 z-60 flex w-full items-center justify-center
          p-4"
      >
        <DialogPrimitive.Popup
          className={`${styles} bg-layer_bg text-text_color
            dark:border-border_color w-full rounded-lg shadow-lg shadow-gray-500
            transition-all focus:outline-none sm:w-fit dark:border
            dark:shadow-md dark:shadow-gray-950`}
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
