import { useEffect, useMemo, useState } from "react";
import { resolveHeroAssetVariantSrc } from "../lib/api";
import type { Hero } from "../types";
import { AutoFitText } from "./AutoFitText";

type Props = {
  open: boolean;
  heroes: Hero[];
  selectedHero: Hero | null;
  onClose: () => void;
  onChooseHero: (hero: Hero) => Promise<void> | void;
};

function wrapIndex(index: number, length: number) {
  if (length <= 0) {
    return 0;
  }

  return ((index % length) + length) % length;
}

export function HeroSelect({ open, heroes, selectedHero, onClose, onChooseHero }: Props) {
  const [currentIndex, setCurrentIndex] = useState(0);
  const [slideDirection, setSlideDirection] = useState<"left" | "right">("right");
  const [slideNonce, setSlideNonce] = useState(0);
  const [missingFullArtHeroCodes, setMissingFullArtHeroCodes] = useState<Record<string, boolean>>({});
  const [submitting, setSubmitting] = useState(false);
  const [closing, setClosing] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) {
      setSubmitting(false);
      setClosing(false);
      setError("");
    }
  }, [open]);

  useEffect(() => {
    if (!open || heroes.length === 0) {
      return;
    }

    const selectedIndex = selectedHero
      ? heroes.findIndex((hero) => hero.hero_code === selectedHero.hero_code)
      : -1;

    setCurrentIndex(selectedIndex >= 0 ? selectedIndex : 0);
    setSlideDirection("right");
    setSlideNonce(0);
    setError("");
  }, [heroes, open, selectedHero]);

  const currentHero = heroes[wrapIndex(currentIndex, heroes.length)] ?? null;
  const missingFullArt = currentHero ? missingFullArtHeroCodes[currentHero.hero_code] === true : false;
  const viewerStageClass = useMemo(
    () => `hero-viewer__stage hero-viewer__stage--${slideDirection}`,
    [slideDirection],
  );

  function handleFullArtError(hero: Hero) {
    setMissingFullArtHeroCodes((current) => ({
      ...current,
      [hero.hero_code]: true,
    }));
  }

  function showPreviousHero() {
    if (heroes.length <= 1) {
      return;
    }

    setSlideDirection("left");
    setSlideNonce((current) => current + 1);
    setCurrentIndex((current) => wrapIndex(current - 1, heroes.length));
    setError("");
  }

  function showNextHero() {
    if (heroes.length <= 1) {
      return;
    }

    setSlideDirection("right");
    setSlideNonce((current) => current + 1);
    setCurrentIndex((current) => wrapIndex(current + 1, heroes.length));
    setError("");
  }

  async function handleChooseHero() {
    if (!currentHero || submitting || closing) {
      return;
    }

    setSubmitting(true);
    setError("");

    try {
      await onChooseHero(currentHero);
      setClosing(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to choose hero");
    } finally {
      setSubmitting(false);
    }
  }

  function handleClose() {
    if (closing) {
      return;
    }

    setClosing(true);
  }

  if (!open) {
    return null;
  }

  return (
    <div className={`overlay hero-select-overlay ${closing ? "overlay--closing" : ""}`} onClick={handleClose}>
      <section
        className={`hero-select surface ${closing ? "hero-select--closing" : ""}`}
        onClick={(event) => event.stopPropagation()}
        onAnimationEnd={() => {
          if (closing) {
            onClose();
          }
        }}
      >
        <header className="hero-select__header">
          <button type="button" className="picker-close" onClick={handleClose}>
            CLOSE
          </button>
        </header>

        <div className="hero-select__body">
          {heroes.length === 0 ? (
            <div className="hero-select__empty">NO HEROES AVAILABLE</div>
          ) : currentHero ? (
            <section className="hero-carousel">
              <div className="hero-carousel__viewer">
                <button
                  type="button"
                  className="hero-carousel__nav hero-carousel__nav--left"
                  onClick={showPreviousHero}
                  disabled={closing}
                  aria-label="Previous hero"
                >
                  {"<"}
                </button>

                <article className="hero-viewer hero-viewer--inline surface">
                  <div key={`${currentHero.hero_code}-${slideNonce}`} className={viewerStageClass}>
                    <div className="hero-viewer__frame">
                      {missingFullArt ? (
                        <div className="hero-viewer__missing-art">
                          <span>THIS HERO HAS NO ART YEAT</span>
                        </div>
                      ) : (
                        <img
                          className="hero-viewer__art"
                          src={resolveHeroAssetVariantSrc(currentHero.hero_code, "full_art")}
                          alt={currentHero.name}
                          onError={() => handleFullArtError(currentHero)}
                        />
                      )}

                      <div className="hero-viewer__description">{currentHero.description}</div>
                      <span className="hero-viewer__anchor hero-viewer__anchor--name">
                        <AutoFitText
                          text={currentHero.name}
                          className="hero-viewer__name"
                          maxFontSize={14}
                          minFontSize={8}
                        />
                      </span>
                      <span className="hero-viewer__anchor hero-viewer__anchor--attack">
                        <span className="hero-viewer__stat">{currentHero.attack_power}</span>
                      </span>
                      <span className="hero-viewer__anchor hero-viewer__anchor--hp">
                        <span className="hero-viewer__stat">{currentHero.health_points}</span>
                      </span>
                    </div>
                  </div>

                  <button
                    type="button"
                    className="hero-viewer__choose hero-viewer__choose--static"
                    onClick={handleChooseHero}
                    disabled={submitting || closing}
                  >
                    {submitting ? "CHOOSING..." : "CHOOSE"}
                  </button>
                </article>

                <button
                  type="button"
                  className="hero-carousel__nav hero-carousel__nav--right"
                  onClick={showNextHero}
                  disabled={closing}
                  aria-label="Next hero"
                >
                  {">"}
                </button>
              </div>

              {error ? <p className="hero-viewer__error">{error}</p> : null}
            </section>
          ) : null}
        </div>
      </section>
    </div>
  );
}
