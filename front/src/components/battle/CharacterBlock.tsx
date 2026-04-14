import { resolveHeroAssetVariantSrc } from "../../lib/api";
import type { MaskedBattlePlayerState } from "./types";

type Props = {
  player: MaskedBattlePlayerState;
  side: "player" | "enemy";
};

export function CharacterBlock({ player, side }: Props) {
  return (
    <section className={`battle-character battle-character--${side}`}>
      <div className="battle-character__stats">
        <span className="battle-character__stat">HP {player.hero_hp}</span>
        <span className="battle-character__stat">MP {player.mana}</span>
      </div>

      <div className="battle-character__avatar">
        <img src={resolveHeroAssetVariantSrc(player.hero_code, "battle_icon")} alt={player.hero_code} />
      </div>
    </section>
  );
}
