import * as TooltipPrimitive from "@radix-ui/react-tooltip";

function TooltipProvider({ delayDuration = 0, ...props }) {
  return <TooltipPrimitive.Provider delayDuration={delayDuration} {...props} />;
}

export function Tootlip({ ...props }) {
  return (
    <TooltipProvider>
      <TooltipPrimitive.Root {...props} />
    </TooltipProvider>
  );
}

export function TooltipTrigger({ ...props }) {
  return <TooltipPrimitive.Trigger {...props} />;
}

export function TooltipContent({ styles, sideOffset = 4, children, ...props }) {
  return (
    <TooltipPrimitive.Portal>
      <TooltipPrimitive.Content
        className={`${styles} bg-layer_bg text-text_color animate-in fade-in-0
          zoom-in-95 data-[state=closed]:animate-out
          data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95
          data-[side=bottom]:slide-in-from-top-2
          data-[side=left]:slide-in-from-right-2
          data-[side=right]:slide-in-from-left-2
          data-[side=top]:slide-in-from-bottom-2 border-border_color z-50 w-fit
          origin-(--radix-tooltip-content-transform-origin) rounded-lg border
          p-2 text-xs text-balance shadow-md dark:shadow-gray-950`}
        sideOffset={sideOffset}
        {...props}
      >
        {children}
      </TooltipPrimitive.Content>
    </TooltipPrimitive.Portal>
  );
}
