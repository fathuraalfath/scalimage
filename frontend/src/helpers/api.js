// ponytail: Minimal shared API & DOM download helpers; single source of truth for fetch & blob downloads.
export const BACKEND_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

/**
 * Uploads one or multiple files to the backend /api/upload endpoint.
 * @param {FileList | File[]} files
 * @returns {Promise<Array<{id: string, width: number, height: number, format: string}>>}
 */
export async function uploadImages(files) {
  const formData = new FormData();
  const fileList = Array.from(files);
  for (const file of fileList) {
    formData.append('images', file);
  }

  const res = await fetch(`${BACKEND_URL}/api/upload`, {
    method: 'POST',
    body: formData,
  });

  if (!res.ok) {
    const errorText = await res.text();
    throw new Error(errorText || 'Failed to upload files');
  }

  return res.json();
}

/**
 * Triggers a client-side browser file download from a Blob and safely revokes the URL.
 * @param {Blob} blob
 * @param {string} filename
 */
export function downloadBlob(blob, filename) {
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}
