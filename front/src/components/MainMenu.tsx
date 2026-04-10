import { resolveImageSrc } from "../lib/api";
import type { Hero, MeResponse } from "../types";

type Props = {
  me: MeResponse | null;
  selectedHero: Hero | null;
  heroes: Hero[];
  heroPickerOpen: boolean;
  setHeroPickerOpen: (open: boolean) => void;
  chooseHero: (hero: Hero) => void;
  onInventory: () => void;
};

export function MainMenu(props: Props) {
  return (
    <>
      <section className="main-menu surface">
        <div className="video-stage" aria-hidden="true">
          <div className="video-stage__glow" />
          <div className="video-stage__label">shared video zone</div>
        </div>

        <header className="menu-topbar">
          <button type="button" className="top-slot top-slot--left">
            Friends
          </button>
          <h1 className="menu-title">PROJECT ROSE</h1>
          <button type="button" className="top-slot top-slot--right">
            Balance
          </button>
        </header>

        <section className="hero-focus">
          <button type="button" className="hero-avatar" onClick={() => props.setHeroPickerOpen(true)}>
            {props.selectedHero ? (
              <img src={resolveImageSrc(props.selectedHero.image_key)} alt={props.selectedHero.name} />
            ) : (
              <span>Hero</span>
            )}
          </button>
          <div className="hero-nameplate">{props.selectedHero?.name ?? "No hero selected"}</div>
          <div className="player-tag">
            {props.me ? `${props.me.username} / rating ${props.me.rating}` : "guest / no session"}
          </div>
        </section>

        <section className="menu-actions">
          <button type="button" className="menu-button menu-button--primary">
            Start Match
          </button>
          <button type="button" className="menu-button" onClick={props.onInventory}>
            Inventory
          </button>
          <button type="button" className="menu-panel">
            Shop Placeholder
          </button>
        </section>
      </section>

      {props.heroPickerOpen ? (
        <div className="overlay" onClick={() => props.setHeroPickerOpen(false)}>
          <div className="picker surface" onClick={(event) => event.stopPropagation()}>
            <div className="section-head">
              <div>
                <p className="eyebrow">Hero Select</p>
                <h2>Выбор персонажа</h2>
              </div>
              <button type="button" className="picker-close" onClick={() => props.setHeroPickerOpen(false)}>
                Close
              </button>
            </div>
            <div className="hero-grid">
              {props.heroes.map((hero) => (
                <button
                  key={hero.hero_code}
                  type="button"
                  className={`hero-card ${hero.hero_code === props.selectedHero?.hero_code ? "hero-card--active" : ""}`}
                  onClick={() => props.chooseHero(hero)}
                >
                  <span className="hero-card__avatar">
                    <img src={resolveImageSrc(hero.image_key)} alt={hero.name} />
                  </span>
                  <strong>{hero.name}</strong>
                </button>
              ))}
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}
