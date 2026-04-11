import { resolveCardAssetVariantSrc } from "../lib/api";
import type { BattleCard } from "../types";

type Props = {
  card: BattleCard;
  onClose: () => void;
};

function buildCombatInfo(card: BattleCard): string {
  const skillCooldown = card.skill?.base_cooldown ?? card.base_cooldown;
  return `ATK CD ${card.base_cooldown} | SPLASH ${card.splash_radius} | SKILL CD ${skillCooldown}`;
}

export function CardViewer({ card, onClose }: Props) {
  return (
    <div className="overlay">
      <section className="card-viewer surface">
        <button type="button" className="card-viewer__close" onClick={onClose} aria-label={`Close ${card.name} viewer`}>
          x
        </button>

        <article className="card-viewer__frame">
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

          <div className="card-viewer__description">{card.description}</div>
          <div className="card-viewer__combat">{buildCombatInfo(card)}</div>
          <div className="card-viewer__name">{card.name}</div>
          <div className="card-viewer__race">{card.card_type}</div>
        </article>
      </section>
    </div>
  );
}
