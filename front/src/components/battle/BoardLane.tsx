import { BoardSlot } from "./BoardSlot";
import type { BattleUnitState } from "./types";

type Props = {
  units: Array<BattleUnitState | null>;
  side: "player" | "enemy";
};

export function BoardLane({ units, side }: Props) {
  return (
    <div className={`battle-board-lane battle-board-lane--${side}`}>
      {units.map((unit, index) => (
        <BoardSlot key={`${side}-${index}-${unit?.instance_id ?? "empty"}`} unit={unit} side={side} />
      ))}
    </div>
  );
}
