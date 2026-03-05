const tones = ["crimson", "iron", "ash", "bone", "toxic"] as const;

export type AssetTone = (typeof tones)[number];

function hash(input: string): number {
  let value = 0;
  for (let index = 0; index < input.length; index += 1) {
    value = (value * 31 + input.charCodeAt(index)) >>> 0;
  }
  return value;
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
