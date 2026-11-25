import { useState } from "react";
import MerchantLogo from "./MerchantLogo";

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
        <MerchantLogo
          styles="size-28 rounded-2xl! lg:size-40"
          src={logoUrl}
          alt={`${merchantName} logo`}
          fallback={
            <span className="text-fg-muted text-5xl font-semibold lg:text-7xl">
              {merchantName.slice(0, 2).toUpperCase()}
            </span>
          }
        />
      </div>
    );
  }

  const side = images.slice(1, 3);

  return (
    <div
      className={`${side.length > 0 ? "lg:grid-cols-[2fr_1fr]" : ""} grid w-full
        grid-cols-1 gap-2 lg:h-125 lg:overflow-hidden lg:rounded-2xl`}
    >
      <Carousel
        images={images}
        merchantName={merchantName}
        hasSide={side.length > 0}
      />
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

function Carousel({ images, merchantName, hasSide }) {
  const [current, setCurrent] = useState(0);

  function handleScroll(e) {
    const el = e.currentTarget;
    if (el.clientWidth === 0) return;
    setCurrent(Math.round(el.scrollLeft / el.clientWidth));
  }

  return (
    <div className="relative h-68 w-full lg:h-full">
      <div
        className="flex size-full snap-x snap-mandatory scrollbar-none
          overflow-x-auto overscroll-x-contain lg:overflow-hidden"
        onScroll={handleScroll}
      >
        {images.map((image, i) => (
          <div
            key={image.url}
            className={`relative h-full w-full shrink-0 snap-center
            ${i > 0 ? "lg:hidden" : ""}`}
          >
            <GalleryImage
              image={image}
              sizes={
                i === 0 && hasSide ? "(min-width: 1024px) 66vw, 100vw" : "100vw"
              }
              alt={merchantName}
              priority={i === 0}
            />
          </div>
        ))}
      </div>
      {images.length > 1 && (
        <span
          className="pointer-events-none absolute right-5 bottom-9 rounded-full
            bg-black/60 px-2.5 py-1 text-sm font-medium text-white tabular-nums
            lg:hidden"
        >
          {current + 1}/{images.length}
        </span>
      )}
    </div>
  );
}
