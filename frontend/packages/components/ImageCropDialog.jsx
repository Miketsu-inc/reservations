import { Image01Icon } from "@hugeicons/core-free-icons";
import { useCallback, useState } from "react";
import Cropper from "react-easy-crop";
import Button from "./Button.jsx";
import Icon from "./Icon.jsx";
import Slider from "./Slider.jsx";
import {
  ResponsiveDialog,
  ResponsiveDialogClose,
  ResponsiveDialogContent,
} from "./ResponsiveDialog.jsx";

const MIN_ZOOM = 1;
const MAX_ZOOM = 3;
// the crop area is smaller than the cropper, so the rest of the image stays visible around it
const CROP_AREA_RATIO = 0.82;

// The crop passed to onSave is in pixels, relative to the image as it is displayed (EXIF orientation applied)
export default function ImageCropDialog({
  imageSrc,
  title = "Crop image",
  cropShape = "round",
  isSaving,
  onSave,
  ...dialogProps
}) {
  const [containerWidth, setContainerWidth] = useState(0);
  const [crop, setCrop] = useState({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(MIN_ZOOM);
  const [cropPixels, setCropPixels] = useState(null);

  const containerRef = useCallback((container) => {
    if (!container) return;

    const observer = new ResizeObserver(([entry]) =>
      setContainerWidth(entry.contentRect.width)
    );
    observer.observe(container);

    return () => observer.disconnect();
  }, []);

  const cropSize = Math.round(containerWidth * CROP_AREA_RATIO);

  return (
    <ResponsiveDialog {...dialogProps}>
      <ResponsiveDialogContent styles="sm:w-md">
        <div className="flex w-full flex-col gap-4 p-2 sm:p-4">
          <div>
            <p className="text-lg font-semibold">{title}</p>
            <p className="text-fg-muted text-sm">
              Drag to reposition, pinch or use the slider to zoom.
            </p>
          </div>
          <div
            ref={containerRef}
            className="bg-bg-sunken relative aspect-square w-full
              overflow-hidden rounded-xl"
          >
            {imageSrc && cropSize > 0 && (
              <Cropper
                image={imageSrc}
                crop={crop}
                zoom={zoom}
                minZoom={MIN_ZOOM}
                maxZoom={MAX_ZOOM}
                aspect={1}
                cropShape={cropShape}
                cropSize={{ width: cropSize, height: cropSize }}
                objectFit="cover"
                showGrid={false}
                style={{
                  cropAreaStyle: {
                    // tints the image outside the crop area with the dialog background
                    color: "color-mix(in oklch, var(--bg) 72%, transparent)",
                    border: "2px solid white",
                    borderRadius: cropShape === "round" ? "50%" : "1rem",
                  },
                }}
                onCropChange={setCrop}
                onZoomChange={setZoom}
                onCropComplete={(_, pixels) =>
                  setCropPixels({
                    x: Math.round(pixels.x),
                    y: Math.round(pixels.y),
                    width: Math.round(pixels.width),
                    height: Math.round(pixels.height),
                  })
                }
              />
            )}
          </div>
          <div className="flex flex-row items-center gap-5">
            <Icon icon={Image01Icon} styles="size-4 text-fg-muted shrink-0" />
            <Slider
              aria-label="Zoom"
              styles="flex-1"
              min={MIN_ZOOM}
              max={MAX_ZOOM}
              step={0.01}
              value={zoom}
              onValueChange={setZoom}
            />
            <Icon icon={Image01Icon} styles="size-5 text-fg-muted shrink-0" />
          </div>
          <div className="flex flex-row justify-end gap-3">
            <ResponsiveDialogClose asChild>
              <Button
                variant="secondary"
                styles="py-2 px-3 hidden lg:block"
                buttonText="Cancel"
                type="button"
                disabled={isSaving}
              />
            </ResponsiveDialogClose>
            <Button
              styles="py-2 px-3 w-full lg:w-auto"
              buttonText="Save"
              type="button"
              isLoading={isSaving}
              disabled={!cropPixels}
              onClick={() => onSave(cropPixels)}
            />
          </div>
        </div>
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}
