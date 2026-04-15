import { CharacterBlock } from "./CharacterBlock";
import type { MaskedBattlePlayerState } from "./types";

type Props = {
  player: MaskedBattlePlayerState;
  maxHp: number;
  isActive?: boolean;
};

export function GamerCharacter({ player, maxHp, isActive = false }: Props) {
  return <CharacterBlock player={player} maxHp={maxHp} side="player" isActive={isActive} />;
}
