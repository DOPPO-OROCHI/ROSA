import { CharacterBlock } from "./CharacterBlock";
import type { MaskedBattlePlayerState } from "./types";

type Props = {
  player: MaskedBattlePlayerState;
};

export function EnemyCharacter({ player }: Props) {
  return <CharacterBlock player={player} side="enemy" />;
}
