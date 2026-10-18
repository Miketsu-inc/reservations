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
  alignOffset = 0,
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
        alignOffset={alignOffset}
        side={side}
        sideOffset={sideOffset}
        collisionPadding={collisionPadding}
      >
        <PreviewCardPrimitive.Popup
          data-slot="preview-card-content"
          className={`${styles} bg-bg-overlay text-fg border-border
            shadow-overlay rounded-xl border outline-hidden duration-100`}
          {...props}
        />
      </PreviewCardPrimitive.Positioner>
    </PreviewCardPrimitive.Portal>
  );
}
