import {
  Camera01Icon,
  Delete02Icon,
  ImageUploadIcon,
} from "@hugeicons/core-free-icons";
import { LoaderIcon } from "@reservations/assets";
import { useCroppedImageUpload } from "@reservations/lib";
import { useState } from "react";
import Avatar from "./Avatar.jsx";
import DeleteModal from "./DeleteModal.jsx";
import Icon from "./Icon.jsx";
import ImageCropDialog from "./ImageCropDialog.jsx";
import { Popover, PopoverContent, PopoverTrigger } from "./Popover.jsx";

export default function AvatarUpload({
  img,
  initials,
  alt,
  styles = "",
  url,
  canRemove = !!img,
  onChange,
}) {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const upload = useCroppedImageUpload({
    img,
    url,
    label: "picture",
    onChange,
  });

  return (
    <div className="relative flex w-fit shrink-0" {...upload.dropProps}>
      {upload.pendingStyle ? (
        <div
          role="img"
          aria-label={alt}
          className={`${styles} size-16 rounded-full bg-no-repeat`}
          style={upload.pendingStyle}
        />
      ) : (
        <Avatar styles={styles} img={img} initials={initials} alt={alt} />
      )}
      {(upload.isSaving || upload.isDragging) && (
        <span
          className={`bg-bg-backdrop ring-primary absolute inset-0 flex
          items-center justify-center rounded-full
          ${upload.isDragging ? "ring-2" : ""}`}
        >
          {upload.isSaving && <LoaderIcon styles="size-6 text-fg-on-primary" />}
        </span>
      )}
      {canRemove ? (
        <Popover open={isMenuOpen} onOpenChange={setIsMenuOpen}>
          <PopoverTrigger aria-label="Change picture">
            <CameraBadge />
          </PopoverTrigger>
          <PopoverContent align="start">
            <div
              className="*:hover:bg-hover space-y-1 text-sm *:flex *:w-full
                *:cursor-pointer *:flex-row *:items-center *:gap-2 *:rounded-lg
                *:p-2"
            >
              <button
                type="button"
                onClick={() => {
                  setIsMenuOpen(false);
                  upload.openFilePicker();
                }}
              >
                <Icon icon={ImageUploadIcon} styles="size-4.5 text-fg-muted" />
                Upload new picture
              </button>
              <button
                type="button"
                className="text-fg-danger"
                onClick={() => {
                  setIsMenuOpen(false);
                  setIsDeleteOpen(true);
                }}
              >
                <Icon icon={Delete02Icon} styles="size-4.5" />
                Remove picture
              </button>
            </div>
          </PopoverContent>
        </Popover>
      ) : (
        <button
          type="button"
          aria-label="Upload picture"
          onClick={upload.openFilePicker}
        >
          <CameraBadge />
        </button>
      )}
      <DeleteModal
        open={isDeleteOpen}
        onOpenChange={setIsDeleteOpen}
        itemName="the picture"
        onDelete={upload.remove}
      />
      <input {...upload.inputProps} />
      <ImageCropDialog
        key={upload.cropDialogProps.imageSrc}
        cropShape="round"
        {...upload.cropDialogProps}
      />
    </div>
  );
}

function CameraBadge() {
  return (
    <span
      className="bg-bg-raised border-border shadow-card absolute -right-1
        -bottom-1 cursor-pointer overflow-hidden rounded-full border"
    >
      <span className="hover:bg-hover flex items-center justify-center p-1.5">
        <Icon icon={Camera01Icon} styles="size-4 text-fg" />
      </span>
    </span>
  );
}
