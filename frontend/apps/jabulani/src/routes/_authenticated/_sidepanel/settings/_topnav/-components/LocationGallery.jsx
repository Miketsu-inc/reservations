import {
  ArrowLeft02Icon,
  ArrowRight02Icon,
  Delete02Icon,
  ImageAdd01Icon,
  MoreHorizontalIcon,
  PencilEdit02Icon,
  StarIcon,
} from "@hugeicons/core-free-icons";
import { LoaderIcon } from "@reservations/assets";
import {
  Button,
  DeleteModal,
  Icon,
  Input,
  Popover,
  PopoverContent,
  PopoverTrigger,
  ResponsiveDialog,
  ResponsiveDialogClose,
  ResponsiveDialogContent,
} from "@reservations/components";
import {
  ACCEPTED_IMAGE_TYPES,
  invalidateLocalStorageAuth,
  uploadImage,
  useToast,
  validateImageFile,
} from "@reservations/lib";
import { queryOptions, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRef, useState } from "react";

const ALT_TEXT_MAX_LENGTH = 150;

async function fetchLocationImages(merchantId, locationId) {
  const response = await fetch(
    `/api/v1/merchants/${merchantId}/locations/${locationId}/images`
  );
  const result = await response.json();

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  }

  return result.data;
}

function locationImagesQueryOptions(merchantId, locationId) {
  return queryOptions({
    queryKey: [merchantId, "location-images", locationId],
    queryFn: () => fetchLocationImages(merchantId, locationId),
  });
}

