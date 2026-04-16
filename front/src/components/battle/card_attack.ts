import { useMemo, useState } from "react";
import type { BattleUnitState, MaskedBattlePlayerState } from "./types";

type Params = {
  player: MaskedBattlePlayerState | null;
  enemy: MaskedBattlePlayerState | null;
  isPlayerTurn: boolean;
  busy: boolean;
  finished: boolean;
  onAttack: (attacker: BattleUnitState, target: BattleUnitState) => Promise<void> | void;
};

export function getBoardAttackDisplayValue(unit: BattleUnitState): number {
  return unit.cooldown > 1 ? unit.cooldown : unit.attack;
}

export function getBoardAttackDisplayKind(unit: BattleUnitState): "attack" | "cooldown" {
  return unit.cooldown > 1 ? "cooldown" : "attack";
}

function getTankTargets(enemy: MaskedBattlePlayerState | null): BattleUnitState[] {
  if (!enemy) {
    return [];
  }

  return enemy.table.filter((unit): unit is BattleUnitState => Boolean(unit?.is_tank));
}

function getDefaultTargets(enemy: MaskedBattlePlayerState | null): BattleUnitState[] {
  if (!enemy) {
    return [];
  }

  return enemy.table.filter((unit): unit is BattleUnitState => Boolean(unit));
}

export function useCardAttack({ player, enemy, isPlayerTurn, busy, finished, onAttack }: Params) {
  const [selectedAttackerId, setSelectedAttackerId] = useState("");

  const selectedAttacker = useMemo(() => {
    if (!player || !selectedAttackerId) {
      return null;
    }
    return player.table.find((unit): unit is BattleUnitState => Boolean(unit && unit.instance_id === selectedAttackerId)) ?? null;
  }, [player, selectedAttackerId]);

  const attackTargets = useMemo(() => {
    if (!selectedAttacker || !isPlayerTurn || busy || finished) {
      return [];
    }
    if (selectedAttacker.cooldown > 0) {
      return [];
    }

    const tanks = getTankTargets(enemy);
    return (tanks.length > 0 ? tanks : getDefaultTargets(enemy)).map((unit) => unit.instance_id);
  }, [busy, enemy, finished, isPlayerTurn, selectedAttacker]);

  const infoMessage = useMemo(() => {
    if (!selectedAttacker) {
      return "";
    }
    if (selectedAttacker.cooldown > 0) {
      return `КД АТАКИ - ${selectedAttacker.cooldown}`;
    }
    return "";
  }, [selectedAttacker]);

  function clearSelection() {
    setSelectedAttackerId("");
  }

  function selectAttacker(unit: BattleUnitState) {
    setSelectedAttackerId((current) => (current === unit.instance_id ? "" : unit.instance_id));
  }

  async function tryAttack(target: BattleUnitState) {
    if (!selectedAttacker) {
      clearSelection();
      return;
    }
    if (!attackTargets.includes(target.instance_id)) {
      clearSelection();
      return;
    }

    await onAttack(selectedAttacker, target);
    clearSelection();
  }

  return {
    selectedAttackerId,
    selectedAttacker,
    attackTargets,
    infoMessage,
    clearSelection,
    selectAttacker,
    tryAttack,
  };
}
