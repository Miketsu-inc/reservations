import { PreviewCard as PreviewCardPrimitive } from "@base-ui/react";
import React from "react";

export function PreviewCard({ ...props }) {
  return <PreviewCardPrimitive.Root data-slot="preview-card" {...props} />;
}

export function PreviewCardTrigger({ asChild = false, children, ...props }) {
  return (
    <PreviewCardPrimitive.Trigger
      data-slot="preview-card-trigger"
      render={
        asChild &&
        (React.isValidElement(children) || typeof children === "function")
          ? children
          : undefined
      }
      {...props}
    >
      {asChild ? undefined : children}
    </PreviewCardPrimitive.Trigger>
  );
}

export function PreviewCardContent({
  styles = "",
  align = "center",
  side = "bottom",
  sideOffset = 8,
  collisionPadding = 5,
  ...props
}) {
  return (
    <PreviewCardPrimitive.Portal>
      <PreviewCardPrimitive.Positioner
        className="z-70"
        align={align}
        side={side}
        sideOffset={sideOffset}
        collisionPadding={collisionPadding}
      >
        <PreviewCardPrimitive.Popup
          data-slot="preview-card-content"
          className={`${styles} bg-layer_bg text-text_color border-border_color
            data-[side=bottom]:slide-in-from-top-2
            data-[side=inline-end]:slide-in-from-left-2
            data-[side=inline-start]:slide-in-from-right-2
            data-[side=left]:slide-in-from-right-2
            data-[side=right]:slide-in-from-left-2
            data-[side=top]:slide-in-from-bottom-2 data-open:animate-in
            data-open:fade-in-0 data-open:zoom-in-95 data-closed:animate-out
            data-closed:fade-out-0 data-closed:zoom-out-95
            origin-(--transform-origin) rounded-xl border shadow-lg
            outline-hidden duration-100 dark:shadow-gray-950`}
          {...props}
        />
      </PreviewCardPrimitive.Positioner>
    </PreviewCardPrimitive.Portal>
  );
}
