import type { BattleUnitState, MaskedBattleMatchState, MaskedBattlePlayerState } from "./types";
import { BoardLane } from "./BoardLane";
import { EndTurnButton } from "./EndTurnButton";
import { TurnTimer } from "./TurnTimer";

type Props = {
  match: MaskedBattleMatchState;
  enemy: MaskedBattlePlayerState;
  player: MaskedBattlePlayerState;
  canEndTurn: boolean;
  canPlaySelectedBattleCard: boolean;
  selectedAttackerId?: string;
  attackTargetIds?: string[];
  attackHint?: string;
  onEndTurn: () => void;
  onPlayerUnitSelect: (unit: BattleUnitState) => void;
  onEnemyUnitSelect: (unit: BattleUnitState) => void;
  onBoardClearSelection: () => void;
  onPlayBattleCard: (slotIndex: number) => void;
};

export function BattleField({
  match,
  enemy,
  player,
  canEndTurn,
  canPlaySelectedBattleCard,
  selectedAttackerId = "",
  attackTargetIds = [],
  attackHint = "",
  onEndTurn,
  onPlayerUnitSelect,
  onEnemyUnitSelect,
  onBoardClearSelection,
  onPlayBattleCard,
}: Props) {
  return (
    <section className="battle-field">
      <div className="battle-field__board">
        <BoardLane
          units={enemy.table}
          side="enemy"
          targetUnitIds={attackTargetIds}
          onFilledSlotClick={onEnemyUnitSelect}
          onEmptySlotClick={onBoardClearSelection}
        />
        <div className="battle-field__middle" onClick={onBoardClearSelection}>
          <TurnTimer
            startedAt={match.turn_started_at}
            deadlineAt={match.turn_deadline_at}
            totalSec={match.turn_time_sec}
          />
          <EndTurnButton disabled={!canEndTurn} onEndTurn={onEndTurn} />
        </div>
        <BoardLane
          units={player.table}
          side="player"
          selectedUnitId={selectedAttackerId}
          canPlayIntoEmpty={Boolean(canPlaySelectedBattleCard)}
          onFilledSlotClick={onPlayerUnitSelect}
          onEmptySlotClick={(slotIndex) => {
            if (canPlaySelectedBattleCard) {
              onPlayBattleCard(slotIndex);
              return;
            }
            onBoardClearSelection();
          }}
        />
      </div>
      {attackHint ? <p className="battle-card-attack-hint">{attackHint}</p> : null}
    </section>
  );
}
