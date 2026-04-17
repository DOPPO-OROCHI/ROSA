import { CharacterBlock } from "./CharacterBlock";
import type { MaskedBattlePlayerState } from "./types";

type Props = {
  player: MaskedBattlePlayerState;
  maxHp: number;
  isActive?: boolean;
  heroInstanceId?: string;
  hitToken?: number;
  attackAnimation?: {
    dx: number;
    dy: number;
  } | null;
};

export function GamerCharacter({ player, maxHp, isActive = false, heroInstanceId = "", hitToken = 0, attackAnimation = null }: Props) {
  return <CharacterBlock player={player} maxHp={maxHp} side="player" isActive={isActive} heroInstanceId={heroInstanceId} hitToken={hitToken} attackAnimation={attackAnimation} />;
}
