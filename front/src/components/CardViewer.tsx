import { resolveCardAssetVariantSrc } from "../lib/api";
import type { BattleCard } from "../types";
import { useEffect, useState } from "react";

type Props = {
  card: BattleCard;
  canGoBack: boolean;
  canGoForward: boolean;
  onBack: () => void;
  onForward: () => void;
  onClose: () => void;
};

function buildCombatInfo(card: BattleCard): string {
  const skillCooldown = card.skill?.base_cooldown ?? card.base_cooldown;
  return `ATK CD ${card.base_cooldown} | SPLASH ${card.splash_radius} | SKILL CD ${skillCooldown}`;
}

function getNameFitClass(name: string): string {
  const normalizedLength = name.replace(/\s+/g, "").length;
  if (normalizedLength >= 20) {
    return "card-viewer__name--dense";
  }
  if (normalizedLength >= 15) {
    return "card-viewer__name--compact";
  }
  return "";
}

export function CardViewer({
  card,
  canGoBack,
  canGoForward,
  onBack,
  onForward,
  onClose,
}: Props) {
  const [passiveOpen, setPassiveOpen] = useState(false);
  const [slideDirection, setSlideDirection] = useState<"back" | "forward" | null>(null);
  const passive = card.passive?.code ? card.passive : null;

  useEffect(() => {
    setPassiveOpen(false);
  }, [card.template_id]);

  function shift(direction: "back" | "forward") {
    setPassiveOpen(false);
    setSlideDirection(direction);
    if (direction === "back") {
      onBack();
      return;
    }
    onForward();
  }

  return (
    <div className="overlay">
      <section className="card-viewer surface">
        <button type="button" className="card-viewer__close" onClick={onClose} aria-label={`Close ${card.name} viewer`}>
          x
        </button>
        <button
          type="button"
          className="card-viewer__nav card-viewer__nav--back"
          onClick={() => shift("back")}
          disabled={!canGoBack}
          aria-label="Previous card"
        >
          {"<"}
        </button>
        <button
          type="button"
          className="card-viewer__nav card-viewer__nav--forward"
          onClick={() => shift("forward")}
          disabled={!canGoForward}
          aria-label="Next card"
        >
          {">"}
        </button>

        <article
          key={card.template_id}
          className={`card-viewer__frame ${slideDirection ? `card-viewer__frame--${slideDirection}` : ""}`}
          onAnimationEnd={() => setSlideDirection(null)}
        >
          <img
            className="card-viewer__art"
            src={resolveCardAssetVariantSrc(card.kind, card.template_id, "full_art")}
            alt={card.name}
            onError={(event) => {
              const target = event.currentTarget;
              if (target.dataset.fallbackApplied === "1") {
                return;
              }
              target.dataset.fallbackApplied = "1";
              target.src = resolveCardAssetVariantSrc(card.kind, card.template_id, "view");
            }}
          />

          <div className="card-viewer__anchor card-viewer__anchor--mana">
            <span className="card-viewer__stat">{card.mana_cost}</span>
          </div>
          <div className="card-viewer__anchor card-viewer__anchor--attack">
            <span className="card-viewer__stat">{card.attack}</span>
          </div>
          <div className="card-viewer__anchor card-viewer__anchor--hp">
            <span className="card-viewer__stat">{card.health_points}</span>
          </div>

          <div className={`card-viewer__description ${passiveOpen ? "card-viewer__description--passive" : ""}`}>
            {passiveOpen && passive ? passive.description : card.description}
          </div>
          <div className="card-viewer__combat">{buildCombatInfo(card)}</div>
          <div className={`card-viewer__name ${getNameFitClass(card.name)}`}>{card.name}</div>
          <div className="card-viewer__race">{card.card_type}</div>
          {passive ? (
            <button
              type="button"
              className={`card-viewer__passive-hotspot ${passiveOpen ? "card-viewer__passive-hotspot--active" : ""}`}
              onClick={(event) => {
                event.stopPropagation();
                setPassiveOpen((value) => !value);
              }}
              aria-label={`Show passive ${passive.name}`}
            >
              {passive.name}
            </button>
          ) : null}
        </article>
      </section>
    </div>
  );
}
