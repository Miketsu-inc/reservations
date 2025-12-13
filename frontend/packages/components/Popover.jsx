import * as PopoverPrimitive from "@radix-ui/react-popover";

export function Popover({ ...props }) {
  return <PopoverPrimitive.Root data-slot="popover" {...props} />;
}

export function PopoverClose({ ...props }) {
  return <PopoverPrimitive.PopoverClose {...props} />;
}

export function PopoverTrigger({ asChild, ...props }) {
  return (
    <PopoverPrimitive.Trigger
      data-slot="popover-trigger"
      asChild={asChild}
      {...props}
    />
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
    <PopoverPrimitive.Portal>
      <PopoverPrimitive.Content
        data-slot="popover-content"
        side={side}
        align={align}
        sideOffset={sideOffset}
        className={`${styles} bg-layer_bg text-text_color
          data-[state=open]:animate-in data-[state=closed]:animate-out
          data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0
          data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95
          data-[side=bottom]:slide-in-from-top-2
          data-[side=left]:slide-in-from-right-2
          data-[side=right]:slide-in-from-left-2
          data-[side=top]:slide-in-from-bottom-2 border-border_color z-50 w-48
          origin-(--radix-popover-content-transform-origin) rounded-lg border
          p-2 shadow-md outline-hidden dark:shadow-gray-950`}
        {...props}
      />
    </PopoverPrimitive.Portal>
  );
}

export function PopoverAnchor({ ...props }) {
  return <PopoverPrimitive.Anchor data-slot="popover-anchor" {...props} />;
}
