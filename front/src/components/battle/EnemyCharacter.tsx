import { CharacterBlock } from "./CharacterBlock";
import type { MaskedBattlePlayerState } from "./types";

type Props = {
  player: MaskedBattlePlayerState;
  maxHp: number;
  isActive?: boolean;
  attackTarget?: boolean;
  heroInstanceId?: string;
  hitToken?: number;
  attackAnimation?: {
    dx: number;
    dy: number;
  } | null;
  onClick?: () => void;
};

export function EnemyCharacter({
  player,
  maxHp,
  isActive = false,
  attackTarget = false,
  heroInstanceId = "",
  hitToken = 0,
  attackAnimation = null,
  onClick,
}: Props) {
  return (
    <CharacterBlock
      player={player}
      maxHp={maxHp}
      side="enemy"
      isActive={isActive}
      attackTarget={attackTarget}
      heroInstanceId={heroInstanceId}
      hitToken={hitToken}
      attackAnimation={attackAnimation}
      onClick={onClick}
    />
  );
}
