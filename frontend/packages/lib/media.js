import { invalidateLocalStorageAuth } from "./lib.js";

// 10 MB, has to match the backend limit
export const MAX_IMAGE_SIZE = 10 * 1024 * 1024;
export const ACCEPTED_IMAGE_TYPES = ["image/jpeg", "image/png"];

const POLL_INTERVAL_MS = 1000;
const POLL_TIMEOUT_MS = 60 * 1000;

export function validateImageFile(file) {
  if (!ACCEPTED_IMAGE_TYPES.includes(file.type)) {
    return "Only JPEG and PNG images are supported";
  }

  if (file.size > MAX_IMAGE_SIZE) {
    return "The image must be smaller than 10 MB";
  }

  return null;
}

function wait(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function createUpload(uploadUrl, file) {
  const response = await fetch(uploadUrl, {
    method: "POST",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
    body: JSON.stringify({ content_type: file.type, size: file.size }),
  });

  const result = await response.json();

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  }

  return result.data;
}

async function sendToStorage(upload, file) {
  const response = await fetch(upload.upload_url, {
    method: upload.method,
    headers: upload.headers,
    body: file,
  });

  if (!response.ok) {
    throw new Error("Failed to upload the image, please try again");
  }
}

async function completeUpload(imageId, crop) {
  const response = await fetch(`/api/v1/images/uploads/${imageId}/complete`, {
    method: "POST",
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
    body: JSON.stringify({ crop: crop ?? null }),
  });

  if (!response.ok) {
    const result = await response.json();
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  }
}

async function fetchUpload(imageId) {
  const response = await fetch(`/api/v1/images/uploads/${imageId}`, {
    method: "GET",
    headers: {
      Accept: "application/json",
    },
  });

  const result = await response.json();

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    throw result.error;
  }

  return result.data;
}

export async function uploadImage(uploadUrl, file, crop) {
  const validationError = validateImageFile(file);
  if (validationError) {
    throw new Error(validationError);
  }

  const upload = await createUpload(uploadUrl, file);
  await sendToStorage(upload, file);
  await completeUpload(upload.image_id, crop);

  const deadline = Date.now() + POLL_TIMEOUT_MS;

  while (Date.now() < deadline) {
    await wait(POLL_INTERVAL_MS);

    const image = await fetchUpload(upload.image_id);

    if (image.status === "ready") {
      return image;
    }

    if (image.status === "failed") {
      throw new Error("The image could not be processed");
    }
  }

  throw new Error("Processing the image is taking longer than expected");
}
