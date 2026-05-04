import { useMemo, useState } from "react";
import type { BattleUnitEffect, BattleUnitState, MaskedBattlePlayerState } from "./types";

const EFFECT_STUN = "stun";
const EFFECT_DISARM = "disarm";
const EFFECT_OVERDRIVE = "overdrive";

type Params = {
  player: MaskedBattlePlayerState | null;
  enemy: MaskedBattlePlayerState | null;
  isPlayerTurn: boolean;
  busy: boolean;
  finished: boolean;
  onAttack: (attacker: BattleUnitState, target: BattleUnitState | null, attackHero: boolean) => Promise<void> | void;
};

function hasEffect(effects: BattleUnitEffect[] | undefined, effectType: string): boolean {
  return (effects ?? []).some((effect) => effect.effect_type === effectType);
}

export function getBoardAttackDisplayValue(unit: BattleUnitState): number {
  return unit.cooldown > 1 ? unit.cooldown : unit.attack;
}

export function getBoardAttackDisplayKind(unit: BattleUnitState): "attack" | "cooldown" {
  return unit.cooldown > 1 ? "cooldown" : "attack";
}

export function canUnitAttackNow(unit: BattleUnitState | null | undefined, ownerTurns: number): boolean {
  if (!unit) {
    return false;
  }
  if (unit.cooldown > 0) {
    return false;
  }
  if (unit.summoned_in_turn === ownerTurns) {
    return false;
  }
  if (hasEffect(unit.effects, EFFECT_STUN) || hasEffect(unit.effects, EFFECT_DISARM)) {
    return false;
  }
  const maxAttacks = hasEffect(unit.effects, EFFECT_OVERDRIVE) ? 2 : 1;
  if ((unit.attacks_this_turn ?? 0) >= maxAttacks) {
    return false;
  }
  return true;
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
    if (!canUnitAttackNow(selectedAttacker, player?.turns ?? 0)) {
      return [];
    }

    const tanks = getTankTargets(enemy);
    return (tanks.length > 0 ? tanks : getDefaultTargets(enemy)).map((unit) => unit.instance_id);
  }, [busy, enemy, finished, isPlayerTurn, player?.turns, selectedAttacker]);

  const canAttackHero = useMemo(() => {
    if (!selectedAttacker || !isPlayerTurn || busy || finished) {
      return false;
    }
    if (!canUnitAttackNow(selectedAttacker, player?.turns ?? 0)) {
      return false;
    }

    return getTankTargets(enemy).length === 0;
  }, [busy, enemy, finished, isPlayerTurn, player?.turns, selectedAttacker]);

  const infoMessage = useMemo(() => {
    if (!selectedAttacker) {
      return "";
    }
    if (selectedAttacker.summoned_in_turn === (player?.turns ?? 0)) {
      return "АТАКА НА СЛЕДУЮЩИЙ ХОД";
    }
    if (selectedAttacker.cooldown > 0) {
      return `КД АТАКИ - ${selectedAttacker.cooldown}`;
    }
    if (hasEffect(selectedAttacker.effects, EFFECT_DISARM)) {
      return "DISARM";
    }
    return "";
  }, [player?.turns, selectedAttacker]);

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

    await onAttack(selectedAttacker, target, false);
    clearSelection();
  }

  async function tryAttackHero() {
    if (!selectedAttacker) {
      clearSelection();
      return;
    }
    if (!canAttackHero) {
      clearSelection();
      return;
    }

    await onAttack(selectedAttacker, null, true);
    clearSelection();
  }

  return {
    selectedAttackerId,
    selectedAttacker,
    attackTargets,
    canAttackHero,
    infoMessage,
    clearSelection,
    selectAttacker,
    tryAttack,
    tryAttackHero,
  };
}
