import type { SyntheticEvent } from "react";
import { resolveHeroAssetVariantSrc, resolveImageSrc } from "../lib/api";
import type { Hero, MeResponse } from "../types";
import { AutoFitText } from "./AutoFitText";
import { HeroSelect } from "./HeroSelect";

type Props = {
  me: MeResponse | null;
  selectedHero: Hero | null;
  heroes: Hero[];
  heroPickerOpen: boolean;
  setHeroPickerOpen: (open: boolean) => void;
  chooseHero: (hero: Hero) => Promise<void> | void;
  onStartMatch: () => void;
  inventoryHidden?: boolean;
  startMatchDisabled?: boolean;
  onInventory: () => void;
};

export function MainMenu(props: Props) {
  function handleHeroImageError(event: SyntheticEvent<HTMLImageElement>, hero: Hero) {
    const target = event.currentTarget;
    if (target.dataset.fallbackApplied === "1") {
      return;
    }
    target.dataset.fallbackApplied = "1";
    target.src = resolveImageSrc(hero.image_key);
  }

  return (
    <>
      <section className="main-menu surface">
        <div className="video-stage" aria-hidden="true">
          <div className="video-stage__glow" />
          <div className="video-stage__label">shared video zone</div>
        </div>

        <header className="menu-topbar">
          <button type="button" className="top-slot top-slot--left">
            <AutoFitText text="FRIENDS" className="top-slot__label" maxFontSize={14} minFontSize={8} />
          </button>
          <h1 className="menu-title">PROJECT ROSE</h1>
          <button type="button" className="top-slot top-slot--right">
            <AutoFitText text="BALANCE" className="top-slot__label" maxFontSize={14} minFontSize={8} />
          </button>
        </header>

        <section className="hero-focus">
          <button type="button" className="hero-avatar" onClick={() => props.setHeroPickerOpen(true)}>
            {props.selectedHero ? (
              <img
                src={resolveHeroAssetVariantSrc(props.selectedHero.hero_code, "battle_icon")}
                alt={props.selectedHero.name}
                onError={(event) => handleHeroImageError(event, props.selectedHero!)}
              />
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
          <button
            type="button"
            className="menu-button menu-button--primary"
            onClick={props.onStartMatch}
            disabled={props.startMatchDisabled}
          >
            Start Match
          </button>
          {props.inventoryHidden ? null : (
            <button type="button" className="menu-button" onClick={props.onInventory}>
              Inventory
            </button>
          )}
          <button type="button" className="menu-panel">
            Shop Placeholder
          </button>
        </section>
      </section>

      <HeroSelect
        open={props.heroPickerOpen}
        heroes={props.heroes}
        onClose={() => props.setHeroPickerOpen(false)}
        onChooseHero={props.chooseHero}
      />
    </>
  );
}
