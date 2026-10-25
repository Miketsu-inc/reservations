import {
  Delete02Icon,
  ImageAdd01Icon,
  ImageUploadIcon,
} from "@hugeicons/core-free-icons";
import { LoaderIcon } from "@reservations/assets";
import {
  Button,
  DeleteModal,
  Icon,
  ImageCropDialog,
} from "@reservations/components";
import { useCroppedImageUpload } from "@reservations/lib";
import { useState } from "react";

export default function LogoUpload({ img, alt, url, onChange, children }) {
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const upload = useCroppedImageUpload({ img, url, label: "logo", onChange });
  const [failedImg, setFailedImg] = useState(null);

  return (
    <div className="flex flex-row items-center gap-5">
      <div className="relative size-24 shrink-0" {...upload.dropProps}>
        {upload.pendingStyle ? (
          <div
            role="img"
            aria-label={alt}
            className="size-full rounded-2xl bg-no-repeat"
            style={upload.pendingStyle}
          />
        ) : img && img !== failedImg ? (
          <img
            className="size-full rounded-2xl object-cover"
            src={img}
            alt={alt}
            onError={() => setFailedImg(img)}
          />
        ) : (
          <EmptyState onClick={upload.openFilePicker} />
        )}
        {(upload.isSaving || upload.isDragging) && (
          <span
            className={`bg-bg-backdrop ring-primary pointer-events-none absolute
            inset-0 flex items-center justify-center rounded-2xl
            ${upload.isDragging ? "ring-2" : ""}`}
          >
            {upload.isSaving && (
              <LoaderIcon styles="size-6 text-fg-on-primary" />
            )}
          </span>
        )}
      </div>
      <div className="flex flex-col gap-3">
        {children}
        <div className="flex flex-row gap-2">
          {img ? (
            <>
              <Button
                type="button"
                variant="secondary"
                styles="py-1.5 px-3 text-sm gap-1.5"
                buttonText="Replace"
                disabled={upload.isSaving}
                onClick={upload.openFilePicker}
              >
                <Icon icon={ImageUploadIcon} styles="size-4" />
              </Button>
              <Button
                type="button"
                variant="secondary"
                styles="py-1.5 px-3 text-sm gap-1.5 text-fg-danger"
                buttonText="Remove"
                disabled={upload.isSaving}
                onClick={() => setIsDeleteOpen(true)}
              >
                <Icon icon={Delete02Icon} styles="size-4" />
              </Button>
            </>
          ) : (
            <Button
              type="button"
              variant="secondary"
              styles="py-1.5 px-3 text-sm gap-1.5"
              buttonText="Upload"
              disabled={upload.isSaving}
              onClick={upload.openFilePicker}
            >
              <Icon icon={ImageUploadIcon} styles="size-4" />
            </Button>
          )}
        </div>
      </div>
      <DeleteModal
        open={isDeleteOpen}
        onOpenChange={setIsDeleteOpen}
        itemName="the logo"
        onDelete={upload.remove}
      />
      <input {...upload.inputProps} />
      <ImageCropDialog
        key={upload.cropDialogProps.imageSrc}
        cropShape="rect"
        {...upload.cropDialogProps}
      />
    </div>
  );
}

function EmptyState({ onClick }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="border-border-strong bg-bg-sunken text-fg-muted
        hover:border-border-primary hover:bg-primary-subtle
        hover:text-fg-primary flex size-full cursor-pointer flex-col
        items-center justify-center gap-1 rounded-2xl border border-dashed
        text-xs font-medium transition-colors"
    >
      <Icon icon={ImageAdd01Icon} styles="size-6" />
      Upload logo
    </button>
  );
}
