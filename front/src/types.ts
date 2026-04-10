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
};

export type HeroesResponse = {
  heroes: Hero[];
};

export type BattleCard = {
  kind: "battle";
  template_id: string;
  name: string;
  description: string;
  mana_cost: number;
  health_points: number;
  attack: number;
  base_cooldown: number;
  max_copies: number;
  copies: number;
  image_key: string;
  asset_base_key: string;
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
