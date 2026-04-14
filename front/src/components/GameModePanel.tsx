import { useEffect, useMemo, useState } from "react";

type QueueState = "idle" | "searching" | "pending_match" | "penalty";

type Props = {
  open: boolean;
  searching: boolean;
  queueState: QueueState;
  busy: boolean;
  error: string;
  searchDurationSec: number;
  canQueue: boolean;
  onClose: () => void;
  onFindMatch: () => void;
  onCancelSearch: () => void;
};

export function GameModePanel({
  open,
  searching,
  queueState,
  busy,
  error,
  searchDurationSec,
  canQueue,
  onClose,
  onFindMatch,
  onCancelSearch,
}: Props) {
  const [selectedMode, setSelectedMode] = useState<"ranked">("ranked");
  const [dotCount, setDotCount] = useState(1);

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

  const searchTimer = useMemo(() => {
    const minutes = Math.floor(searchDurationSec / 60)
      .toString()
      .padStart(2, "0");
    const seconds = Math.floor(searchDurationSec % 60)
      .toString()
      .padStart(2, "0");
    return `${minutes}:${seconds}`;
  }, [searchDurationSec]);

  const searchLabel = queueState === "pending_match" ? "МАТЧ НАЙДЕН" : `ПОИСК${".".repeat(dotCount)}`;

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
              className="game-mode-panel__action"
              onClick={onFindMatch}
              disabled={busy || !canQueue}
            >
              {busy ? "ПОИСК..." : "НАЙТИ МАТЧ"}
            </button>

            {error ? <p className="game-mode-panel__error">{error}</p> : null}
          </aside>
        </div>
      ) : null}

      {searching ? (
        <div className="overlay queue-search-overlay" onClick={onCancelSearch}>
          <section className="queue-search-panel surface" onClick={(event) => event.stopPropagation()}>
            <div className="queue-search-panel__top">
              <p className="eyebrow">ПОИСК МАТЧА</p>
              <span className="queue-search-panel__timer">{searchTimer}</span>
            </div>

            <div className="queue-search-panel__status">{searchLabel}</div>

            <button type="button" className="queue-search-panel__cancel" onClick={onCancelSearch}>
              ОТМЕНА
            </button>
          </section>
        </div>
      ) : null}
    </>
  );
}
