export async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
    ...init,
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Request failed: ${response.status}`);
  }

  return response.json() as Promise<T>;
}

export function resolveImageSrc(key?: string, fallback = "/assets/placeholders/hero_image.svg"): string {
  if (!key) {
    return fallback;
  }
  return `/assets/${key.replace(/^\/+/, "").replace(/\/+/g, "/")}.png`;
}

export function resolveCardImageSrc(
  kind?: "battle" | "buff",
  templateId?: string,
  key?: string,
): string {
  if (kind && templateId) {
    return `/assets/cards/${kind}/${templateId}/image.png`;
  }
  return resolveImageSrc(key, "/assets/placeholders/card_image.svg");
}
