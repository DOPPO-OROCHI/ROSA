import { BoardSlot } from "./BoardSlot";
import type { BattleUnitState } from "./types";

type Props = {
  units: Array<BattleUnitState | null>;
  side: "player" | "enemy";
  canPlayIntoEmpty?: boolean;
  selectedUnitId?: string;
  targetUnitIds?: string[];
  animatingUnitId?: string;
  hitUnitIds?: string[];
  onFilledSlotClick?: (unit: BattleUnitState, slotIndex: number) => void;
  onEmptySlotClick?: (slotIndex: number) => void;
};

export function BoardLane({
  units,
  side,
  canPlayIntoEmpty = false,
  selectedUnitId = "",
  targetUnitIds = [],
  animatingUnitId = "",
  hitUnitIds = [],
  onFilledSlotClick,
  onEmptySlotClick,
}: Props) {
  return (
    <div className={`battle-board-lane battle-board-lane--${side}`}>
      {units.map((unit, index) => (
        <BoardSlot
          key={`${side}-${index}-${unit?.instance_id ?? "empty"}`}
          unit={unit}
          side={side}
          playable={unit == null && canPlayIntoEmpty}
          selected={Boolean(unit && unit.instance_id === selectedUnitId)}
          attackTarget={Boolean(unit && targetUnitIds.includes(unit.instance_id))}
          animating={Boolean(unit && unit.instance_id === animatingUnitId)}
          hit={Boolean(unit && hitUnitIds.includes(unit.instance_id))}
          onClick={
            unit
              ? () => onFilledSlotClick?.(unit, index)
              : onEmptySlotClick
                ? () => onEmptySlotClick(index)
                : undefined
          }
        />
      ))}
    </div>
  );
}
