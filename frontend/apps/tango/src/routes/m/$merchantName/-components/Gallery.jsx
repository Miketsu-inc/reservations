// has to match the backend's small variant size
const THUMBNAIL_MAX_SIZE = 640;

function thumbnailWidth(image) {
  const scale = Math.min(
    1,
    THUMBNAIL_MAX_SIZE / Math.max(image.width, image.height)
  );
  return Math.round(image.width * scale);
}

function GalleryImage({ image, sizes, alt, priority }) {
  return (
    <img
      className="absolute inset-0 size-full object-cover"
      src={image.url}
      srcSet={`${image.thumbnail_url} ${thumbnailWidth(image)}w, ${image.url} ${image.width}w`}
      sizes={sizes}
      width={image.width}
      height={image.height}
      alt={image.alt_text || alt}
      loading={priority ? "eager" : "lazy"}
      fetchPriority={priority ? "high" : "auto"}
    />
  );
}

export default function Gallery({ images, merchantName, logoUrl }) {
  if (images.length === 0) {
    return (
      <div
        className="bg-bg-sunken flex h-62.5 w-full items-center justify-center
          lg:h-125 lg:rounded-2xl"
      >
        {logoUrl ? (
          <img
            className="size-28 rounded-2xl object-cover lg:size-40"
            src={logoUrl}
            alt={`${merchantName} logo`}
          />
        ) : (
          <span className="text-fg-muted text-5xl font-semibold lg:text-7xl">
            {merchantName.slice(0, 2).toUpperCase()}
          </span>
        )}
      </div>
    );
  }

  const [cover, ...rest] = images;
  const side = rest.slice(0, 2);

  return (
    <div
      className={`${side.length > 0 ? "lg:grid-cols-[2fr_1fr]" : ""} grid w-full
        grid-cols-1 gap-2 lg:h-125 lg:overflow-hidden lg:rounded-2xl`}
    >
      <div className="relative h-62.5 w-full lg:h-full">
        <GalleryImage
          image={cover}
          sizes={side.length > 0 ? "(min-width: 1024px) 66vw, 100vw" : "100vw"}
          alt={merchantName}
          priority
        />
      </div>
      {side.length > 0 && (
        <div className="hidden h-full w-full flex-col gap-2 lg:flex">
          {side.map((image) => (
            <div key={image.url} className="relative flex-1">
              <GalleryImage image={image} sizes="33vw" alt={merchantName} />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
