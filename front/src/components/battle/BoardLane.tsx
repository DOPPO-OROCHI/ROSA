import { BoardSlot } from "./BoardSlot";
import type { SkillTargetTone } from "./CARD_SKILLS";
import type { BattleUnitState } from "./types";

type Props = {
  units: Array<BattleUnitState | null>;
  side: "player" | "enemy";
  effectSourceLabels?: Record<string, string>;
  cardNameByTemplateId?: Record<string, string>;
  canPlayIntoEmpty?: boolean;
  selectedUnitId?: string;
  selectedSkillCasterId?: string;
  targetUnitIds?: string[];
  readyUnitIds?: string[];
  skillTargetIds?: string[];
  skillTargetTone?: SkillTargetTone | null;
  animatingUnitId?: string;
  hitTokens?: Record<string, number>;
  disabledUnitIds?: string[];
  skillDisabledUnitIds?: string[];
  onFilledSlotClick?: (unit: BattleUnitState, slotIndex: number) => void;
  onEmptySlotClick?: (slotIndex: number) => void;
  onSkillClick?: (unit: BattleUnitState, slotIndex: number) => void;
};

export function BoardLane({
  units,
  side,
  effectSourceLabels = {},
  cardNameByTemplateId = {},
  canPlayIntoEmpty = false,
  selectedUnitId = "",
  selectedSkillCasterId = "",
  targetUnitIds = [],
  readyUnitIds = [],
  skillTargetIds = [],
  skillTargetTone = null,
  animatingUnitId = "",
  hitTokens = {},
  disabledUnitIds = [],
  skillDisabledUnitIds = [],
  onFilledSlotClick,
  onEmptySlotClick,
  onSkillClick,
}: Props) {
  return (
    <div className={`battle-board-lane battle-board-lane--${side}`}>
      {units.map((unit, index) => (
        <BoardSlot
          key={`${side}-${index}-${unit?.instance_id ?? "empty"}`}
          unit={unit}
          side={side}
          effectSourceLabels={effectSourceLabels}
          cardNameByTemplateId={cardNameByTemplateId}
          playable={unit == null && canPlayIntoEmpty}
          selected={Boolean(unit && unit.instance_id === selectedUnitId)}
          skillSelected={Boolean(unit && unit.instance_id === selectedSkillCasterId)}
          attackTarget={Boolean(unit && targetUnitIds.includes(unit.instance_id))}
          attackReady={Boolean(unit && readyUnitIds.includes(unit.instance_id))}
          skillTarget={Boolean(unit && skillTargetIds.includes(unit.instance_id))}
          skillTargetTone={skillTargetTone}
          animating={Boolean(unit && unit.instance_id === animatingUnitId)}
          hitToken={unit ? hitTokens[unit.instance_id] ?? 0 : 0}
          actionDisabled={Boolean(unit && disabledUnitIds.includes(unit.instance_id))}
          skillDisabled={Boolean(unit && skillDisabledUnitIds.includes(unit.instance_id))}
          onClick={
            unit
              ? () => onFilledSlotClick?.(unit, index)
              : onEmptySlotClick
                ? () => onEmptySlotClick(index)
                : undefined
          }
          onSkillClick={unit ? () => onSkillClick?.(unit, index) : undefined}
        />
      ))}
    </div>
  );
}
