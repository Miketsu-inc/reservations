import { Slider as SliderPrimitive } from "@base-ui/react";

export default function Slider({ styles = "", ...props }) {
  return (
    <SliderPrimitive.Root data-slot="slider" className={styles} {...props}>
      <SliderPrimitive.Control
        className="flex w-full touch-none items-center py-3 select-none"
      >
        <SliderPrimitive.Track
          className="bg-bg-inactive relative h-1.5 w-full rounded-full"
        >
          <SliderPrimitive.Indicator className="bg-primary rounded-full" />
          <SliderPrimitive.Thumb
            className="border-border-strong shadow-card
              focus-visible:outline-primary size-5 cursor-grab rounded-full
              border bg-white outline-offset-2 focus-visible:outline-2
              active:cursor-grabbing"
          />
        </SliderPrimitive.Track>
      </SliderPrimitive.Control>
    </SliderPrimitive.Root>
  );
}
