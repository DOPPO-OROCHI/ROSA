import type { BattleEvent, BattleSkillState } from "../types";

export type UnitRect = {
  left: number;
  top: number;
  width: number;
  height: number;
  centerX: number;
  centerY: number;
};

export type ProjectileTone = "damage" | "heal" | "buff" | "debuff";

export type ProjectileMode = "none" | "single" | "sequence" | "splash";

export type ProjectileSnapshot = {
  id: string;
  tone: ProjectileTone;
  fromX: number;
  fromY: number;
  toX: number;
  toY: number;
  dx: number;
  dy: number;
  width: number;
  height: number;
};

export type ProjectileImpact = {
  id: string;
  tone: ProjectileTone;
  centerX: number;
  centerY: number;
  width: number;
  height: number;
  flashLeft: number;
  flashTop: number;
  flashWidth: number;
  flashHeight: number;
};

export type ProjectileSpread = {
  id: string;
  tone: ProjectileTone;
  fromX: number;
  fromY: number;
  toX: number;
  toY: number;
  dx: number;
  dy: number;
  width: number;
  height: number;
};

export type ProjectileStep = {
  id: string;
  tone: ProjectileTone;
  fromX: number;
  fromY: number;
  toX: number;
  toY: number;
  targetInstanceId: string;
  flashLeft: number;
  flashTop: number;
  flashWidth: number;
  flashHeight: number;
  spreadImpacts?: ProjectileImpact[];
};

export type ProjectileSequence = {
  id: string;
  steps: ProjectileStep[];
};

export type ProjectileRuntimeState = {
  projectiles: ProjectileSnapshot[];
  impacts: ProjectileImpact[];
  spreads: ProjectileSpread[];
};

export type ProjectileRule = {
  mode: ProjectileMode;
  tone: ProjectileTone;
};

export type ProjectileRuleResolver = (skill: BattleSkillState | null, event: BattleEvent) => ProjectileRule;
