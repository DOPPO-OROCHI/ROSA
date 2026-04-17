import type { BattleSkillState, BattleUnitEffect, BattleUnitState, MaskedBattlePlayerState } from "../types";
import type { SkillTargetTone } from "./types";

const EFFECT_STUN = "stun";
const EFFECT_SILENCE = "silence";

function getMaxEffectTurns(effects: BattleUnitEffect[], effectType: string): number {
  return effects.reduce((max, effect) => {
    if (effect.effect_type !== effectType) {
      return max;
    }
    return Math.max(max, effect.turns_left ?? 0);
  }, 0);
}

export function getUnitSkill(unit: BattleUnitState | null | undefined): BattleSkillState | null {
  const raw = (unit as { skill?: Record<string, unknown> } | null | undefined)?.skill;
  if (!raw) {
    return null;
  }

  return {
    name: String(raw.name ?? raw.Name ?? ""),
    code: String(raw.code ?? raw.Code ?? ""),
    kind: String(raw.kind ?? raw.Kind ?? ""),
    target: String(raw.target ?? raw.Target ?? ""),
    power: Number(raw.power ?? raw.Power ?? 0),
    base_cooldown: Number(raw.base_cooldown ?? raw.BaseCooldown ?? 0),
    cooldown_left: Number(raw.cooldown_left ?? raw.CooldownLeft ?? 0),
    duration: Number(raw.duration ?? raw.Duration ?? 0),
    extra_value: Number(raw.extra_value ?? raw.ExtraValue ?? 0),
    buff_effect: String(raw.buff_effect ?? raw.BuffEffect ?? ""),
    debuff_effect: String(raw.debuff_effect ?? raw.DebuffEffect ?? ""),
    cleanse_mode: String(raw.cleanse_mode ?? raw.CleanseMode ?? ""),
    ignore_tank: Boolean(raw.ignore_tank ?? raw.IgnoreTank ?? false),
    apply_count: Number(raw.apply_count ?? raw.ApplyCount ?? raw.hit_count ?? raw.HitCount ?? 0),
  };
}

export function getStunTurns(unit: BattleUnitState | null | undefined): number {
  return unit ? getMaxEffectTurns(unit.effects ?? [], EFFECT_STUN) : 0;
}

export function getSilenceTurns(unit: BattleUnitState | null | undefined): number {
  return unit ? getMaxEffectTurns(unit.effects ?? [], EFFECT_SILENCE) : 0;
}

export function isUnitStunned(unit: BattleUnitState | null | undefined): boolean {
  return getStunTurns(unit) > 0;
}

export function isUnitSilenced(unit: BattleUnitState | null | undefined): boolean {
  return getSilenceTurns(unit) > 0;
}

export function hasBoardSkill(unit: BattleUnitState | null | undefined): boolean {
  return Boolean(unit?.has_skill && getUnitSkill(unit));
}

export function getBoardSkillCooldown(unit: BattleUnitState | null | undefined): number {
  return getUnitSkill(unit)?.cooldown_left ?? 0;
}

export function isBoardSkillReady(unit: BattleUnitState | null | undefined): boolean {
  return hasBoardSkill(unit) && getBoardSkillCooldown(unit) <= 0 && !isUnitSilenced(unit) && !isUnitStunned(unit);
}

export function getBoardSkillLabel(unit: BattleUnitState | null | undefined): string {
  if (!unit || !unit.has_skill || !unit.skill) {
    return "";
  }

  const silenceTurns = getSilenceTurns(unit);
  if (silenceTurns > 0) {
    return `SILENCE ${silenceTurns}`;
  }

  const cooldown = getBoardSkillCooldown(unit);
  if (cooldown > 0) {
    return `CD ${cooldown}`;
  }

  return "SKILL";
}

export function getSkillTargetTone(skill: BattleSkillState | null | undefined): SkillTargetTone {
  switch (skill?.kind) {
    case "heal":
      return "heal";
    case "buff":
    case "summon":
      return "buff";
    case "debuff":
      return "debuff";
    default:
      return "damage";
  }
}

export function skillNeedsManualTarget(skill: BattleSkillState | null | undefined): boolean {
  if (!skill) {
    return false;
  }

  return skill.target === "ally_single" || skill.target === "enemy_single" || skill.target === "enemy_splash";
}

export function canSkillTargetEnemyHero(skill: BattleSkillState | null | undefined): boolean {
  if (!skill) {
    return false;
  }
  if (skill.target !== "enemy_single") {
    return false;
  }

  return skill.kind === "damage" || skill.kind === "kill" || skill.kind === "explode" || skill.kind === "hybrid";
}

export function getEnemyTankTargets(enemy: MaskedBattlePlayerState | null): BattleUnitState[] {
  if (!enemy) {
    return [];
  }

  return enemy.table.filter((unit): unit is BattleUnitState => Boolean(unit?.is_tank));
}

export function getManualSkillTargets(
  caster: BattleUnitState,
  player: MaskedBattlePlayerState | null,
  enemy: MaskedBattlePlayerState | null,
): { ids: string[]; canTargetHero: boolean } {
  const skill = getUnitSkill(caster);
  if (!skill) {
    return { ids: [], canTargetHero: false };
  }

  if (skill.target === "ally_single") {
    return {
      ids: (player?.table ?? [])
        .filter((unit): unit is BattleUnitState => Boolean(unit && unit.instance_id !== caster.instance_id))
        .map((unit) => unit.instance_id),
      canTargetHero: false,
    };
  }

  if (skill.target === "enemy_single" || skill.target === "enemy_splash") {
    const tanks = !skill.ignore_tank ? getEnemyTankTargets(enemy) : [];
    const targets = tanks.length > 0
      ? tanks
      : (enemy?.table ?? []).filter((unit): unit is BattleUnitState => Boolean(unit));

    return {
      ids: targets.map((unit) => unit.instance_id),
      canTargetHero: tanks.length === 0 && canSkillTargetEnemyHero(skill),
    };
  }

  return { ids: [], canTargetHero: false };
}

export function getUnitAuraState(unit: BattleUnitState | null | undefined): "none" | "buff" | "debuff" | "both" {
  if (!unit) {
    return "none";
  }

  const hasBuff = (unit.effects ?? []).some((effect) => effect.polarity === "buff");
  const hasDebuff = (unit.effects ?? []).some((effect) => effect.polarity === "debuff");

  if (hasBuff && hasDebuff) {
    return "both";
  }
  if (hasBuff) {
    return "buff";
  }
  if (hasDebuff) {
    return "debuff";
  }

  return "none";
}