export default function LocationGallery({ merchantId, locationId, canEdit }) {
  const queryClient = useQueryClient();
  const { showToast } = useToast();
  const { data, isLoading } = useQuery(
    locationImagesQueryOptions(merchantId, locationId)
  );
  const inputRef = useRef(null);
  const [uploadingCount, setUploadingCount] = useState(0);
  const [isDragging, setIsDragging] = useState(false);
  const [editingImage, setEditingImage] = useState(null);
  const [deletingImage, setDeletingImage] = useState(null);

  const imagesUrl = `/api/v1/merchants/${merchantId}/locations/${locationId}/images`;
  const queryKey = locationImagesQueryOptions(merchantId, locationId).queryKey;

  if (isLoading || !data) {
    return <div className="bg-bg-sunken h-40 animate-pulse rounded-xl" />;
  }

  const images = data.images;
  const remaining = data.limit - images.length - uploadingCount;
  const canAdd = canEdit && remaining > 0;
  const isEmpty = images.length === 0 && uploadingCount === 0;
  const tileCount = images.length + uploadingCount + (canAdd ? 1 : 0);

  // the first tile is the large cover, the others fill the column next to it. A single
  // tile next to the cover is stretched to its height, so there is no empty space
  function tileStyles(position) {
    if (position === 0) {
      return tileCount >= 3
        ? "col-span-2 aspect-4/3 sm:row-span-2 sm:aspect-auto"
        : "col-span-2 aspect-4/3 sm:row-span-2";
    }
    if (position === 1 && tileCount === 2) {
      return "aspect-4/3 sm:row-span-2 sm:aspect-auto";
    }
    return "aspect-4/3";
  }

  function refresh() {
    return queryClient.invalidateQueries({ queryKey });
  }

  function openFilePicker() {
    inputRef.current?.click();
  }

  async function addFiles(fileList) {
    const files = Array.from(fileList);

    if (files.length > remaining) {
      showToast({
        message: `You can add up to ${data.limit} photos to this location with your subscription`,
        variant: "error",
      });
    }

    const validFiles = [];
    for (const file of files.slice(0, Math.max(remaining, 0))) {
      const error = validateImageFile(file);
      if (error) {
        showToast({ message: `${file.name}: ${error}`, variant: "error" });
      } else {
        validFiles.push(file);
      }
    }

    setUploadingCount((count) => count + validFiles.length);

    await Promise.all(
      validFiles.map(async (file) => {
        try {
          await uploadImage(imagesUrl, file);
          await refresh();
        } catch (error) {
          showToast({
            message: `${file.name}: ${error.message}`,
            variant: "error",
          });
        } finally {
          setUploadingCount((count) => count - 1);
        }
      })
    );
  }

  async function removeImage(image) {
    try {
      const response = await fetch(`${imagesUrl}/${image.image_id}`, {
        method: "DELETE",
        headers: {
          Accept: "application/json",
        },
      });

      if (!response.ok) {
        const result = await response.json();
        invalidateLocalStorageAuth(response.status);
        showToast({ message: result.error.message, variant: "error" });
      } else {
        await refresh();
      }
    } catch (error) {
      showToast({ message: error.message, variant: "error" });
    }
  }

  async function moveImage(index, newIndex) {
    const reordered = [...images];
    const [moved] = reordered.splice(index, 1);
    reordered.splice(newIndex, 0, moved);

    queryClient.setQueryData(queryKey, { ...data, images: reordered });

    try {
      const response = await fetch(`${imagesUrl}/order`, {
        method: "PUT",
        headers: {
          Accept: "application/json",
          "content-type": "application/json",
        },
        body: JSON.stringify({
          image_ids: reordered.map((image) => image.image_id),
        }),
      });

      if (!response.ok) {
        const result = await response.json();
        invalidateLocalStorageAuth(response.status);
        showToast({ message: result.error.message, variant: "error" });
        await refresh();
      }
    } catch (error) {
      showToast({ message: error.message, variant: "error" });
      await refresh();
    }
  }

  // throws so the description dialog stays open and shows the error
  async function saveAltText(image, altText) {
    const response = await fetch(`${imagesUrl}/${image.image_id}`, {
      method: "PATCH",
      headers: {
        Accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify({ alt_text: altText }),
    });

    if (!response.ok) {
      const result = await response.json();
      invalidateLocalStorageAuth(response.status);
      throw new Error(result.error.message);
    }

    queryClient.setQueryData(queryKey, (current) => ({
      ...current,
      images: current.images.map((img) =>
        img.image_id === image.image_id ? { ...img, alt_text: altText } : img
      ),
    }));
  }

  const dropHandlers = canAdd
    ? {
        onDragOver: (event) => {
          event.preventDefault();
          setIsDragging(true);
        },
        onDragLeave: (event) => {
          if (!event.currentTarget.contains(event.relatedTarget)) {
            setIsDragging(false);
          }
        },
        onDrop: (event) => {
          event.preventDefault();
          setIsDragging(false);
          if (event.dataTransfer.files.length) {
            addFiles(event.dataTransfer.files);
          }
        },
      }
    : {};

  return (
    <div className="flex flex-col gap-4" {...dropHandlers}>
      <div className="flex flex-row items-start justify-between gap-4">
        <div>
          <h3 className="text-fg text-sm font-semibold">Photos</h3>
          <p className="text-fg-muted mt-0.5 text-sm">
            Shown on your booking page. The first photo is the cover.
          </p>
        </div>
        <span
          className="bg-bg-sunken text-fg-muted shrink-0 rounded-full px-2.5
            py-1 text-xs font-medium tabular-nums"
        >
          {images.length} of {data.limit}
        </span>
      </div>

      {isEmpty ? (
        canEdit ? (
          <button
            type="button"
            onClick={openFilePicker}
            className={`${
              isDragging
                ? "border-border-primary bg-primary-subtle"
                : `border-border-strong bg-bg-sunken hover:border-border-primary
                  hover:bg-primary-subtle`
              } flex h-48 w-full cursor-pointer flex-col items-center
              justify-center gap-2 rounded-xl border border-dashed px-4
              text-center transition-colors`}
          >
            <span className="bg-bg-raised shadow-card rounded-full p-2.5">
              <Icon icon={ImageAdd01Icon} styles="size-5 text-fg-primary" />
            </span>
            <span className="text-fg text-sm font-medium">
              Add photos of this location
            </span>
            <span className="text-fg-muted text-xs">
              Drag and drop or click to upload. JPG or PNG, up to 10 MB each.
            </span>
          </button>
        ) : (
          <p className="text-fg-muted text-sm">No photos yet.</p>
        )
      ) : (
        <div
          className={`${isDragging ? "outline-primary outline-2 outline-offset-4 outline-dashed" : ""}
            grid max-w-3xl grid-cols-2 gap-3 rounded-xl sm:grid-cols-3`}
        >
          {images.map((image, index) => (
            <PhotoTile
              key={image.image_id}
              styles={tileStyles(index)}
              image={image}
              index={index}
              count={images.length}
              canEdit={canEdit}
              onMove={(newIndex) => moveImage(index, newIndex)}
              onEditDescription={() => setEditingImage(image)}
              onDelete={() => setDeletingImage(image)}
            />
          ))}
          {Array.from({ length: uploadingCount }, (_, i) => (
            <div
              key={`uploading-${i}`}
              className={`${tileStyles(images.length + i)} bg-bg-sunken
                text-fg-muted flex flex-col items-center justify-center gap-2
                rounded-xl text-xs`}
            >
              <LoaderIcon styles="size-5 text-fg-primary" />
              Uploading…
            </div>
          ))}
          {canAdd && (
            <button
              type="button"
              onClick={openFilePicker}
              className={`${tileStyles(tileCount - 1)} border-border-strong
                text-fg-muted hover:border-border-primary
                hover:bg-primary-subtle hover:text-fg-primary flex
                cursor-pointer flex-col items-center justify-center gap-1.5
                rounded-xl border border-dashed text-xs font-medium
                transition-colors`}
            >
              <Icon icon={ImageAdd01Icon} styles="size-5" />
              Add photos
            </button>
          )}
        </div>
      )}

      {canEdit && !isEmpty && remaining <= 0 && uploadingCount === 0 && (
        <p className="text-fg-muted text-xs">
          You have reached the photo limit of your subscription for this
          location.
        </p>
      )}

      <input
        ref={inputRef}
        type="file"
        multiple
        accept={ACCEPTED_IMAGE_TYPES.join(",")}
        className="hidden"
        onChange={(event) => {
          const files = event.target.files;
          if (files?.length) addFiles(files);
          // allows selecting the same files again
          event.target.value = "";
        }}
      />

      <DescriptionDialog
        key={editingImage?.image_id}
        image={editingImage}
        onClose={() => setEditingImage(null)}
        onSave={saveAltText}
      />

      <DeleteModal
        open={deletingImage !== null}
        onOpenChange={(open) => !open && setDeletingImage(null)}
        itemName="this photo"
        onDelete={() => removeImage(deletingImage)}
      />
    </div>
  );
}

function PhotoTile({
  styles,
  image,
  index,
  count,
  canEdit,
  onMove,
  onEditDescription,
  onDelete,
}) {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const isCover = index === 0;

  function select(action) {
    setIsMenuOpen(false);
    action();
  }

  return (
    <figure
      className={`${styles} group bg-bg-sunken relative overflow-hidden
        rounded-xl`}
    >
      <img
        src={isCover ? image.url : image.thumbnail_url}
        alt={image.alt_text}
        className="size-full object-cover"
        loading="lazy"
      />

      {isCover && (
        <span
          className="bg-bg-overlay text-fg shadow-card absolute top-2 left-2
            flex items-center gap-1 rounded-full px-2 py-0.5 text-xs
            font-medium"
        >
          <Icon icon={StarIcon} styles="size-3.5 text-fg-warning" />
          Cover
        </span>
      )}

      {image.alt_text && (
        <figcaption
          className="absolute inset-x-0 bottom-0 truncate bg-linear-to-t
            from-black/70 to-transparent px-3 pt-6 pb-2 text-xs text-white"
        >
          {image.alt_text}
        </figcaption>
      )}

      {canEdit && (
        <Popover open={isMenuOpen} onOpenChange={setIsMenuOpen}>
          <PopoverTrigger
            aria-label="Photo options"
            className="bg-bg-overlay text-fg shadow-card hover:text-fg-primary
              absolute top-2 right-2 cursor-pointer rounded-full p-1.5
              transition-opacity group-hover:opacity-100
              focus-visible:opacity-100 data-popup-open:opacity-100
              lg:opacity-0"
          >
            <Icon icon={MoreHorizontalIcon} styles="size-4" />
          </PopoverTrigger>
          <PopoverContent align="end">
            <div
              className="*:hover:bg-hover space-y-0.5 *:flex *:w-full
                *:cursor-pointer *:flex-row *:items-center *:gap-2 *:rounded-lg
                *:p-2"
            >
              {!isCover && (
                <button type="button" onClick={() => select(() => onMove(0))}>
                  <Icon icon={StarIcon} styles="size-4 text-fg-muted" />
                  Set as cover
                </button>
              )}
              {index > 0 && (
                <button
                  type="button"
                  onClick={() => select(() => onMove(index - 1))}
                >
                  <Icon icon={ArrowLeft02Icon} styles="size-4 text-fg-muted" />
                  Move earlier
                </button>
              )}
              {index < count - 1 && (
                <button
                  type="button"
                  onClick={() => select(() => onMove(index + 1))}
                >
                  <Icon icon={ArrowRight02Icon} styles="size-4 text-fg-muted" />
                  Move later
                </button>
              )}
              <button type="button" onClick={() => select(onEditDescription)}>
                <Icon icon={PencilEdit02Icon} styles="size-4 text-fg-muted" />
                {image.alt_text ? "Edit description" : "Add description"}
              </button>
              <button
                type="button"
                className="text-fg-danger"
                onClick={() => select(onDelete)}
              >
                <Icon icon={Delete02Icon} styles="size-4" />
                Delete photo
              </button>
            </div>
          </PopoverContent>
        </Popover>
      )}
    </figure>
  );
}

function DescriptionDialog({ image, onClose, onSave }) {
  const [altText, setAltText] = useState(image?.alt_text ?? "");
  const [isSaving, setIsSaving] = useState(false);
  const { showToast } = useToast();

  async function save() {
    setIsSaving(true);

    try {
      await onSave(image, altText.trim());
      onClose();
    } catch (error) {
      showToast({ message: error.message, variant: "error" });
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <ResponsiveDialog
      open={image !== null}
      onOpenChange={(open) => !open && !isSaving && onClose()}
    >
      <ResponsiveDialogContent styles="sm:w-md">
        {image && (
          <form
            className="flex flex-col gap-4 p-4"
            onSubmit={(event) => {
              event.preventDefault();
              save();
            }}
          >
            <div>
              <p className="text-lg font-semibold">Photo description</p>
              <p className="text-fg-muted mt-0.5 text-sm">
                Describe what the photo shows. It helps customers using screen
                readers and search engines.
              </p>
            </div>
            <img
              src={image.thumbnail_url}
              alt=""
              className="bg-bg-sunken aspect-video w-full rounded-lg
                object-cover"
            />
            <div>
              <Input
                type="text"
                id="location_image_alt_text"
                name="alt_text"
                labelText="Description"
                placeholder="e.g. Reception area with waiting chairs"
                required={false}
                maxLength={ALT_TEXT_MAX_LENGTH}
                value={altText}
                inputData={({ value }) => setAltText(value)}
              />
              <p className="text-fg-muted mt-1 text-right text-xs tabular-nums">
                {altText.length} / {ALT_TEXT_MAX_LENGTH}
              </p>
            </div>
            <div className="flex flex-row items-center justify-end gap-3">
              <ResponsiveDialogClose asChild>
                <Button
                  type="button"
                  variant="secondary"
                  styles="py-2 px-3 hidden lg:block"
                  buttonText="Cancel"
                  disabled={isSaving}
                />
              </ResponsiveDialogClose>
              <Button
                type="submit"
                styles="py-2 px-3 w-full lg:w-auto"
                buttonText="Save"
                isLoading={isSaving}
                disabled={altText.trim() === image.alt_text}
              />
            </div>
          </form>
        )}
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}
