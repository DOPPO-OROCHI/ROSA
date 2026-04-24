export type MeResponse = {
  user_id: number;
  username: string;
  first_name: string;
  rating: number;
  xp: number;
  selected_hero_code?: string;
  selected_hero_name?: string;
};

export type Hero = {
  hero_id: number;
  hero_code: string;
  name: string;
  description: string;
  image_key: string;
  level: number;
  health_points: number;
  attack_power: number;
  attack_cooldown: number;
  ability: {
    name: string;
    code: string;
    description: string;
    kind: string;
    target: string;
    cool_down: number;
    mana_cost: number;
    power: number;
    duration: number;
    extra_value: number;
    apply_count: number;
    buff_effect: string;
    debuff_effect: string;
    ignore_tank: boolean;
  };
};

export type HeroesResponse = {
  heroes: Hero[];
};

export type BattleCard = {
  kind: "battle";
  template_id: string;
  name: string;
  description: string;
  card_type: string;
  mana_cost: number;
  health_points: number;
  attack: number;
  splash_radius: number;
  base_cooldown: number;
  has_skill?: boolean;
  skill?: {
    base_cooldown: number;
  } | null;
  max_copies: number;
  copies: number;
  image_key: string;
  asset_base_key: string;
  passive?: {
    name: string;
    code: string;
    description: string;
    kind: string;
    trigger: string;
    effect_kind: string;
    target: string;
    target_race: string;
    power: number;
    duration: number;
    extra_value: number;
    apply_count: number;
    buff_effect: string;
    debuff_effect: string;
    condition: string;
    condition_race: string;
    condition_value: number;
    event_filter: string;
    event_race: string;
    scale_mode: string;
    event_is_tank: boolean;
    ignore_tank: boolean;
  } | null;
};

export type BuffCard = {
  kind: "buff";
  template_id: string;
  name: string;
  description: string;
  mana_cost: number;
  buff_type: string;
  buff_value: number;
  duration: number;
  max_copies: number;
  copies: number;
  image_key: string;
  asset_base_key: string;
};

export type CardsResponse = {
  battle: BattleCard[];
  buff: BuffCard[];
};

export type DeckEntry = {
  kind: "battle" | "buff";
  template_id: string;
  count: number;
};

export type DeckResponse = {
  entries: DeckEntry[];
};

export type QueueState = "idle" | "searching" | "pending_match" | "penalty";

export type JoinQueueResponse = {
  state: QueueState;
  opponent_user_id?: number;
};

export type LeaveQueueResponse = {
  state: QueueState;
};

export type QueueStatusResponse = {
  state: QueueState;
  opponent_user_id?: number;
  search_duration_sec?: number;
  penalty_until?: string;
  accept_deadline_at?: string;
  accepted_by_me?: boolean;
  accepted_by_opponent?: boolean;
};
