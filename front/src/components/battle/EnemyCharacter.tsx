import { CharacterBlock } from "./CharacterBlock";
import type { MaskedBattlePlayerState } from "./types";

type Props = {
  player: MaskedBattlePlayerState;
  maxHp: number;
  isActive?: boolean;
};

export function EnemyCharacter({ player, maxHp, isActive = false }: Props) {
  return <CharacterBlock player={player} maxHp={maxHp} side="enemy" isActive={isActive} />;
}
