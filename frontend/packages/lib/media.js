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

export async function jsonRequest(url, method, body) {
  const response = await fetch(url, {
    method: method,
    headers: {
      Accept: "application/json",
      "content-type": "application/json",
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!response.ok) {
    invalidateLocalStorageAuth(response.status);
    const result = await response.json();
    throw new Error(result.error.message);
  }

  if (response.status === 204 || response.status === 202) {
    return null;
  }

  const result = await response.json();
  return result.data;
}

function wait(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export async function uploadImage(uploadUrl, file, crop) {
  const validationError = validateImageFile(file);
  if (validationError) {
    throw new Error(validationError);
  }

  const upload = await jsonRequest(uploadUrl, "POST", {
    content_type: file.type,
    size: file.size,
  });

  const storageResponse = await fetch(upload.upload_url, {
    method: upload.method,
    headers: upload.headers,
    body: file,
  });

  if (!storageResponse.ok) {
    throw new Error("Failed to upload the image, please try again");
  }

  await jsonRequest(
    `/api/v1/media/uploads/${upload.image_id}/complete`,
    "POST",
    { crop: crop ?? null }
  );

  const deadline = Date.now() + POLL_TIMEOUT_MS;

  while (Date.now() < deadline) {
    await wait(POLL_INTERVAL_MS);

    const image = await jsonRequest(
      `/api/v1/media/uploads/${upload.image_id}`,
      "GET"
    );

    if (image.status === "ready") {
      return image;
    }

    if (image.status === "failed") {
      throw new Error("The image could not be processed");
    }
  }

  throw new Error("Processing the image is taking longer than expected");
}

export async function deleteImage(url) {
  await jsonRequest(url, "DELETE");
}
