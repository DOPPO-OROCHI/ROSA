import { BoardLane } from "./BoardLane";
import { EndTurnButton } from "./EndTurnButton";
import { TurnTimer } from "./TurnTimer";
import type { MaskedBattleMatchState, MaskedBattlePlayerState } from "./types";

type Props = {
  match: MaskedBattleMatchState;
  enemy: MaskedBattlePlayerState;
  player: MaskedBattlePlayerState;
  canEndTurn: boolean;
  onEndTurn: () => void;
};

export function BattleField({ match, enemy, player, canEndTurn, onEndTurn }: Props) {
  return (
    <section className="battle-field">
      <div className="battle-field__board">
        <BoardLane units={enemy.table} side="enemy" />
        <div className="battle-field__middle">
          <TurnTimer
            startedAt={match.turn_started_at}
            deadlineAt={match.turn_deadline_at}
            totalSec={match.turn_time_sec}
          />
          <EndTurnButton disabled={!canEndTurn} onEndTurn={onEndTurn} />
        </div>
        <BoardLane units={player.table} side="player" />
      </div>
    </section>
  );
}
