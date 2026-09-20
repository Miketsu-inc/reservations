import { Popover as PopoverPrimitive } from "@base-ui/react";
import React from "react";

export function Popover({ ...props }) {
  return <PopoverPrimitive.Root data-slot="popover" {...props} />;
}

export function PopoverClose({
  asChild = false,
  nativeButton = true,
  children,
  ...props
}) {
  return (
    <PopoverPrimitive.Close
      render={
        asChild &&
        (React.isValidElement(children) || typeof children === "function")
          ? children
          : undefined
      }
      nativeButton={nativeButton}
      {...props}
    >
      {asChild ? undefined : children}
    </PopoverPrimitive.Close>
  );
}

export function PopoverTrigger({
  asChild = false,
  nativeButton = true,
  children,
  ...props
}) {
  return (
    <PopoverPrimitive.Trigger
      data-slot="popover-trigger"
      render={
        asChild &&
        (React.isValidElement(children) || typeof children === "function")
          ? children
          : undefined
      }
      nativeButton={nativeButton}
      {...props}
    >
      {asChild ? undefined : children}
    </PopoverPrimitive.Trigger>
  );
}

export function PopoverContent({
  styles,
  align = "center",
  side = "bottom",
  sideOffset = 4,
  ...props
}) {
  return (
    // TODO: see if removed 'forceMount' from Select is causing problems.
    // could use 'keepMounted' if required
    <PopoverPrimitive.Portal>
      <PopoverPrimitive.Backdrop />
      <PopoverPrimitive.Positioner
        className="z-50"
        align={align}
        side={side}
        sideOffset={sideOffset}
      >
        <PopoverPrimitive.Popup
          data-slot="popover-content"
          className={`${styles} bg-layer_bg text-text_color border-border_color
            data-[side=bottom]:slide-in-from-top-2
            data-[side=inline-end]:slide-in-from-left-2
            data-[side=inline-start]:slide-in-from-right-2
            data-[side=left]:slide-in-from-right-2
            data-[side=right]:slide-in-from-left-2
            data-[side=top]:slide-in-from-bottom-2 data-open:animate-in
            data-open:fade-in-0 data-open:zoom-in-95 data-closed:animate-out
            data-closed:fade-out-0 data-closed:zoom-out-95 w-48
            origin-(--transform-origin) rounded-lg border p-2 shadow-md
            outline-hidden duration-100 dark:shadow-gray-950`}
          {...props}
        />
      </PopoverPrimitive.Positioner>
    </PopoverPrimitive.Portal>
  );
}
