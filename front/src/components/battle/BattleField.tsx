import type { BattleUnitState, MaskedBattleMatchState, MaskedBattlePlayerState } from "./types";
import { BoardLane } from "./BoardLane";
import type { SkillTargetTone } from "./CARD_SKILLS";
import { EndTurnButton } from "./EndTurnButton";
import { TurnTimer } from "./TurnTimer";

function toDisplayName(value: string): string {
  return value
    .split(/[_:\-]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

type Props = {
  match: MaskedBattleMatchState;
  enemy: MaskedBattlePlayerState;
  player: MaskedBattlePlayerState;
  canEndTurn: boolean;
  canPlaySelectedBattleCard: boolean;
  selectedAttackerId?: string;
  selectedSkillCasterId?: string;
  attackTargetIds?: string[];
  readyUnitIds?: string[];
  skillTargetIds?: string[];
  skillTargetTone?: SkillTargetTone | null;
  attackHint?: string;
  animatingUnitId?: string;
  hitTokens?: Record<string, number>;
  disabledPlayerUnitIds?: string[];
  skillDisabledPlayerUnitIds?: string[];
  onEndTurn: () => void;
  onPlayerUnitSelect: (unit: BattleUnitState) => void;
  onEnemyUnitSelect: (unit: BattleUnitState) => void;
  onBoardClearSelection: () => void;
  onPlayBattleCard: (slotIndex: number) => void;
  onPlayerUnitSkill: (unit: BattleUnitState) => void;
};

export function BattleField({
  match,
  enemy,
  player,
  canEndTurn,
  canPlaySelectedBattleCard,
  selectedAttackerId = "",
  selectedSkillCasterId = "",
  attackTargetIds = [],
  readyUnitIds = [],
  skillTargetIds = [],
  skillTargetTone = null,
  attackHint = "",
  animatingUnitId = "",
  hitTokens = {},
  disabledPlayerUnitIds = [],
  skillDisabledPlayerUnitIds = [],
  onEndTurn,
  onPlayerUnitSelect,
  onEnemyUnitSelect,
  onBoardClearSelection,
  onPlayBattleCard,
  onPlayerUnitSkill,
}: Props) {
  const effectSourceLabels = Object.fromEntries(
    match.players.flatMap((battlePlayer, playerIndex) => {
      const heroEntry = battlePlayer
        ? [[`hero:p${playerIndex}`, toDisplayName(battlePlayer.hero_code)] as const]
        : [];
      const unitEntries = (battlePlayer?.table ?? [])
        .filter((unit): unit is BattleUnitState => Boolean(unit))
        .map((unit) => [unit.instance_id, toDisplayName(unit.template_id)] as const);

      return [...heroEntry, ...unitEntries];
    }),
  );

  return (
    <section className="battle-field">
      <div className="battle-field__board">
        <BoardLane
          units={enemy.table}
          side="enemy"
          effectSourceLabels={effectSourceLabels}
          targetUnitIds={attackTargetIds}
          readyUnitIds={[]}
          skillTargetIds={skillTargetIds}
          skillTargetTone={skillTargetTone}
          animatingUnitId={animatingUnitId}
          hitTokens={hitTokens}
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
          effectSourceLabels={effectSourceLabels}
          selectedUnitId={selectedAttackerId}
          selectedSkillCasterId={selectedSkillCasterId}
          readyUnitIds={readyUnitIds}
          canPlayIntoEmpty={Boolean(canPlaySelectedBattleCard)}
          animatingUnitId={animatingUnitId}
          hitTokens={hitTokens}
          disabledUnitIds={disabledPlayerUnitIds}
          skillDisabledUnitIds={skillDisabledPlayerUnitIds}
          skillTargetIds={skillTargetIds}
          skillTargetTone={skillTargetTone}
          onFilledSlotClick={onPlayerUnitSelect}
          onSkillClick={onPlayerUnitSkill}
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
