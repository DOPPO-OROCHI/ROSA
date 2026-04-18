import { useEffect, useMemo, useState } from "react";
import { resolveCardAssetVariantSrc, resolveHeroAssetVariantSrc } from "../lib/api";
import type { Hero } from "../types";

type QueueState = "idle" | "searching" | "pending_match" | "penalty";
type QueueDeckCard = {
  templateId: string;
  name: string;
  count: number;
};

type Props = {
  open: boolean;
  searching: boolean;
  queueState: QueueState;
  penaltyUntil?: string;
  busy: boolean;
  error: string;
  queueHint: string;
  searchDurationSec: number;
  canQueue: boolean;
  selectedHero: Hero | null;
  deckCards: QueueDeckCard[];
  onClose: () => void;
  onFindMatch: () => void;
  onCancelSearch: () => void;
};

export function GameModePanel({
  open,
  searching,
  queueState,
  penaltyUntil,
  busy,
  error,
  queueHint,
  searchDurationSec,
  canQueue,
  selectedHero,
  deckCards,
  onClose,
  onFindMatch,
  onCancelSearch,
}: Props) {
  const [selectedMode, setSelectedMode] = useState<"ranked">("ranked");
  const [dotCount, setDotCount] = useState(1);
  const [penaltyNow, setPenaltyNow] = useState(() => Date.now());

  useEffect(() => {
    if (!searching) {
      setDotCount(1);
      return;
    }

    const id = window.setInterval(() => {
      setDotCount((current) => (current >= 3 ? 1 : current + 1));
    }, 300);

    return () => window.clearInterval(id);
  }, [searching]);

  useEffect(() => {
    if (queueState !== "penalty" || !penaltyUntil) {
      return;
    }

    setPenaltyNow(Date.now());
    const id = window.setInterval(() => {
      setPenaltyNow(Date.now());
    }, 1000);

    return () => window.clearInterval(id);
  }, [penaltyUntil, queueState]);

  const searchTimer = useMemo(() => {
    const minutes = Math.floor(searchDurationSec / 60)
      .toString()
      .padStart(2, "0");
    const seconds = Math.floor(searchDurationSec % 60)
      .toString()
      .padStart(2, "0");
    return `${minutes}:${seconds}`;
  }, [searchDurationSec]);

  const penaltyRemainingSec = useMemo(() => {
    if (queueState !== "penalty" || !penaltyUntil) {
      return 0;
    }

    const deadline = Date.parse(penaltyUntil);
    if (Number.isNaN(deadline)) {
      return 0;
    }

    return Math.max(0, Math.ceil((deadline - penaltyNow) / 1000));
  }, [penaltyNow, penaltyUntil, queueState]);

  const penaltyTimer = useMemo(() => {
    const minutes = Math.floor(penaltyRemainingSec / 60)
      .toString()
      .padStart(2, "0");
    const seconds = Math.floor(penaltyRemainingSec % 60)
      .toString()
      .padStart(2, "0");
    return `${minutes}:${seconds}`;
  }, [penaltyRemainingSec]);

  const searchLabel = queueState === "pending_match" ? "МАТЧ НАЙДЕН" : `ПОИСК${".".repeat(dotCount)}`;
  const penaltyActive = queueState === "penalty" && penaltyRemainingSec > 0;
  const actionDisabled = penaltyActive || busy || !canQueue;
  const visibleError = penaltyActive ? "" : error;

  return (
    <>
      {open ? (
        <div className="overlay game-mode-overlay" onClick={onClose}>
          <aside className="game-mode-panel surface" onClick={(event) => event.stopPropagation()}>
            <header className="game-mode-panel__header">
              <p className="eyebrow">РЕЖИМ ИГРЫ</p>
              <button type="button" className="picker-close" onClick={onClose}>
                X
              </button>
            </header>

            <div className="game-mode-panel__art" aria-hidden="true">
              <div className="game-mode-panel__art-glow" />
            </div>

            <div className="game-mode-panel__modes">
              <button
                type="button"
                className={`game-mode-card ${selectedMode === "ranked" ? "game-mode-card--active" : ""}`}
                onClick={() => setSelectedMode("ranked")}
              >
                <span className="game-mode-card__label">РЕЙТИНГОВЫЙ МАТЧ</span>
              </button>
            </div>

            <button
              type="button"
              className={`game-mode-panel__action${penaltyActive ? " game-mode-panel__action--penalty" : ""}`}
              onClick={onFindMatch}
              disabled={actionDisabled}
            >
              {penaltyActive ? (
                <>
                  <span className="game-mode-panel__action-label">НАЙТИ МАТЧ</span>
                  <span className="game-mode-panel__action-timer">{penaltyTimer}</span>
                </>
              ) : (
                <span className="game-mode-panel__action-label">{busy ? "ПОИСК..." : "НАЙТИ МАТЧ"}</span>
              )}
            </button>

            {queueHint ? <p className="game-mode-panel__hint">{queueHint}</p> : null}
            {visibleError ? <p className="game-mode-panel__error">{visibleError}</p> : null}
          </aside>
        </div>
      ) : null}

      {searching ? (
        <div className="overlay queue-search-overlay" onClick={onCancelSearch}>
          <div className="queue-search-stack" onClick={(event) => event.stopPropagation()}>
            <section className="queue-search-panel surface">
              <div className="queue-search-panel__top">
                <p className="eyebrow">ПОИСК МАТЧА</p>
                <span className="queue-search-panel__timer">{searchTimer}</span>
              </div>

              <div className="queue-search-panel__status">{searchLabel}</div>

              <button type="button" className="queue-search-panel__cancel" onClick={onCancelSearch}>
                ОТМЕНА
              </button>
            </section>

            <section className="queue-loadout surface">
              <div className="queue-loadout__hero">
                {selectedHero ? (
                  <img
                    src={resolveHeroAssetVariantSrc(selectedHero.hero_code, "battle_icon")}
                    alt={selectedHero.name}
                  />
                ) : null}
              </div>

              <div className="queue-loadout__deck">
                {deckCards.map((card) => (
                  <article key={card.templateId} className="queue-loadout-card">
                    <img
                      className="queue-loadout-card__art"
                      src={resolveCardAssetVariantSrc("battle", card.templateId, "view")}
                      alt={card.name}
                    />
                    {card.count > 1 ? <span className="queue-loadout-card__badge">x{card.count}</span> : null}
                  </article>
                ))}
              </div>
            </section>
          </div>
        </div>
      ) : null}
    </>
  );
}
