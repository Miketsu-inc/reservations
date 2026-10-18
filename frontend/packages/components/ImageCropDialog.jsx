import { useState } from "react";
import Cropper from "react-easy-crop";
import Button from "./Button.jsx";
import {
  ResponsiveDialog,
  ResponsiveDialogClose,
  ResponsiveDialogContent,
} from "./ResponsiveDialog.jsx";

// Lets the user select the part of the image to keep. The crop passed to onSave
// is in pixels, relative to the image as it is displayed (EXIF orientation applied)
export default function ImageCropDialog({
  imageSrc,
  title = "Crop image",
  cropShape = "round",
  aspect = 1,
  isSaving,
  onSave,
  ...dialogProps
}) {
  const [crop, setCrop] = useState({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [cropPixels, setCropPixels] = useState(null);

  return (
    <ResponsiveDialog {...dialogProps}>
      <ResponsiveDialogContent styles="sm:w-md">
        <div className="flex flex-col gap-4 p-4">
          <p className="text-lg font-semibold">{title}</p>
          <div
            className="bg-bg-sunken relative h-72 w-full overflow-hidden
              rounded-xl sm:h-80"
          >
            {imageSrc && (
              <Cropper
                image={imageSrc}
                crop={crop}
                zoom={zoom}
                aspect={aspect}
                cropShape={cropShape}
                showGrid={false}
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
          <input
            type="range"
            aria-label="Zoom"
            className="accent-primary w-full cursor-pointer"
            min={1}
            max={3}
            step={0.01}
            value={zoom}
            onChange={(e) => setZoom(Number(e.target.value))}
          />
          <div className="flex flex-row items-center justify-end gap-4">
            <ResponsiveDialogClose asChild>
              <Button
                variant="secondary"
                styles="py-2 px-3"
                buttonText="Cancel"
                type="button"
                disabled={isSaving}
              />
            </ResponsiveDialogClose>
            <Button
              styles="py-2 px-3"
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
