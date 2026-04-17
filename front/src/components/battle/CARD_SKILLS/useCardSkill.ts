import { useEffect, useMemo, useState } from "react";
import type { BattleUnitState, MaskedBattlePlayerState } from "../types";
import type { SkillTargetTone } from "./types";
import {
  getUnitSkill,
  getBoardSkillCooldown,
  getManualSkillTargets,
  getSkillTargetTone,
  getSilenceTurns,
  getStunTurns,
  isUnitSilenced,
  isUnitStunned,
  skillNeedsManualTarget,
} from "./utils";

type Params = {
  player: MaskedBattlePlayerState | null;
  enemy: MaskedBattlePlayerState | null;
  isPlayerTurn: boolean;
  busy: boolean;
  finished: boolean;
  onInfo: (message: string) => void;
  onCast: (caster: BattleUnitState, target: BattleUnitState | null, attackHero: boolean) => Promise<void> | void;
};

export function useCardSkill({ player, enemy, isPlayerTurn, busy, finished, onInfo, onCast }: Params) {
  const [selectedCasterId, setSelectedCasterId] = useState("");

  const selectedCaster = useMemo(() => {
    if (!player || !selectedCasterId) {
      return null;
    }
    return player.table.find((unit): unit is BattleUnitState => Boolean(unit && unit.instance_id === selectedCasterId)) ?? null;
  }, [player, selectedCasterId]);

  const targetTone = useMemo<SkillTargetTone | null>(() => {
    const skill = getUnitSkill(selectedCaster);
    if (!skill) {
      return null;
    }
    return getSkillTargetTone(skill);
  }, [selectedCaster]);

  const targetState = useMemo(() => {
    const skill = getUnitSkill(selectedCaster);
    if (!selectedCaster || !skill || !skillNeedsManualTarget(skill)) {
      return { ids: [], canTargetHero: false };
    }

    return getManualSkillTargets(selectedCaster, player, enemy);
  }, [enemy, player, selectedCaster]);

  useEffect(() => {
    if (!selectedCasterId) {
      return;
    }

    if (!selectedCaster || !isPlayerTurn || busy || finished) {
      setSelectedCasterId("");
    }
  }, [busy, finished, isPlayerTurn, selectedCaster, selectedCasterId]);

  function clearSelection() {
    setSelectedCasterId("");
  }

  async function castSkill(caster: BattleUnitState, target: BattleUnitState | null, attackHero: boolean) {
    await onCast(caster, target, attackHero);
    clearSelection();
  }

  async function selectCaster(unit: BattleUnitState) {
    const skill = getUnitSkill(unit);
    if (!isPlayerTurn || busy || finished || !unit.has_skill || !skill) {
      onInfo("СКИЛЛ НЕДОСТУПЕН");
      return;
    }

    const stunTurns = getStunTurns(unit);
    if (stunTurns > 0) {
      onInfo(`STUN ${stunTurns}`);
      return;
    }

    const silenceTurns = getSilenceTurns(unit);
    if (silenceTurns > 0) {
      onInfo(`SILENCE ${silenceTurns}`);
      return;
    }

    const cooldown = getBoardSkillCooldown(unit);
    if (cooldown > 0) {
      onInfo(`КД СКИЛЛА - ${cooldown}`);
      return;
    }

    if (!skillNeedsManualTarget(skill)) {
      onInfo("КАСТУЕМ СКИЛЛ");
      await castSkill(unit, null, false);
      return;
    }

    const manualTargets = getManualSkillTargets(unit, player, enemy);
    if (manualTargets.ids.length === 0 && !manualTargets.canTargetHero) {
      onInfo("НЕТ ДОСТУПНЫХ ЦЕЛЕЙ");
      return;
    }

    setSelectedCasterId((current) => (current === unit.instance_id ? "" : unit.instance_id));
  }

  async function tryCastOnUnit(unit: BattleUnitState) {
    if (!selectedCaster || !targetState.ids.includes(unit.instance_id)) {
      clearSelection();
      return;
    }

    await castSkill(selectedCaster, unit, false);
  }

  async function tryCastOnHero() {
    if (!selectedCaster || !targetState.canTargetHero) {
      clearSelection();
      return;
    }

    await castSkill(selectedCaster, null, true);
  }

  function canUseSkillButton(unit: BattleUnitState) {
    return unit.has_skill && Boolean(getUnitSkill(unit));
  }

  function isSkillButtonDisabled(unit: BattleUnitState) {
    return isUnitStunned(unit) || isUnitSilenced(unit) || getBoardSkillCooldown(unit) > 0 || !canUseSkillButton(unit);
  }

  return {
    selectedCasterId,
    selectedCaster,
    skillTargetIds: targetState.ids,
    canTargetHero: targetState.canTargetHero,
    targetTone,
    clearSelection,
    selectCaster,
    tryCastOnUnit,
    tryCastOnHero,
    canUseSkillButton,
    isSkillButtonDisabled,
  };
}
