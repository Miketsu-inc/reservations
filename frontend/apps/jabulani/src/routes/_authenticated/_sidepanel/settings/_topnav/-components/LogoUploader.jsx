import {
  Camera01Icon,
  Delete02Icon,
  ImageAdd01Icon,
  ImageUploadIcon,
} from "@hugeicons/core-free-icons";
import { LoaderIcon } from "@reservations/assets";
import { Button, DeleteModal, Icon } from "@reservations/components";
import {
  ACCEPTED_IMAGE_TYPES,
  useToast,
  validateImageFile,
} from "@reservations/lib";
import { useRef, useState } from "react";

export default function LogoUploader({
  logoUrl,
  isUploading,
  onFileSelected,
  onRemove,
}) {
  const inputRef = useRef(null);
  const { showToast } = useToast();
  const [isDragging, setIsDragging] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);

  function handleFile(file) {
    const error = validateImageFile(file);
    if (error) {
      showToast({ message: error, variant: "error" });
      return;
    }

    onFileSelected(file);
  }

  function openFilePicker() {
    if (!isUploading) inputRef.current?.click();
  }

  return (
    <div className="flex flex-col gap-5 sm:flex-row sm:items-center">
      <button
        type="button"
        aria-label={logoUrl ? "Replace logo" : "Upload logo"}
        disabled={isUploading}
        onClick={openFilePicker}
        onDragOver={(event) => {
          event.preventDefault();
          setIsDragging(true);
        }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={(event) => {
          event.preventDefault();
          setIsDragging(false);
          const file = event.dataTransfer.files[0];
          if (file && !isUploading) handleFile(file);
        }}
        className={`${
          isDragging
            ? "border-border-primary bg-primary-subtle"
            : logoUrl
              ? "border-border"
              : "border-border-strong bg-bg-sunken hover:border-border-primary hover:bg-primary-subtle border-dashed"
        } group focus-visible:outline-primary relative flex size-32 shrink-0
        cursor-pointer items-center justify-center overflow-hidden rounded-2xl
        border transition-colors focus-visible:outline-2
        focus-visible:outline-offset-2 disabled:cursor-default`}
      >
        {logoUrl ? (
          <>
            <img src={logoUrl} alt="Logo" className="size-full object-cover" />
            {!isUploading && (
              <span
                className="bg-bg-backdrop text-fg-on-primary absolute inset-0
                  flex flex-col items-center justify-center gap-1 text-xs
                  font-medium opacity-0 transition-opacity
                  group-hover:opacity-100 group-focus-visible:opacity-100"
              >
                <Icon icon={Camera01Icon} styles="size-5" />
                Replace
              </span>
            )}
          </>
        ) : (
          <span
            className="text-fg-muted flex flex-col items-center gap-1.5 px-3
              text-xs"
          >
            <Icon icon={ImageAdd01Icon} styles="size-6" />
            Drop or click
          </span>
        )}
        {isUploading && (
          <span
            className="bg-bg-backdrop absolute inset-0 flex items-center
              justify-center"
          >
            <LoaderIcon styles="size-6 text-fg-on-primary" />
          </span>
        )}
      </button>

      <div className="flex flex-col gap-3">
        <div>
          <p className="text-fg text-sm font-medium">
            {logoUrl ? "Your logo" : "No logo yet"}
          </p>
          <p className="text-fg-muted mt-0.5 max-w-sm text-sm">
            A square image works best, at least 512 × 512 px. JPG or PNG, up to
            10 MB.
          </p>
        </div>
        <div className="flex flex-row gap-2">
          <Button
            type="button"
            variant="secondary"
            styles="px-3 py-1.5 text-sm gap-2"
            buttonText={logoUrl ? "Replace" : "Upload logo"}
            disabled={isUploading}
            onClick={openFilePicker}
          >
            <Icon icon={ImageUploadIcon} styles="size-4" />
          </Button>
          {logoUrl && (
            <DeleteModal
              open={isDeleteOpen}
              onOpenChange={setIsDeleteOpen}
              itemName="your logo"
              onDelete={onRemove}
              trigger={
                <Button
                  type="button"
                  variant="secondary"
                  styles="px-3 py-1.5 text-sm gap-2 text-fg-danger"
                  buttonText="Remove"
                  disabled={isUploading}
                >
                  <Icon icon={Delete02Icon} styles="size-4" />
                </Button>
              }
            />
          )}
        </div>
      </div>

      <input
        ref={inputRef}
        type="file"
        accept={ACCEPTED_IMAGE_TYPES.join(",")}
        className="hidden"
        onChange={(event) => {
          const file = event.target.files[0];
          // allows selecting the same file again
          event.target.value = "";
          if (file) handleFile(file);
        }}
      />
    </div>
  );
}
