import { Drawer as DrawerPrimitive } from "@base-ui/react";
import React from "react";

function DrawerProvider({ ...props }) {
  return <DrawerPrimitive.Provider {...props} />;
}

export function Drawer({ ...props }) {
  return (
    <DrawerProvider>
      <DrawerPrimitive.Root {...props} />
    </DrawerProvider>
  );
}

export function DrawerTrigger({ asChild = false, children, ...props }) {
  return (
    <DrawerPrimitive.Trigger
      render={
        asChild &&
        (React.isValidElement(children) || typeof children === "function")
          ? children
          : undefined
      }
      {...props}
    >
      {asChild ? undefined : children}
    </DrawerPrimitive.Trigger>
  );
}

export function DrawerContent({ styles, popUpStyles, ...props }) {
  return (
    <DrawerPrimitive.Portal>
      <DrawerPrimitive.Backdrop
        className="fixed inset-0 min-h-dvh bg-black
          opacity-[calc(var(--backdrop-opacity)*(1-var(--drawer-swipe-progress)))]
          transition-opacity duration-[450ms] ease-[cubic-bezier(0.32,0.72,0,1)]
          [--backdrop-opacity:0.2] [--bleed:3rem] data-[ending-style]:opacity-0
          data-[ending-style]:duration-[calc(var(--drawer-swipe-strength)*400ms)]
          data-[starting-style]:opacity-0 data-[swiping]:duration-0
          supports-[-webkit-touch-callout:none]:absolute
          dark:[--backdrop-opacity:0.7]"
      />
      <DrawerPrimitive.Viewport
        className="fixed inset-0 flex items-end justify-center"
      >
        <DrawerPrimitive.Popup
          className={`bg-layer_bg text-text_color outline-border_color
            -mb-[3rem] max-h-[calc(80vh+3rem)] w-full
            [transform:translateY(var(--drawer-swipe-movement-y))] touch-auto
            overflow-y-auto overscroll-contain rounded-t-2xl px-4 pt-4
            pb-[calc(1.5rem+env(safe-area-inset-bottom,0px)+3rem)] outline
            outline-1 transition-transform duration-[450ms]
            ease-[cubic-bezier(0.32,0.72,0,1)]
            data-[ending-style]:[transform:translateY(calc(100%-3rem+2px))]
            data-[ending-style]:duration-[calc(var(--drawer-swipe-strength)*400ms)]
            data-[starting-style]:[transform:translateY(calc(100%-3rem+2px))]
            data-[swiping]:select-none ${popUpStyles}`}
        >
          <div
            className="mx-auto mb-4 h-1 w-12 rounded-full bg-gray-300
              dark:bg-gray-600"
          />
          <DrawerPrimitive.Content
            className={`${styles} mx-auto flex w-full max-w-md flex-col
              justify-center`}
            {...props}
          />
        </DrawerPrimitive.Popup>
      </DrawerPrimitive.Viewport>
    </DrawerPrimitive.Portal>
  );
}

export function DrawerClose({ asChild = false, children, ...props }) {
  return (
    <DrawerPrimitive.Close
      render={
        asChild &&
        (React.isValidElement(children) || typeof children === "function")
          ? children
          : undefined
      }
      {...props}
    >
      {asChild ? undefined : children}
    </DrawerPrimitive.Close>
  );
}
