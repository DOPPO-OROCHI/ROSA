import type { CSSProperties } from "react";
import { resolveCardAssetVariantSrc } from "../../lib/api";
import type { BattleCardInMatch } from "./types";

export type BattleCardViewerOrigin = {
  left: number;
  top: number;
  width: number;
  height: number;
};

type Props = {
  card: BattleCardInMatch;
  onClose: () => void;
  onExited: () => void;
  origin: BattleCardViewerOrigin | null;
  shellHeight: number;
  closing?: boolean;
};

function buildCombatInfo(card: BattleCardInMatch): string {
  const skillCd = card.base_cooldown ?? 0;
  return `SPLASH ${card.splash_radius} | SKILL CD ${skillCd}`;
}

function pickDescriptionFontSize(length: number) {
  if (length > 180) {
    return "0.42rem";
  }
  if (length > 140) {
    return "0.45rem";
  }
  if (length > 100) {
    return "0.48rem";
  }
  if (length > 70) {
    return "0.52rem";
  }
  return "0.56rem";
}

function pickCombatFontSize(length: number) {
  if (length > 48) {
    return "0.34rem";
  }
  if (length > 36) {
    return "0.38rem";
  }
  return "0.42rem";
}

export function BattleCardViewer({ card, onClose, onExited, origin, shellHeight, closing = false }: Props) {
  const isBattleCard = card.kind === "battle";
  const combatInfo = buildCombatInfo(card);
  const finalWidth = 154;
  const finalHeight = Math.round((finalWidth * 780) / 520);
  const finalLeft = 16;
  const finalTop = Math.max(122, shellHeight - finalHeight - 138);
  const dx = origin ? origin.left - finalLeft : 0;
  const dy = origin ? origin.top - finalTop : 0;
  const scaleX = origin ? origin.width / finalWidth : 0.82;
  const scaleY = origin ? origin.height / finalHeight : 0.82;
  const style = {
    left: `${finalLeft}px`,
    top: `${finalTop}px`,
    width: `${finalWidth}px`,
    ["--battle-viewer-dx" as string]: `${dx}px`,
    ["--battle-viewer-dy" as string]: `${dy}px`,
    ["--battle-viewer-scale-x" as string]: `${scaleX}`,
    ["--battle-viewer-scale-y" as string]: `${scaleY}`,
    ["--battle-viewer-description-font" as string]: pickDescriptionFontSize(card.description.length),
    ["--battle-viewer-combat-font" as string]: pickCombatFontSize(combatInfo.length),
  } as CSSProperties;

  return (
    <div className="battle-card-viewer-layer" onClick={onClose}>
      <aside
        className={`battle-card-viewer battle-card-viewer--left ${closing ? "battle-card-viewer--closing" : "battle-card-viewer--open"}`}
        style={style}
        onClick={(event) => event.stopPropagation()}
        onAnimationEnd={() => {
          if (closing) {
            onExited();
          }
        }}
        aria-label={`Просмотр карты ${card.name}`}
      >
        <button type="button" className="battle-card-viewer__close" onClick={onClose} aria-label={`Закрыть ${card.name}`}>
          x
        </button>

        <article className="battle-card-viewer__frame">
          <img
            className="battle-card-viewer__art"
            src={resolveCardAssetVariantSrc(card.kind as "battle" | "buff", card.template_id, "full_art")}
            alt={card.name}
            onError={(event) => {
              const target = event.currentTarget;
              if (target.dataset.fallbackApplied === "1") {
                return;
              }
              target.dataset.fallbackApplied = "1";
              target.src = resolveCardAssetVariantSrc(card.kind as "battle" | "buff", card.template_id, "view");
            }}
          />

          <div className="battle-card-viewer__anchor battle-card-viewer__anchor--mana">
            <span className="battle-card-viewer__stat">{card.mana_cost}</span>
          </div>
          {isBattleCard ? (
            <>
              <div className="battle-card-viewer__anchor battle-card-viewer__anchor--attack">
                <span className="battle-card-viewer__stat">{card.attack}</span>
              </div>
              <div className="battle-card-viewer__anchor battle-card-viewer__anchor--hp">
                <span className="battle-card-viewer__stat">{card.health_points}</span>
              </div>
            </>
          ) : null}

          <div className="battle-card-viewer__description">{card.description}</div>
          <div className="battle-card-viewer__combat">{combatInfo}</div>
          <div className="battle-card-viewer__name">{card.name}</div>
          <div className="battle-card-viewer__race">{card.card_type || card.kind.toUpperCase()}</div>
        </article>
      </aside>
    </div>
  );
}
