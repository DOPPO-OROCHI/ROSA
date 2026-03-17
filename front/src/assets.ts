const tones = ["crimson", "iron", "ash", "bone", "toxic"] as const;

export type AssetTone = (typeof tones)[number];

const assetBaseUrl = (import.meta.env.VITE_ASSET_BASE_URL || "/assets").replace(/\/+$/, "");
const assetVersion = (import.meta.env.VITE_ASSET_VERSION || "1").trim();

function hash(input: string): number {
  let value = 0;
  for (let index = 0; index < input.length; index += 1) {
    value = (value * 31 + input.charCodeAt(index)) >>> 0;
  }
  return value;
}

function normalizeKey(key: string): string {
  return key.replace(/^\/+/, "").replace(/\/+/g, "/");
}

function buildAssetUrl(key: string, extension: string): string {
  const base = `${assetBaseUrl}/${normalizeKey(key)}.${extension}`;
  return assetVersion ? `${base}?v=${encodeURIComponent(assetVersion)}` : base;
}

export function getAssetTone(key: string): AssetTone {
  if (!key) {
    return "ash";
  }
  return tones[hash(key) % tones.length];
}

export function resolveAssetLabel(key: string): string {
  if (!key) {
    return "No Signal";
  }

  const parts = key.split("/").filter(Boolean);
  const raw = parts[parts.length - 2] || parts[parts.length - 1] || key;
  return raw.replace(/_/g, " ").slice(0, 18);
}

export function resolveHeroImageKey(heroCode: string): string {
  return `heroes/${heroCode}/image`;
}

export function resolveBattleCardImageKey(templateId: string): string {
  return `cards/battle/${templateId}/image`;
}

export function resolveBuffCardImageKey(templateId: string): string {
  return `cards/buff/${templateId}/image`;
}

export function resolveImageSrc(key?: string): string {
  if (!key) {
    return buildAssetUrl("placeholders/card_image", "svg");
  }
  return buildAssetUrl(key, "png");
}

export function resolveCardFallbackSrc(): string {
  return buildAssetUrl("placeholders/card_image", "svg");
}

export function resolveHeroFallbackSrc(): string {
  return buildAssetUrl("placeholders/hero_image", "svg");
}

export function resolveBoardBackgroundSrc(): string {
  return buildAssetUrl("placeholders/board_bg", "svg");
}
