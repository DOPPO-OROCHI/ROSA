import { useEffect, useState } from "react";
import { resolveCardFallbackSrc, resolveImageSrc } from "../assets";
import type { GameCardData } from "./GameCard";
import styles from "./inventar_cardsview.module.css";

type Props = {
  data: GameCardData;
};

const CARD_WIDTH = 400;
const CARD_HEIGHT = 600;
const DESC_LEFT = 56;
const DESC_TOP = 342;
const DESC_WIDTH = 288;
const DESC_LINE_HEIGHT = 22;
const DESC_MAX_LINES = 2;

const previewCache = new Map<string, string>();

function cacheKey(data: GameCardData): string {
  return [
    data.kind,
    data.imageKey,
    data.name,
    data.description,
    data.mana ?? 0,
    data.attack ?? 0,
    data.hp ?? 0,
    data.buffValue ?? 0,
    data.duration ?? 0,
  ].join("|");
}

function loadImage(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image();
    image.decoding = "async";
    image.onload = () => resolve(image);
    image.onerror = () => reject(new Error(`Failed to load image: ${src}`));
    image.src = src;
  });
}

function drawStat(ctx: CanvasRenderingContext2D, value: number, x: number, y: number) {
  const text = String(value);
  ctx.save();
  ctx.font = '900 56px "Bebas Neue", Impact, "Arial Narrow", sans-serif';
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.lineJoin = "round";
  ctx.strokeStyle = "rgba(0, 0, 0, 0.96)";
  ctx.lineWidth = 10;
  ctx.fillStyle = "#fff1cc";
  ctx.shadowColor = "rgba(0, 0, 0, 0.9)";
  ctx.shadowBlur = 18;
  ctx.strokeText(text, x, y);
  ctx.fillText(text, x, y);
  ctx.restore();
}

function wrapText(ctx: CanvasRenderingContext2D, text: string, maxWidth: number, maxLines: number): string[] {
  const words = text.trim().split(/\s+/).filter(Boolean);
  const lines: string[] = [];
  let current = "";

  for (const word of words) {
    const next = current ? `${current} ${word}` : word;
    if (ctx.measureText(next).width <= maxWidth) {
      current = next;
      continue;
    }
    if (current) {
      lines.push(current);
    }
    current = word;
    if (lines.length === maxLines - 1) {
      break;
    }
  }

  if (lines.length < maxLines && current) {
    lines.push(current);
  }

  if (lines.length === 0) {
    return [];
  }

  const consumed = lines.join(" ");
  if (consumed.length < text.trim().length) {
    const last = lines[lines.length - 1].replace(/[. ]+$/, "");
    lines[lines.length - 1] = `${last}...`;
  }

  return lines;
}

function drawDescription(ctx: CanvasRenderingContext2D, text: string) {
  if (!text.trim()) {
    return;
  }

  ctx.save();
  ctx.font = '700 18px "Trebuchet MS", Arial, sans-serif';
  ctx.textAlign = "center";
  ctx.textBaseline = "top";
  ctx.fillStyle = "#ddd1b9";
  ctx.shadowColor = "rgba(0, 0, 0, 0.82)";
  ctx.shadowBlur = 10;

  const lines = wrapText(ctx, text, DESC_WIDTH, DESC_MAX_LINES);
  lines.forEach((line, index) => {
    ctx.fillText(line, DESC_LEFT + DESC_WIDTH / 2, DESC_TOP + index * DESC_LINE_HEIGHT);
  });
  ctx.restore();
}

async function buildPreview(data: GameCardData): Promise<string> {
  const imageSrc = resolveImageSrc(data.imageKey);
  const fallbackSrc = resolveCardFallbackSrc();
  const baseImage = await loadImage(imageSrc).catch(() => loadImage(fallbackSrc));
  const canvas = document.createElement("canvas");
  const ratio = window.devicePixelRatio || 1;

  canvas.width = CARD_WIDTH * ratio;
  canvas.height = CARD_HEIGHT * ratio;

  const ctx = canvas.getContext("2d");
  if (!ctx) {
    return fallbackSrc;
  }

  ctx.scale(ratio, ratio);
  ctx.clearRect(0, 0, CARD_WIDTH, CARD_HEIGHT);
  ctx.drawImage(baseImage, 0, 0, CARD_WIDTH, CARD_HEIGHT);

  const attackValue = data.kind === "battle" ? data.attack ?? 0 : data.buffValue ?? 0;
  const hpValue = data.kind === "battle" ? data.hp ?? 0 : data.duration ?? 0;

  drawStat(ctx, data.mana ?? 0, 48, 48);
  drawStat(ctx, attackValue, 51, 514);
  drawStat(ctx, hpValue, 349, 514);
  drawDescription(ctx, data.description);

  return canvas.toDataURL("image/png");
}

export function CatalogCardPreview({ data }: Props) {
  const key = cacheKey(data);
  const [src, setSrc] = useState<string>(() => previewCache.get(key) ?? "");

  useEffect(() => {
    let cancelled = false;

    if (previewCache.has(key)) {
      setSrc(previewCache.get(key) ?? "");
      return () => {
        cancelled = true;
      };
    }

    buildPreview(data).then((nextSrc) => {
      if (cancelled) {
        return;
      }
      previewCache.set(key, nextSrc);
      setSrc(nextSrc);
    });

    return () => {
      cancelled = true;
    };
  }, [data, key]);

  if (!src) {
    return <div className={styles.catalogPreviewSkeleton} />;
  }

  return <img className={styles.catalogPreviewImage} src={src} alt={data.name} loading="lazy" />;
}
