export type BattleActionType =
  | "end_turn"
  | "play_battle_card"
  | "play_buff_card"
  | "card_attack"
  | "hero_attack"
  | "hero_spell"
  | "leave_match"
  | "card_skill";

export type BattlePhase = "START" | "MAIN";
export type BattleResult = "ON_GOING" | "P1_WIN" | "P2_WIN" | "DRAW";

export type BattleCardInMatch = {
  instance_id: string;
  kind: string;
  template_id: string;
  gamer_card_id: number;
  card_level: number;
  name: string;
  description: string;
  mana_cost: number;
  attack: number;
  health_points: number;
  card_type: string;
  image_key: string;
  asset_base_key: string;
  splash_radius: number;
  base_cooldown: number;
  has_skill: boolean;
  skill_image_key: string;
};

export type BattleUnitEffect = {
  effect_type: string;
  turns_left: number;
  value: number;
  extra_value: number;
  source_type: string;
  polarity: string;
  source_instance_id: string;
  dispellable: boolean;
  targeting: string;
};

export type BattleSkillState = {
  name: string;
  code: string;
  kind: string;
  target: string;
  power: number;
  base_cooldown: number;
  cooldown_left: number;
  duration: number;
  extra_value: number;
  buff_effect: string;
  debuff_effect: string;
  cleanse_mode: string;
  ignore_tank: boolean;
  apply_count: number;
};

export type BattleUnitState = {
  instance_id: string;
  template_id: string;
  gamer_card_id: number;
  card_level: number;
  hp: number;
  max_hp: number;
  attack: number;
  splash_radius: number;
  is_tank: boolean;
  card_type: string;
  base_cooldown: number;
  cooldown: number;
  attacks_this_turn: number;
  summoned_in_turn: number;
  image_key: string;
  asset_base_key: string;
  has_skill: boolean;
  skill_image_key: string;
  skill: BattleSkillState | null;
  effects: BattleUnitEffect[];
  resurrected_used: boolean;
};

export type BattleEventTarget = {
  instance_id: string;
  template_id?: string;
  amount?: number;
  died?: boolean;
  new_hp?: number;
};

export type BattleEvent = {
  id?: string;
  event_id?: string;
  type: string;
  effect_kind?: string;
  player_index?: number;
  source_kind?: string;
  source_instance_id?: string;
  source_template_id?: string;
  source_hero_code?: string;
  source_card_template_id?: string;
  vfx_key?: string;
  sfx_key?: string;
  target_slot?: number;
  targets?: BattleEventTarget[];
  visible_for_player_index?: number;
};

export type MaskedBattlePlayerState = {
  player_id: number;
  user_id: number;
  hero_id: number;
  hero_code: string;
  hero_hp: number;
  hero_level: number;
  hero_attack_power: number;
  hero_attack_cooldown: number;
  hero_attack_base_cooldown: number;
  hero_splash_radius: number;
  hero_ability_cooldown: number;
  hero_ability_base_cooldown?: number;
  hero_ability_mana_cost?: number;
  mana: number;
  turns: number;
  table: Array<BattleUnitState | null>;
  hand?: BattleCardInMatch[];
  deck?: BattleCardInMatch[];
  discard?: BattleCardInMatch[];
  graveyard?: BattleCardInMatch[];
  hand_count?: number;
  deck_count?: number;
  discard_count?: number;
  graveyard_count?: number;
};

export type MaskedBattleMatchState = {
  match_id: number;
  version: number;
  active_player: number;
  phase: BattlePhase;
  finished: boolean;
  result: BattleResult;
  players: [MaskedBattlePlayerState | null, MaskedBattlePlayerState | null];
  events?: BattleEvent[];
  turn_started_at: number;
  turn_deadline_at: number;
  turn_time_sec: number;
  loading_ready: [boolean, boolean];
  started_at?: number;
  server_now: number;
};

export type ApplyBattleActionRequest = {
  type: BattleActionType;
  card_instance_id?: string;
  target_instance_id?: string;
  attack_hero?: boolean;
  expected_version: number;
  target_slot?: number;
};
