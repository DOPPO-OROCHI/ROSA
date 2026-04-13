import { useEffect, useState, type SyntheticEvent } from "react";
import { resolveHeroAssetVariantSrc, resolveImageSrc } from "../lib/api";
import type { Hero } from "../types";
import { AutoFitText } from "./AutoFitText";

type Props = {
  open: boolean;
  heroes: Hero[];
  onClose: () => void;
  onChooseHero: (hero: Hero) => Promise<void> | void;
};

export function HeroSelect({ open, heroes, onClose, onChooseHero }: Props) {
  const [previewHero, setPreviewHero] = useState<Hero | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) {
      setPreviewHero(null);
      setSubmitting(false);
      setError("");
    }
  }, [open]);

  if (!open) {
    return null;
  }

  function handleHeroImageError(event: SyntheticEvent<HTMLImageElement>, hero: Hero) {
    const target = event.currentTarget;
    if (target.dataset.fallbackApplied === "1") {
      return;
    }
    target.dataset.fallbackApplied = "1";
    target.src = resolveImageSrc(hero.image_key);
  }

  async function handleChooseHero() {
    if (!previewHero || submitting) {
      return;
    }

    setSubmitting(true);
    setError("");

    try {
      await onChooseHero(previewHero);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось выбрать персонажа");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="overlay hero-select-overlay" onClick={onClose}>
      <section className="hero-select surface" onClick={(event) => event.stopPropagation()}>
        <header className="hero-select__header">
          <p className="eyebrow">ВЫБОР ПЕРСОНАЖА</p>
          <button type="button" className="picker-close" onClick={onClose}>
            ЗАКРЫТЬ
          </button>
        </header>

        <div className="hero-select__body">
          {heroes.length === 0 ? (
            <div className="hero-select__empty">ЗАЙДИТЕ В СВОЙ АККАУНТ ЧТОБЫ УВИДЕТЬ ВАШИХ ПЕРСОНАЖЕЙ</div>
          ) : (
            <div className="hero-select__grid">
              {heroes.map((hero) => (
                <button
                  key={hero.hero_code}
                  type="button"
                  className="hero-select-card"
                  onClick={() => {
                    setPreviewHero(hero);
                    setError("");
                  }}
                >
                  <span className="hero-select-card__frame">
                    <img
                      className="hero-select-card__art"
                      src={resolveHeroAssetVariantSrc(hero.hero_code, "view")}
                      alt={hero.name}
                      onError={(event) => handleHeroImageError(event, hero)}
                    />
                    <span className="hero-select-card__anchor hero-select-card__anchor--name">
                      <AutoFitText
                        text={hero.name}
                        className="hero-select-card__name"
                        maxFontSize={11.2}
                        minFontSize={6}
                      />
                    </span>
                    <span className="hero-select-card__anchor hero-select-card__anchor--attack">
                      <span className="hero-select-card__stat">{hero.attack_power}</span>
                    </span>
                    <span className="hero-select-card__anchor hero-select-card__anchor--hp">
                      <span className="hero-select-card__stat">{hero.health_points}</span>
                    </span>
                  </span>
                </button>
              ))}
            </div>
          )}
        </div>

        {previewHero ? (
          <div className="hero-viewer-backdrop" onClick={() => setPreviewHero(null)}>
            <section className="hero-viewer surface" onClick={(event) => event.stopPropagation()}>
              <button type="button" className="hero-viewer__close" onClick={() => setPreviewHero(null)}>
                ЗАКРЫТЬ
              </button>

              <article className="hero-viewer__frame">
                <img
                  className="hero-viewer__art"
                  src={resolveHeroAssetVariantSrc(previewHero.hero_code, "full_art")}
                  alt={previewHero.name}
                  onError={(event) => handleHeroImageError(event, previewHero)}
                />

                <div className="hero-viewer__future hero-viewer__future--left" aria-hidden="true" />
                <div className="hero-viewer__future hero-viewer__future--right" aria-hidden="true" />

                <span className="hero-viewer__anchor hero-viewer__anchor--name">
                  <AutoFitText
                    text={previewHero.name}
                    className="hero-viewer__name"
                    maxFontSize={14}
                    minFontSize={8}
                  />
                </span>
                <span className="hero-viewer__anchor hero-viewer__anchor--attack">
                  <span className="hero-viewer__stat">{previewHero.attack_power}</span>
                </span>
                <span className="hero-viewer__anchor hero-viewer__anchor--hp">
                  <span className="hero-viewer__stat">{previewHero.health_points}</span>
                </span>

                <button
                  type="button"
                  className="hero-viewer__choose"
                  onClick={handleChooseHero}
                  disabled={submitting}
                >
                  {submitting ? "ВЫБИРАЕМ..." : "ВЫБРАТЬ"}
                </button>
              </article>

              {error ? <p className="hero-viewer__error">{error}</p> : null}
            </section>
          </div>
        ) : null}
      </section>
    </div>
  );
}
