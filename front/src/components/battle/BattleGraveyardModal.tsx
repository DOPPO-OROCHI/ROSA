import type { MouseEvent } from "react";
import { resolveCardImageSrc } from "../../lib/api";
import type { BattleCardInMatch } from "./types";

type GraveStack = {
  card: BattleCardInMatch;
  count: number;
};

type Props = {
  cards: BattleCardInMatch[];
  loading?: boolean;
  error?: string;
  onClose: () => void;
  onOpenCard?: (card: BattleCardInMatch, originRect: DOMRect) => void;
};

function groupDiscard(cards: BattleCardInMatch[]): GraveStack[] {
  const grouped = cards.reduce<GraveStack[]>((acc, card) => {
    const existing = acc.find((entry) => entry.card.template_id === card.template_id);
    if (existing) {
      existing.count += 1;
      return acc;
    }

    acc.push({ card, count: 1 });
    return acc;
  }, []);

  return grouped.reverse();
}

export function BattleGraveyardModal({ cards, loading = false, error = "", onClose, onOpenCard }: Props) {
  const stacks = groupDiscard(cards);

  return (
    <div className="battle-graveyard-modal-layer" onClick={onClose}>
      <section className="battle-graveyard-modal" onClick={(event) => event.stopPropagation()} aria-label="Graveyard">
        <div className="battle-graveyard-modal__header">
          <h2 className="battle-graveyard-modal__title">ПАДШИЕ</h2>
          <button type="button" className="battle-graveyard-modal__close" onClick={onClose} aria-label="Close graveyard">
            x
          </button>
        </div>

        {loading && stacks.length === 0 ? (
          <p className="battle-graveyard-modal__empty">Loading...</p>
        ) : error ? (
          <p className="battle-graveyard-modal__empty">{error}</p>
        ) : stacks.length === 0 ? (
          <p className="battle-graveyard-modal__empty">No fallen cards yet.</p>
        ) : (
          <div className="battle-graveyard-modal__grid">
            {stacks.map(({ card, count }) => (
              <button
                key={`${card.template_id}-${card.instance_id}`}
                type="button"
                className="battle-graveyard-modal__card"
                onClick={(event: MouseEvent<HTMLButtonElement>) => onOpenCard?.(card, event.currentTarget.getBoundingClientRect())}
              >
                <img
                  className="battle-graveyard-modal__art"
                  src={resolveCardImageSrc(card.kind as "battle" | "buff", card.template_id, card.image_key)}
                  alt={card.name}
                  loading="lazy"
                  onError={(event) => {
                    const target = event.currentTarget;
                    if (target.dataset.fallbackApplied === "1") {
                      return;
                    }
                    target.dataset.fallbackApplied = "1";
                    target.src = card.image_key
                      ? `/assets/${card.image_key.replace(/^\/+/, "").replace(/\/+/g, "/")}.png`
                      : "/assets/placeholders/card_image.svg";
                  }}
                />
                <span className="battle-graveyard-modal__stat battle-graveyard-modal__stat--mana">{card.mana_cost}</span>
                <span className="battle-graveyard-modal__stat battle-graveyard-modal__stat--attack">{card.attack}</span>
                <span className="battle-graveyard-modal__stat battle-graveyard-modal__stat--hp">{card.health_points}</span>
                <span className="battle-graveyard-modal__card-name">{card.name}</span>
                {count > 1 ? <span className="battle-graveyard-modal__count">x{count}</span> : null}
              </button>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
