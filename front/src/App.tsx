import { type CSSProperties, type MouseEvent, useEffect, useMemo, useRef, useState } from "react";
import {
  getAssetTone,
  resolveAssetLabel,
  resolveBattleCardImageKey,
  resolveBoardBackgroundSrc,
  resolveBuffCardImageKey,
  resolveCardFallbackSrc,
  resolveHeroFallbackSrc,
  resolveHeroImageKey,
  resolveImageSrc,
} from "./assets";
import { apiUrl } from "./config";
import {
  bootstrapTelegramWebApp,
  exitMiniAppFullscreen,
  getTelegramInitData,
  isMiniAppFullscreen,
  requestMiniAppFullscreen,
} from "./telegram";
import { GameCard, type GameCardData } from "./components/GameCard";

type TabId = "home" | "inventory";

type ApiError = { error?: string };

type MeResponse = {
  user_id: number;
  tg_id: number;
  username: string;
  first_name: string;
  rating: number;
  xp: number;
  selected_hero_code?: string;
  selected_hero_name?: string;
};

type OwnedHero = {
  hero_id: number;
  hero_code: string;
  name: string;
  level: number;
  health_points: number;
  attack_power: number;
  attack_cooldown: number;
  splash_radius: number;
  description: string;
  image_key: string;
  asset_base_key: string;
};

type BattleCard = {
  kind: "battle";
  template_id: string;
  name: string;
  description: string;
  card_type: string;
  mana_cost: number;
  health_points: number;
  attack: number;
  splash_radius: number;
  cooldown: number;
  is_tank: boolean;
  buff_slot: boolean;
  max_copies: number;
  owned_card_id: number;
  copies: number;
  level: number;
  xp: number;
  image_key: string;
  asset_base_key: string;
  skill_name?: string;
  skill_code?: string;
  skill_trigger?: string;
  skill_target?: string;
  skill_cooldown?: number;
};

type BuffCard = {
  kind: "buff";
  template_id: string;
  name: string;
  description: string;
  mana_cost: number;
  buff_type: string;
  buff_value: number;
  only_for: string;
  duration: number;
  max_copies: number;
  owned_card_id: number;
  copies: number;
  level: number;
  xp: number;
  image_key: string;
  asset_base_key: string;
};

type CardsResponse = {
  battle: BattleCard[];
  buff: BuffCard[];
};

type CardCatalogEntry = {
  kind: "battle" | "buff";
  template_id: string;
  name: string;
  description: string;
  mana_cost: number;
  card_type?: string;
  image_key: string;
  attack?: number;
  health_points?: number;
  cooldown?: number;
  max_copies?: number;
  duration?: number;
  buff_value?: number;
  buff_type?: string;
  skill_name?: string;
  skill_code?: string;
  skill_trigger?: string;
  skill_target?: string;
  skill_cooldown?: number;
};

type DeckEntry = {
  kind: "battle" | "buff";
  template_id: string;
  count: number;
};

type DeckResponse = {
  entries: DeckEntry[];
};

type CardsInMatch = {
  InstanceID?: string;
  Kind?: string;
  TemplateID?: string;
  CardLevel?: number;
  instance_id?: string;
  kind?: string;
  template_id?: string;
  card_level?: number;
};

type UnitState = {
  InstanceID?: string;
  TemplateID?: string;
  HP?: number;
  Attack?: number;
  MaxHP?: number;
  BaseCooldown?: number;
  Cooldown?: number;
  IsTank?: boolean;
  SummonedInTurn?: number;
  SkillName?: string;
  SkillCode?: string;
  SkillTrigger?: string;
  SkillTarget?: string;
  SkillCooldown?: number;
  SkillCooldownLeft?: number;
  Effects?: Array<{ EffectType?: string; TurnsLeft?: number; Value?: number }>;
  instance_id?: string;
  template_id?: string;
  hp?: number;
  attack?: number;
  max_hp?: number;
  base_cooldown?: number;
  cooldown?: number;
  is_tank?: boolean;
  summoned_in_turn?: number;
  skill_name?: string;
  skill_code?: string;
  skill_trigger?: string;
  skill_target?: string;
  skill_cooldown?: number;
  skill_cooldown_left?: number;
  effects?: Array<{ effect_type?: string; turns_left?: number; value?: number }>;
};

type GraveEntryState = {
  Unit?: UnitState;
  unit?: UnitState;
  DiedAtTurn?: number;
  died_at_turn?: number;
};

type MatchPlayer = {
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
  hero_ability_mana_cost?: number;
  HeroAbilityManaCost?: number;
  mana: number;
  turns: number;
  table: Array<UnitState | null>;
  hand?: CardsInMatch[];
  deck?: CardsInMatch[];
  discard?: CardsInMatch[];
  graveyard?: GraveEntryState[];
  GraveYard?: GraveEntryState[];
  hand_count?: number;
  deck_count?: number;
  disc_count?: number;
};

type MatchEvent = {
  type: string;
  player_index?: number;
  source_kind?: string;
  source_instance_id?: string;
  source_hero_code?: string;
  target_slot?: number;
  targets?: Array<{
    instance_id?: string;
    template_id?: string;
    amount?: number;
    died?: boolean;
    new_hp?: number;
  }>;
  vfx_key?: string;
  sfx_key?: string;
  source_template_id?: string;
  source_card_template_id?: string;
};

type BoardAttackAnimation = {
  sourceKind: "unit" | "hero";
  sourceInstanceId?: string;
  sourceSide?: "own" | "enemy";
  dx: number;
  dy: number;
  targetIds: string[];
};

type MatchState = {
  match_id: number;
  version: number;
  active_player: number;
  phase: string;
  finished: boolean;
  result: string;
  turn_started_at?: number;
  turn_time_sec?: number;
  turn_deadline_at?: number;
  server_now?: number;
  players: [MatchPlayer | null, MatchPlayer | null];
  events?: MatchEvent[];
};

type MatchmakingState = "idle" | "searching" | "pending_match" | "penalty";
type MatchmakingMode = "ranked";

type QueueStatusResponse = {
  state: MatchmakingState;
  opponent_user_id?: number;
  search_duration_sec?: number;
  penalty_until?: string;
  accept_deadline_at?: string;
  accepted_by_me?: boolean;
  accepted_by_opponent?: boolean;
};

type AcceptQueueResponse = {
  state: MatchmakingState;
};

type DeclineQueueResponse = {
  state: MatchmakingState;
};

type DragAttackState = {
  sourceId: string;
  sourceX: number;
  sourceY: number;
  currentX: number;
  currentY: number;
};

type ToastEntry = {
  id: number;
  message: string;
  tone: "info" | "error";
};

type CatalogKind = "battle" | "buff";
type CatalogSort = "mana" | "attack" | "hp" | "tank";

type CardPreview = {
  kind: "battle" | "buff";
  name: string;
  description: string;
  imageKey: string;
  race?: string;
  mana?: number;
  hp?: number;
  attack?: number;
  cooldown?: number;
  skillCooldown?: number;
  buffType?: string;
  buffValue?: number;
  duration?: number;
};

const SKILL_TRIGGER_ACTIVE = "active";
const TARGET_NONE = "none";
const TARGET_SELF = "self";
const TARGET_ALLY_UNIT = "ally_unit";
const TARGET_ENEMY_UNIT = "enemy_unit";
const TARGET_ALLY_ALL = "ally_all";
const TARGET_ENEMY_ALL = "enemy_all";
const TARGET_BOTH_ALL = "both_all";
const TARGET_ENEMY_SPLASH = "enemy_splash";
const TARGET_ALLY_SPLASH = "ally_splash";
const TARGET_ALLY_GRAVE_SINGLE = "ally_grave_single";
const SKILL_DAMAGE_SINGLE = "damage_single";

type HeroSpellTargetMode = "own-unit" | "enemy-unit" | "enemy-any" | "enemy-hero-only";

type SkillMeta = {
  name: string;
  code: string;
  trigger: string;
  target: string;
  cooldown: number;
};

const skillFallbackByTemplate: Record<string, SkillMeta> = {
  imperial_guardian: {
    name: "Осколочные гранаты",
    code: "damage_splash",
    trigger: SKILL_TRIGGER_ACTIVE,
    target: TARGET_ENEMY_SPLASH,
    cooldown: 2,
  },
  apply_buff: {
    name: "Крупнокалиберные боеприпасы",
    code: "apply_buff",
    trigger: SKILL_TRIGGER_ACTIVE,
    target: TARGET_SELF,
    cooldown: 3,
  },
  machine_gun_crew: {
    name: "Крупнокалиберные боеприпасы",
    code: "apply_buff",
    trigger: SKILL_TRIGGER_ACTIVE,
    target: TARGET_SELF,
    cooldown: 3,
  },
  snipers: {
    name: "Устранение",
    code: "damage_single",
    trigger: SKILL_TRIGGER_ACTIVE,
    target: TARGET_ENEMY_UNIT,
    cooldown: 5,
  },
  cursed_pack: {
    name: "Рваные раны",
    code: "apply_debuff",
    trigger: SKILL_TRIGGER_ACTIVE,
    target: TARGET_ENEMY_UNIT,
    cooldown: 3,
  },
  cyber_scout: {
    name: "Сканирование местности",
    code: "reveal_enemy_hand",
    trigger: SKILL_TRIGGER_ACTIVE,
    target: TARGET_NONE,
    cooldown: 3,
  },
  succubus: {
    name: "Ужас",
    code: "inc_enemy_cd_single",
    trigger: SKILL_TRIGGER_ACTIVE,
    target: TARGET_ENEMY_UNIT,
    cooldown: 2,
  },
};

const defaultDeck: DeckEntry[] = [
  { kind: "battle", template_id: "imperial_guardian", count: 5 },
  { kind: "battle", template_id: "mechanical_knight", count: 3 },
  { kind: "battle", template_id: "drones", count: 4 },
  { kind: "buff", template_id: "adrenalin", count: 4 },
  { kind: "buff", template_id: "linear_actuator", count: 2 },
  { kind: "buff", template_id: "processor_update", count: 2 },
];

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    ...(init?.headers as Record<string, string> | undefined),
  };

  if (init?.body !== undefined && !("Content-Type" in headers)) {
    headers["Content-Type"] = "application/json";
  }

  const requestInit: RequestInit = {
    credentials: "include",
    headers,
    ...init,
  };

  let response = await fetch(apiUrl(path), requestInit);

  if (response.status === 401 && path !== "/auth/telegram") {
    const initData = getTelegramInitData();
    if (initData) {
      const authResp = await fetch(apiUrl("/auth/telegram"), {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ initData }),
      });
      if (authResp.ok) {
        response = await fetch(apiUrl(path), requestInit);
      }
    }
  }

  if (!response.ok) {
    let message = `${response.status} ${response.statusText}`;
    try {
      const data = (await response.json()) as ApiError;
      if (data.error) {
        message = data.error;
      }
    } catch {
      // keep default message
    }
    throw new Error(message);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

function AssetImage(props: {
  imageKey?: string;
  alt: string;
  fallbackSrc: string;
  className: string;
}) {
  return (
    <img
      className={props.className}
      src={resolveImageSrc(props.imageKey)}
      alt={props.alt}
      loading="lazy"
      onError={(event) => {
        const target = event.currentTarget;
        if (target.dataset.fallbackApplied === "1") {
          return;
        }
        target.dataset.fallbackApplied = "1";
        target.src = props.fallbackSrc;
      }}
    />
  );
}

function totalDeck(entries: DeckEntry[]): number {
  return entries.reduce((sum, entry) => sum + entry.count, 0);
}

function cardInstanceId(card: CardsInMatch): string {
  return card.instance_id ?? card.InstanceID ?? "";
}

function cardTemplateId(card: CardsInMatch): string {
  return card.template_id ?? card.TemplateID ?? "";
}

function cardKind(card: CardsInMatch): string {
  return card.kind ?? card.Kind ?? "";
}

function unitTemplateId(unit: UnitState): string {
  return unit.template_id ?? unit.TemplateID ?? "";
}

function unitInstanceId(unit: UnitState): string {
  return unit.instance_id ?? unit.InstanceID ?? "";
}

function unitHP(unit: UnitState): number {
  return unit.hp ?? unit.HP ?? 0;
}

function unitMaxHP(unit: UnitState): number {
  return unit.max_hp ?? unit.MaxHP ?? 0;
}

function unitAttack(unit: UnitState): number {
  return unit.attack ?? unit.Attack ?? 0;
}

function unitCooldown(unit: UnitState): number {
  return unit.cooldown ?? unit.Cooldown ?? 0;
}

function unitBaseCooldown(unit: UnitState): number {
  return unit.base_cooldown ?? unit.BaseCooldown ?? 0;
}

function unitIsTank(unit: UnitState): boolean {
  return unit.is_tank ?? unit.IsTank ?? false;
}

function cardRaceLabel(cardType?: string): string {
  if (!cardType) {
    return "НЕИЗВЕСТНО";
  }
  const normalized = cardType.trim().toLowerCase();
  switch (normalized) {
    case "human":
    case "человек":
      return "ЛЮДИ";
    case "mechanical":
    case "машина":
      return "МЕХАНИКА";
    case "demonical":
    case "демонический":
    case "vespid":
    case "веспид":
      return "ВЕСПИДЫ";
    default:
      return cardType.toUpperCase();
  }
}

function toGameCardData(preview: CardPreview): GameCardData {
  return {
    kind: preview.kind,
    name: preview.name,
    description: preview.description,
    imageKey: preview.imageKey,
    race: preview.race,
    mana: preview.mana,
    attack: preview.attack,
    hp: preview.hp,
    cooldown: preview.cooldown,
    skillCooldown: preview.skillCooldown,
    buffType: preview.buffType,
    buffValue: preview.buffValue,
    duration: preview.duration,
  };
}

function unitSummonedInTurn(unit: UnitState): number {
  return unit.summoned_in_turn ?? unit.SummonedInTurn ?? -1;
}

function unitSkillName(unit: UnitState): string {
  return unit.skill_name ?? unit.SkillName ?? "";
}

function unitSkillCode(unit: UnitState): string {
  return unit.skill_code ?? unit.SkillCode ?? "";
}

function unitSkillTrigger(unit: UnitState): string {
  return unit.skill_trigger ?? unit.SkillTrigger ?? "";
}

function unitSkillTarget(unit: UnitState): string {
  return unit.skill_target ?? unit.SkillTarget ?? "";
}

function unitSkillCooldown(unit: UnitState): number {
  return unit.skill_cooldown ?? unit.SkillCooldown ?? 0;
}

function unitSkillCooldownLeft(unit: UnitState): number {
  return unit.skill_cooldown_left ?? unit.SkillCooldownLeft ?? 0;
}

function isMatchState(value: unknown): value is MatchState {
  return Boolean(
    value &&
      typeof value === "object" &&
      "match_id" in value &&
      "players" in value,
  );
}

function heroAbilityManaCost(player: MatchPlayer): number {
  return player.hero_ability_mana_cost ?? player.HeroAbilityManaCost ?? 0;
}

function ProfilePanel(props: {
  me: MeResponse | null;
  matches: MatchState[];
  onClose: () => void;
}) {
  const initials = (props.me?.first_name?.[0] || props.me?.username?.[0] || "?").toUpperCase();

  return (
    <div className="profile-overlay" onClick={props.onClose}>
      <aside className="profile-panel" onClick={(event) => event.stopPropagation()}>
        <button className="ghost-button" onClick={props.onClose}>
          Close
        </button>
        <div className="profile-avatar">{initials}</div>
        <h2>{props.me?.first_name || props.me?.username || "Unknown Soldier"}</h2>
        <p className="muted">@{props.me?.username || "no_username"}</p>
        <div className="profile-metrics">
          <div className="metric-tile">
            <span>Rating</span>
            <strong>{props.me?.rating ?? "-"}</strong>
          </div>
          <div className="metric-tile">
            <span>User ID</span>
            <strong>{props.me?.user_id ?? "-"}</strong>
          </div>
        </div>
        <div className="profile-history">
          <h3>Match History</h3>
          {props.matches.length === 0 ? (
            <p className="muted">No battles yet.</p>
          ) : (
            props.matches.map((match) => (
              <div key={match.match_id} className="history-row">
                <span>#{match.match_id}</span>
                <span>{match.result}</span>
                <span>v{match.version}</span>
              </div>
            ))
          )}
        </div>
      </aside>
    </div>
  );
}

export default function App() {
  if (import.meta.env.DEV) {
    const cardPreview: CardPreview = {
      kind: "battle",
      name: "ОМНИЦИДЫ",
      description: "Наносит периодический урон противнику",
      imageKey: "cards/battle/omnicides/image",
      race: "ВЕСПИДЫ",
      mana: 1,
      attack: 3,
      hp: 3,
      cooldown: 0,
      skillCooldown: 3,
    };

    return (
      <div className="card-viewer-overlay">
        <div className="card-viewer-window">
          <GameCard data={toGameCardData(cardPreview)} mode="viewer" />
        </div>
      </div>
    );
  }

  const [tab, setTab] = useState<TabId>("home");
  const [loading, setLoading] = useState(false);
  const [showProfile, setShowProfile] = useState(false);
  const [toasts, setToasts] = useState<ToastEntry[]>([]);

  const [me, setMe] = useState<MeResponse | null>(null);
  const [heroes, setHeroes] = useState<OwnedHero[]>([]);
  const [cards, setCards] = useState<CardsResponse | null>(null);
  const [deckEntries, setDeckEntries] = useState<DeckEntry[]>(defaultDeck);
  const [deckInspectorKey, setDeckInspectorKey] = useState<string | null>(null);
  const [deckOverviewOpen, setDeckOverviewOpen] = useState(false);
  const [heroPickerOpen, setHeroPickerOpen] = useState(false);
  const [heldHero, setHeldHero] = useState<OwnedHero | null>(null);
  const [cardPreview, setCardPreview] = useState<CardPreview | null>(null);
  const [catalogKind, setCatalogKind] = useState<CatalogKind>("battle");
  const [catalogSort, setCatalogSort] = useState<CatalogSort>("mana");
  const [catalogPage, setCatalogPage] = useState(0);
  const [matches, setMatches] = useState<MatchState[]>([]);
  const [selectedMatchId, setSelectedMatchId] = useState<number | null>(null);
  const [selectedMatch, setSelectedMatch] = useState<MatchState | null>(null);

  const [actionStatus, setActionStatus] = useState("Select a card, then issue an order.");
  const [selectedHandCardId, setSelectedHandCardId] = useState("");
  const [selectedOwnUnitId, setSelectedOwnUnitId] = useState("");
  const [selectedEnemyUnitId, setSelectedEnemyUnitId] = useState("");
  const [selectedSkillCasterId, setSelectedSkillCasterId] = useState("");
  const [heroSpellArmed, setHeroSpellArmed] = useState(false);
  const [heroAttackArmed, setHeroAttackArmed] = useState(false);
  const [openedGraveSide, setOpenedGraveSide] = useState<"own" | "enemy" | null>(null);
  const [eventQueue, setEventQueue] = useState<MatchEvent[]>([]);
  const [activeEvent, setActiveEvent] = useState<MatchEvent | null>(null);
  const [boardAttackAnimation, setBoardAttackAnimation] = useState<BoardAttackAnimation | null>(null);
  const [miniAppFullscreen, setMiniAppFullscreen] = useState(false);
  const [queuePanelOpen, setQueuePanelOpen] = useState(false);
  const [selectedQueueMode, setSelectedQueueMode] = useState<MatchmakingMode | null>(null);
  const [queueStatus, setQueueStatus] = useState<QueueStatusResponse>({ state: "idle" });

  const streamRef = useRef<EventSource | null>(null);
  const battleBoardRef = useRef<HTMLElement | null>(null);
  const [dragAttack, setDragAttack] = useState<DragAttackState | null>(null);
  const toastIdRef = useRef(1);
  const [ownHeroHpPeak, setOwnHeroHpPeak] = useState(0);
  const [enemyHeroHpPeak, setEnemyHeroHpPeak] = useState(0);
  const [clockTickMs, setClockTickMs] = useState(() => Date.now());
  const [serverClockOffsetSec, setServerClockOffsetSec] = useState(0);
  const [drawFxTick, setDrawFxTick] = useState(0);
  const prevHandCountRef = useRef<number | null>(null);
  const eventPlaybackTimerRef = useRef<number | null>(null);
  const queuePrimedRef = useRef(false);
  const lastQueuedBatchRef = useRef<string>("");
  const previousQueueStateRef = useRef<MatchmakingState>("idle");

  function pushToast(message: string, tone: ToastEntry["tone"] = "info") {
    const id = toastIdRef.current++;
    setToasts((prev) => [...prev, { id, message, tone }]);
    window.setTimeout(() => {
      setToasts((prev) => prev.filter((entry) => entry.id !== id));
    }, 1200);
  }

  useEffect(() => {
    bootstrapTelegramWebApp();
    setMiniAppFullscreen(isMiniAppFullscreen());
  }, []);

  useEffect(() => {
    const syncFullscreen = () => setMiniAppFullscreen(isMiniAppFullscreen());
    document.addEventListener("fullscreenchange", syncFullscreen);
    window.addEventListener("resize", syncFullscreen);
    return () => {
      document.removeEventListener("fullscreenchange", syncFullscreen);
      window.removeEventListener("resize", syncFullscreen);
    };
  }, []);

  useEffect(() => {
    if (!selectedMatchId) {
      streamRef.current?.close();
      streamRef.current = null;
      setEventQueue([]);
      setActiveEvent(null);
      setBoardAttackAnimation(null);
      queuePrimedRef.current = false;
      lastQueuedBatchRef.current = "";
      return;
    }

    const stream = new EventSource(apiUrl(`/matches/${selectedMatchId}/stream`), {
      withCredentials: true,
    });
    stream.addEventListener("state", (event) => {
      try {
        const state = JSON.parse((event as MessageEvent<string>).data) as MatchState;
        ingestMatchState(state, "queue");
      } catch {
        pushToast("Failed to parse battle feed", "error");
      }
    });
    stream.onerror = () => {
      pushToast("Battle feed disconnected", "error");
    };
    streamRef.current = stream;
    return () => stream.close();
  }, [selectedMatchId]);

  useEffect(() => {
    queuePrimedRef.current = false;
    lastQueuedBatchRef.current = "";
    setEventQueue([]);
    setActiveEvent(null);
    setBoardAttackAnimation(null);
    if (eventPlaybackTimerRef.current) {
      window.clearTimeout(eventPlaybackTimerRef.current);
      eventPlaybackTimerRef.current = null;
    }
  }, [selectedMatch?.match_id]);

  const myPlayer = useMemo(
    () => selectedMatch?.players.find((player) => player?.user_id === me?.user_id) ?? null,
    [selectedMatch, me],
  );
  const enemyPlayer = useMemo(
    () => selectedMatch?.players.find((player) => player?.user_id !== me?.user_id) ?? null,
    [selectedMatch, me],
  );
  const activeBattle = Boolean(selectedMatch && !selectedMatch.finished);
  const ownHeroEventId = myPlayer ? `hero:p${myPlayer.player_id}` : "";
  const enemyHeroEventId = enemyPlayer ? `hero:p${enemyPlayer.player_id}` : "";
  const isMyTurn = Boolean(
    selectedMatch &&
      myPlayer &&
      selectedMatch.active_player === myPlayer.player_id,
  );
  const myHandCount = myPlayer?.hand?.length ?? myPlayer?.hand_count ?? 0;
  const selectedSkillCaster = useMemo(() => {
    if (!myPlayer || !selectedSkillCasterId) {
      return null;
    }
    return (
      myPlayer.table.find((entry) => entry && unitInstanceId(entry) === selectedSkillCasterId) ??
      null
    );
  }, [myPlayer, selectedSkillCasterId]);
  useEffect(() => {
    if (!selectedSkillCasterId) {
      return;
    }
    if (!selectedSkillCaster) {
      setSelectedSkillCasterId("");
    }
  }, [selectedSkillCasterId, selectedSkillCaster]);
  useEffect(() => {
    if (!activeBattle) {
      return;
    }
    const timerId = window.setInterval(() => {
      setClockTickMs(Date.now());
    }, 1000);
    return () => window.clearInterval(timerId);
  }, [activeBattle]);
  useEffect(() => {
    if (!selectedMatch?.server_now) {
      return;
    }
    const nowSec = Math.floor(Date.now() / 1000);
    setServerClockOffsetSec(selectedMatch.server_now - nowSec);
  }, [selectedMatch?.server_now, selectedMatch?.match_id]);

  useEffect(() => {
    if (!activeBattle || !myPlayer) {
      prevHandCountRef.current = null;
      return;
    }
    const prev = prevHandCountRef.current;
    if (prev === null) {
      prevHandCountRef.current = myHandCount;
      return;
    }
    if (myHandCount > prev) {
      setDrawFxTick((value) => value + 1);
    }
    prevHandCountRef.current = myHandCount;
  }, [activeBattle, myPlayer, myHandCount]);

  useEffect(() => {
    if (activeEvent || eventQueue.length === 0) {
      return;
    }
    const nextEvent = eventQueue[0];
    setActiveEvent(nextEvent);
    setBoardAttackAnimation(buildBoardAttackAnimation(nextEvent));
    const duration = eventDurationMs(nextEvent);
    eventPlaybackTimerRef.current = window.setTimeout(() => {
      setBoardAttackAnimation(null);
      setActiveEvent(null);
      setEventQueue((prev) => prev.slice(1));
      eventPlaybackTimerRef.current = null;
    }, duration);
  }, [activeEvent, eventQueue]);

  useEffect(() => {
    return () => {
      if (eventPlaybackTimerRef.current) {
        window.clearTimeout(eventPlaybackTimerRef.current);
        eventPlaybackTimerRef.current = null;
      }
    };
  }, []);
  const clientNowSec = Math.floor(clockTickMs / 1000);
  const syncedNowSec = clientNowSec + serverClockOffsetSec;
  const turnTotalSec = Math.max(1, selectedMatch?.turn_time_sec ?? 45);
  const turnDeadlineAt = (() => {
    const explicitDeadline = selectedMatch?.turn_deadline_at ?? 0;
    if (explicitDeadline > 0) {
      return explicitDeadline;
    }
    const startedAt = selectedMatch?.turn_started_at ?? 0;
    if (startedAt > 0) {
      return startedAt + turnTotalSec;
    }
    return 0;
  })();
  const hasTurnTimer = Boolean(activeBattle && (turnDeadlineAt > 0 || selectedMatch?.phase === "MAIN"));
  const turnSecondsLeft = turnDeadlineAt > 0 ? Math.max(0, turnDeadlineAt - syncedNowSec) : 0;
  const turnProgress = hasTurnTimer && turnDeadlineAt > 0 ? Math.max(0, Math.min(1, turnSecondsLeft / turnTotalSec)) : 0;
  useEffect(() => {
    if (!dragAttack) {
      return;
    }

    function updateDragPosition(clientX: number, clientY: number) {
      const rect = battleBoardRef.current?.getBoundingClientRect();
      if (!rect) {
        return;
      }
      setDragAttack((prev) =>
        prev
          ? {
              ...prev,
              currentX: clientX - rect.left,
              currentY: clientY - rect.top,
            }
          : null,
      );
    }

    function handlePointerMove(event: globalThis.PointerEvent) {
      updateDragPosition(event.clientX, event.clientY);
    }

    function handlePointerUp(event: globalThis.PointerEvent) {
      const target = document.elementFromPoint(event.clientX, event.clientY)?.closest<HTMLElement>(
        "[data-attack-target]",
      );
      const targetKind = target?.dataset.attackTarget;
      const targetUnitId = target?.dataset.unitId;

      if (targetKind === "enemy-hero") {
        void runTask(handleEnemyHeroClick);
      } else if (targetKind === "enemy-unit" && targetUnitId) {
        const unit = enemyPlayer?.table.find((entry) => entry && unitInstanceId(entry) === targetUnitId);
        if (unit) {
          void runTask(() => handleEnemyUnitClick(unit));
        }
      }

      setDragAttack(null);
    }

    function handlePointerCancel() {
      setDragAttack(null);
    }

    window.addEventListener("pointermove", handlePointerMove);
    window.addEventListener("pointerup", handlePointerUp);
    window.addEventListener("pointercancel", handlePointerCancel);

    return () => {
      window.removeEventListener("pointermove", handlePointerMove);
      window.removeEventListener("pointerup", handlePointerUp);
      window.removeEventListener("pointercancel", handlePointerCancel);
    };
  }, [dragAttack, enemyPlayer]);

  function getBoardRelativePoint(clientX: number, clientY: number): { x: number; y: number } | null {
    const rect = battleBoardRef.current?.getBoundingClientRect();
    if (!rect) {
      return null;
    }
    return { x: clientX - rect.left, y: clientY - rect.top };
  }

  function getElementCenterInBoard(el: HTMLElement): { x: number; y: number } | null {
    const boardRect = battleBoardRef.current?.getBoundingClientRect();
    if (!boardRect) {
      return null;
    }
    const rect = el.getBoundingClientRect();
    return {
      x: rect.left + rect.width / 2 - boardRect.left,
      y: rect.top + rect.height / 2 - boardRect.top,
    };
  }

  function sideForPlayerIndex(playerIndex: number): "own" | "enemy" {
    return myPlayer && playerIndex === myPlayer.player_id ? "own" : "enemy";
  }

  function heroInstanceIdForPlayerIndex(playerIndex: number): string {
    return `hero:p${playerIndex}`;
  }

  function pointForUnitInstance(instanceId: string): { x: number; y: number } | null {
    if (!battleBoardRef.current || !instanceId) {
      return null;
    }
    const node = battleBoardRef.current.querySelector<HTMLElement>(`[data-unit-id="${instanceId}"]`);
    return node ? getElementCenterInBoard(node) : null;
  }

  function pointForHeroSide(side: "own" | "enemy"): { x: number; y: number } | null {
    if (!battleBoardRef.current) {
      return null;
    }
    const node = battleBoardRef.current.querySelector<HTMLElement>(`[data-hero-side="${side}"]`);
    return node ? getElementCenterInBoard(node) : null;
  }

  function eventDurationMs(event: MatchEvent): number {
    switch (event.type) {
      case "attack":
      case "hero_attack":
        return 460;
      case "card_skill":
      case "hero_spell":
        return 880;
      case "death":
        return 520;
      default:
        return 620;
    }
  }

  function buildBoardAttackAnimation(event: MatchEvent): BoardAttackAnimation | null {
    if (!event.player_index) {
      if (event.player_index !== 0) {
        return null;
      }
    }
    if (event.type !== "attack" && event.type !== "hero_attack") {
      return null;
    }
    const firstTargetId = event.targets?.[0]?.instance_id ?? "";
    if (!firstTargetId) {
      return null;
    }

    let sourcePoint: { x: number; y: number } | null = null;
    if (event.source_kind === "unit" && event.source_instance_id) {
      sourcePoint = pointForUnitInstance(event.source_instance_id);
    } else if (event.source_kind === "hero") {
      sourcePoint = pointForHeroSide(sideForPlayerIndex(event.player_index ?? 0));
    }

    let targetPoint: { x: number; y: number } | null = null;
    if (firstTargetId.startsWith("hero:p")) {
      const heroPlayerIndex = Number(firstTargetId.replace("hero:p", ""));
      targetPoint = pointForHeroSide(sideForPlayerIndex(heroPlayerIndex));
    } else {
      targetPoint = pointForUnitInstance(firstTargetId);
    }

    if (!sourcePoint || !targetPoint) {
      return null;
    }

    const dx = (targetPoint.x - sourcePoint.x) * 0.36;
    const dy = (targetPoint.y - sourcePoint.y) * 0.36;
    return {
      sourceKind: event.source_kind === "hero" ? "hero" : "unit",
      sourceInstanceId: event.source_instance_id,
      sourceSide: sideForPlayerIndex(event.player_index ?? 0),
      dx,
      dy,
      targetIds: (event.targets ?? []).map((target) => target.instance_id ?? "").filter(Boolean),
    };
  }

  function eventSourceImageKey(event: MatchEvent): string {
    if (event.source_kind === "hero" && event.source_hero_code) {
      return resolveHeroImageKey(event.source_hero_code);
    }
    const templateId = event.source_template_id ?? event.source_card_template_id ?? "";
    return templateId ? cardImageKeyForTemplate(templateId) : resolveCardFallbackSrc();
  }

  function eventSourceLabel(event: MatchEvent): string {
    if (event.source_kind === "hero" && event.source_hero_code) {
      return resolveAssetLabel(event.source_hero_code);
    }
    const templateId = event.source_template_id ?? event.source_card_template_id ?? "";
    return templateId ? resolveAssetLabel(templateId) : "Battle Action";
  }

  function eventTitle(event: MatchEvent): string {
    switch (event.type) {
      case "attack":
        return "Attack";
      case "hero_attack":
        return "Hero Attack";
      case "card_skill":
        return "Card Skill";
      case "hero_spell":
        return "Hero Skill";
      case "summon":
        return "Summon";
      case "buff":
        return "Buff";
      case "heal":
        return "Heal";
      case "death":
        return "Death";
      case "resurrect":
        return "Resurrect";
      default:
        return event.type || "Action";
    }
  }

  function eventBatchKey(state: MatchState | null): string {
    if (!state) {
      return "";
    }
    return `${state.match_id}:${state.version}:${JSON.stringify(state.events ?? [])}`;
  }

  function ingestMatchState(state: MatchState | null, mode: "prime" | "queue" = "queue") {
    setSelectedMatch(state);
    if (!state) {
      return;
    }
    const batchKey = eventBatchKey(state);
    if (!queuePrimedRef.current || mode === "prime") {
      queuePrimedRef.current = true;
      lastQueuedBatchRef.current = batchKey;
      return;
    }
    if (lastQueuedBatchRef.current === batchKey) {
      return;
    }
    lastQueuedBatchRef.current = batchKey;
    if ((state.events?.length ?? 0) > 0) {
      setEventQueue((prev) => [...prev, ...(state.events ?? [])]);
    }
  }

  const cardCatalog = useMemo(() => {
    const next = new Map<string, CardCatalogEntry>();
    for (const card of cards?.battle ?? []) {
      next.set(card.template_id, card);
    }
    for (const card of cards?.buff ?? []) {
      next.set(card.template_id, card);
    }
    return next;
  }, [cards]);

  async function runTask(task: () => Promise<void>) {
    setLoading(true);
    try {
      await task();
    } catch (error) {
      pushToast(error instanceof Error ? error.message : "Unknown error", "error");
    } finally {
      setLoading(false);
    }
  }

  async function refreshMe() {
    setMe(await apiFetch<MeResponse>("/me"));
  }

  async function refreshHeroes() {
    const data = await apiFetch<{ heroes: OwnedHero[] }>("/heroes");
    setHeroes(data.heroes);
  }

  async function refreshCards() {
    setCards(await apiFetch<CardsResponse>("/cards"));
  }

  async function refreshDeck() {
    const data = await apiFetch<DeckResponse>("/deck");
    setDeckEntries(data.entries);
  }

  async function refreshMatches() {
    const data = await apiFetch<MatchState[]>("/matches");
    setMatches(data);
    const active = data.find((match) => !match.finished) ?? null;
    if (selectedMatchId) {
      const current = data.find((match) => match.match_id === selectedMatchId) ?? null;
      ingestMatchState(current, "prime");
      if (!current || current.finished) {
        setSelectedMatchId(null);
      }
      return;
    }
    if (active) {
      setSelectedMatchId(active.match_id);
      ingestMatchState(active, "prime");
      pushToast(`Battle #${active.match_id} ready`);
    }
  }

  async function refreshQueueStatus() {
    try {
      const data = await apiFetch<QueueStatusResponse>("/queue/status");
      setQueueStatus(data);
    } catch {
      setQueueStatus({ state: "idle" });
    }
  }

  async function refreshMatch(matchId: number) {
    const data = await apiFetch<MatchState>(`/matches/${matchId}`);
    if (data.finished) {
      setSelectedMatchId(null);
      ingestMatchState(null, "prime");
      pushToast(`Battle #${matchId} already finished`);
      return;
    }
    setSelectedMatchId(matchId);
    ingestMatchState(data, "prime");
  }

  async function refreshAll() {
    await Promise.all([
      refreshMe(),
      refreshHeroes(),
      refreshCards(),
      refreshDeck(),
      refreshMatches(),
      refreshQueueStatus(),
    ]);
  }

  function isUnauthorizedError(error: unknown): boolean {
    if (!(error instanceof Error)) {
      return false;
    }
    const message = error.message.toLowerCase();
    return message.includes("unauthorized") || message.includes("401");
  }

  useEffect(() => {
    let cancelled = false;

    async function bootstrapAuth() {
      try {
        await refreshAll();
        return;
      } catch (error) {
        if (!isUnauthorizedError(error)) {
          pushToast(error instanceof Error ? error.message : "Failed to initialize profile", "error");
          return;
        }
      }

      const initData = getTelegramInitData();
      if (!initData) {
        return;
      }

      try {
        await apiFetch<void>("/auth/telegram", {
          method: "POST",
          body: JSON.stringify({ initData }),
        });
        if (cancelled) {
          return;
        }
        await refreshAll();
        if (!cancelled) {
          pushToast("Logged in via Telegram");
        }
      } catch (error) {
        if (!cancelled) {
          pushToast(error instanceof Error ? error.message : "Telegram auth failed", "error");
        }
      }
    }

    void bootstrapAuth();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (activeBattle) {
      return;
    }

    let cancelled = false;
    const syncQueue = async () => {
      try {
        const data = await apiFetch<QueueStatusResponse>("/queue/status");
        if (!cancelled) {
          setQueueStatus(data);
        }
      } catch {
        if (!cancelled) {
          setQueueStatus({ state: "idle" });
        }
      }
    };

    void syncQueue();
    const timerId = window.setInterval(() => {
      void syncQueue();
    }, 1000);

    return () => {
      cancelled = true;
      window.clearInterval(timerId);
    };
  }, [activeBattle]);

  useEffect(() => {
    if (queueStatus.state === "searching" || queueStatus.state === "pending_match") {
      setQueuePanelOpen(false);
    }
  }, [queueStatus.state]);

  useEffect(() => {
    const previous = previousQueueStateRef.current;
    previousQueueStateRef.current = queueStatus.state;
    if (activeBattle) {
      return;
    }
    if (previous === "pending_match" && queueStatus.state === "idle") {
      void refreshMatches();
    }
  }, [activeBattle, queueStatus.state]);

  useEffect(() => {
    if (!me || activeBattle) {
      return;
    }
    const pollId = window.setInterval(() => {
      void refreshMatches();
    }, 3000);
    return () => window.clearInterval(pollId);
  }, [me, activeBattle]);

  useEffect(() => {
    if (!myPlayer) {
      setOwnHeroHpPeak(0);
      return;
    }
    setOwnHeroHpPeak((prev) => Math.max(prev, myPlayer.hero_hp));
  }, [myPlayer]);

  useEffect(() => {
    if (!enemyPlayer) {
      setEnemyHeroHpPeak(0);
      return;
    }
    setEnemyHeroHpPeak((prev) => Math.max(prev, enemyPlayer.hero_hp));
  }, [enemyPlayer]);

  async function selectHero(heroCode: string) {
    await apiFetch("/heroes/select", {
      method: "POST",
      body: JSON.stringify({ hero_code: heroCode }),
    });
    await Promise.all([refreshMe(), refreshHeroes()]);
    pushToast(`Hero selected: ${heroCode}`);
  }

  async function saveDefaultDeck() {
    await apiFetch("/deck", {
      method: "POST",
      body: JSON.stringify({ entries: defaultDeck }),
    });
    setDeckEntries(defaultDeck);
    setDeckInspectorKey(null);
    pushToast("Standard combat deck loaded");
  }

  async function persistDeck(entries: DeckEntry[]) {
    const cleaned = entries.filter((entry) => entry.count > 0);
    setDeckEntries(cleaned);
    if (totalDeck(cleaned) !== 20) {
      return;
    }
    await apiFetch("/deck", {
      method: "POST",
      body: JSON.stringify({ entries: cleaned }),
    });
  }

  async function retryTelegramAuth() {
    const initData = getTelegramInitData();
    if (!initData) {
      throw new Error("Open game from Telegram bot");
    }
    await apiFetch<void>("/auth/telegram", {
      method: "POST",
      body: JSON.stringify({ initData }),
    });
    await refreshAll();
    pushToast("Telegram auth restored");
  }

  function cardPoolInfo(kind: DeckEntry["kind"], templateId: string) {
    if (kind === "battle") {
      const card = cards?.battle.find((entry) => entry.template_id === templateId);
      return {
        owned: card?.copies ?? 0,
        maxCopies: card?.max_copies ?? 0,
      };
    }
    const card = cards?.buff.find((entry) => entry.template_id === templateId);
    return {
      owned: card?.copies ?? 0,
      maxCopies: card?.max_copies ?? 0,
    };
  }

  async function addCardToDeck(kind: DeckEntry["kind"], templateId: string) {
    const total = totalDeck(deckEntries);
    if (total >= 20) {
      pushToast("Deck is full (20 cards)", "error");
      return;
    }

    const { owned, maxCopies } = cardPoolInfo(kind, templateId);
    const limit = Math.min(owned, maxCopies);
    if (limit <= 0) {
      pushToast("No copies available", "error");
      return;
    }

    const current = deckEntries.find((entry) => entry.kind === kind && entry.template_id === templateId);
    const currentCount = current?.count ?? 0;
    if (currentCount >= limit) {
      pushToast("Copy limit reached", "error");
      return;
    }

    const next = current
      ? deckEntries.map((entry) =>
          entry.kind === kind && entry.template_id === templateId
            ? { ...entry, count: entry.count + 1 }
            : entry,
        )
      : [...deckEntries, { kind, template_id: templateId, count: 1 }];

    await persistDeck(next);
  }

  async function removeCardFromDeck(kind: DeckEntry["kind"], templateId: string) {
    const current = deckEntries.find((entry) => entry.kind === kind && entry.template_id === templateId);
    if (!current) {
      return;
    }

    const next = deckEntries
      .map((entry) =>
        entry.kind === kind && entry.template_id === templateId
          ? { ...entry, count: Math.max(0, entry.count - 1) }
          : entry,
      )
      .filter((entry) => entry.count > 0);

    await persistDeck(next);
    const remaining = next.find((entry) => entry.kind === kind && entry.template_id === templateId);
    if (!remaining) {
      setDeckInspectorKey(null);
    }
  }

  async function joinMatchmakingQueue() {
    if (!selectedQueueMode) {
      pushToast("Select a mode first", "error");
      return;
    }

    const next = await apiFetch<QueueStatusResponse>("/queue/join", {
      method: "POST",
    });
    setQueueStatus(next);
    setQueuePanelOpen(false);
    await refreshQueueStatus();
  }

  async function leaveMatchmakingQueue() {
    const next = await apiFetch<QueueStatusResponse>("/queue/leave", {
      method: "POST",
    });
    setQueueStatus(next);
    setQueuePanelOpen(false);
  }

  async function acceptMatchmakingReady() {
    const next = await apiFetch<AcceptQueueResponse | MatchState>("/queue/accept", {
      method: "POST",
    });

    if (isMatchState(next)) {
      setQueueStatus({ state: "idle" });
      setSelectedMatchId(next.match_id);
      ingestMatchState(next, "prime");
      pushToast("Match accepted. Entering battle.");
      return;
    }

    setQueueStatus((prev) => ({
      ...prev,
      ...next,
    }));
    await refreshQueueStatus();
  }

  async function declineMatchmakingReady() {
    const next = await apiFetch<DeclineQueueResponse>("/queue/decline", {
      method: "POST",
    });
    setQueueStatus((prev) => ({
      ...prev,
      ...next,
    }));
    await refreshQueueStatus();
  }

  async function handleToggleMiniAppFullscreen() {
    const next = miniAppFullscreen ? await exitMiniAppFullscreen() : await requestMiniAppFullscreen();
    if (!next && !miniAppFullscreen) {
      pushToast("Fullscreen is not supported here", "error");
    }
    setMiniAppFullscreen(isMiniAppFullscreen());
  }

  function targetLabel(target: string): string {
    switch (target) {
      case TARGET_SELF:
        return "self";
      case TARGET_ALLY_UNIT:
        return "ally unit";
      case TARGET_ENEMY_UNIT:
        return "enemy unit";
      case TARGET_ALLY_ALL:
        return "all allies";
      case TARGET_ENEMY_ALL:
        return "all enemies";
      case TARGET_BOTH_ALL:
        return "all units";
      case TARGET_ENEMY_SPLASH:
        return "enemy splash center";
      case TARGET_ALLY_SPLASH:
        return "ally splash center";
      case TARGET_ALLY_GRAVE_SINGLE:
        return "ally grave card";
      case TARGET_NONE:
      default:
        return "no target";
    }
  }

  function heroSpellTargetMode(heroCode: string): HeroSpellTargetMode {
    if (heroCode === "imperial_commander" || heroCode === "black_cell" || heroCode === "karn" || heroCode === "slavic_priest") {
      return "own-unit";
    }
    if (heroCode === "the_system") {
      return "enemy-unit";
    }
    if (heroCode === "suprime_lider") {
      return "enemy-any";
    }
    return "enemy-hero-only";
  }

  function canSelectUnitAsHeroSpellTarget(side: "own" | "enemy"): boolean {
    if (!heroSpellArmed || !myPlayer) {
      return false;
    }
    const mode = heroSpellTargetMode(myPlayer.hero_code);
    if (side === "own") {
      return mode === "own-unit";
    }
    return mode === "enemy-unit" || mode === "enemy-any";
  }

  function canSelectEnemyHeroAsHeroSpellTarget(): boolean {
    if (!heroSpellArmed || !myPlayer) {
      return false;
    }
    const mode = heroSpellTargetMode(myPlayer.hero_code);
    return mode === "enemy-any" || mode === "enemy-hero-only";
  }

  function canCardSkillTargetEnemyHero(caster: UnitState): boolean {
    const target = unitSkillTarget(caster);
    const code = unitSkillCode(caster);
    return target === TARGET_ENEMY_UNIT && code === SKILL_DAMAGE_SINGLE;
  }

  function canSelectUnitAsSkillTarget(caster: UnitState, side: "own" | "enemy", unit: UnitState): boolean {
    const target = unitSkillTarget(caster);
    if (side === "own") {
      if (target === TARGET_SELF) {
        return unitInstanceId(unit) === unitInstanceId(caster);
      }
      return target === TARGET_ALLY_UNIT || target === TARGET_ALLY_SPLASH;
    }
    return target === TARGET_ENEMY_UNIT || target === TARGET_ENEMY_SPLASH;
  }

  async function castSkill(casterId: string, targetInstanceId = "", attackHero = false) {
    await applyAction(
      {
        type: "card_skill",
        card_instance_id: casterId,
        target_instance_id: targetInstanceId,
        attack_hero: attackHero,
      },
      "Card skill activated",
    );
  }

  async function startSkillCast(unit: UnitState) {
    const casterId = unitInstanceId(unit);
    const meta = cardCatalogEntry(unitTemplateId(unit));
    const fallback = skillFallbackByTemplate[unitTemplateId(unit)];
    const skillCode = unitSkillCode(unit) || meta?.skill_code || fallback?.code || "";
    const skillTrigger = unitSkillTrigger(unit) || meta?.skill_trigger || fallback?.trigger || "";
    const skillTarget = unitSkillTarget(unit) || meta?.skill_target || fallback?.target || TARGET_NONE;
    const skillName = unitSkillName(unit) || meta?.skill_name || fallback?.name || skillCode;
    if (!skillCode || skillTrigger !== SKILL_TRIGGER_ACTIVE) {
      pushToast("This unit has no active skill", "error");
      return;
    }
    if (!isMyTurn) {
      pushToast("Wait for your turn", "error");
      return;
    }
    if (unitSkillCooldownLeft(unit) > 0) {
      pushToast(`Skill on cooldown (${unitSkillCooldownLeft(unit)})`, "error");
      return;
    }
    const target = skillTarget;
    if (target === TARGET_ALLY_GRAVE_SINGLE) {
      pushToast("Grave target picker is not in UI yet", "error");
      return;
    }
    if (target === TARGET_NONE || target === TARGET_ALLY_ALL || target === TARGET_ENEMY_ALL || target === TARGET_BOTH_ALL) {
      await castSkill(casterId);
      return;
    }
    if (target === TARGET_SELF) {
      await castSkill(casterId, casterId);
      return;
    }
    setDragAttack(null);
    setHeroSpellArmed(false);
    setHeroAttackArmed(false);
    setSelectedSkillCasterId(casterId);
    setSelectedOwnUnitId("");
    setSelectedEnemyUnitId("");
    setActionStatus(`Skill mode: ${skillName || resolveAssetLabel(unitTemplateId(unit))}. Pick ${targetLabel(target)}.`);
  }

  function clearSelections() {
    setSelectedHandCardId("");
    setSelectedOwnUnitId("");
    setSelectedEnemyUnitId("");
    setSelectedSkillCasterId("");
    setHeroSpellArmed(false);
    setHeroAttackArmed(false);
    setDragAttack(null);
  }

  async function applyAction(payload: Record<string, unknown>, successText: string) {
    if (!selectedMatch) {
      pushToast("No match selected", "error");
      return;
    }
    const next = await apiFetch<MatchState>(`/matches/${selectedMatch.match_id}/actions`, {
      method: "POST",
      body: JSON.stringify({
        card_instance_id: "",
        target_instance_id: "",
        attack_hero: false,
        target_slot: 0,
        ...payload,
        expected_version: selectedMatch.version,
      }),
    });
    ingestMatchState(next, "queue");
    clearSelections();
    setActionStatus(successText);
    await refreshMatches();
    if (next.finished) {
      streamRef.current?.close();
      streamRef.current = null;
      setSelectedMatchId(null);
      pushToast(`Battle finished: ${next.result}`);
    }
  }

  const selectedCard = myPlayer?.hand?.find(
    (card) => cardInstanceId(card) === selectedHandCardId,
  );
  const myHand = myPlayer?.hand ?? [];
  const ownGraveyard = useMemo(
    () => myPlayer?.graveyard ?? myPlayer?.GraveYard ?? [],
    [myPlayer],
  );
  const enemyGraveyard = useMemo(
    () => enemyPlayer?.graveyard ?? enemyPlayer?.GraveYard ?? [],
    [enemyPlayer],
  );
  const openedGraveyard =
    openedGraveSide === "own" ? ownGraveyard : openedGraveSide === "enemy" ? enemyGraveyard : [];
  const displayedDeckCount = myPlayer?.deck?.length ?? myPlayer?.deck_count ?? 0;
  const enemyDisplayedDeckCount = enemyPlayer?.deck?.length ?? enemyPlayer?.deck_count ?? 0;

  function cardCatalogEntry(templateId: string): CardCatalogEntry | undefined {
    return cardCatalog.get(templateId);
  }

  function cardImageKeyForTemplate(templateId: string): string {
    const meta = cardCatalogEntry(templateId);
    if (meta?.image_key) {
      return meta.image_key;
    }
    if (meta?.kind === "buff") {
      return resolveBuffCardImageKey(templateId);
    }
    return resolveBattleCardImageKey(templateId);
  }

  function renderGraveEntry(entry: GraveEntryState, index: number) {
    const unit = entry.unit ?? entry.Unit ?? null;
    if (!unit) {
      return null;
    }
    const templateId = unitTemplateId(unit);
    const meta = cardCatalogEntry(templateId);
    const diedAtTurn = entry.died_at_turn ?? entry.DiedAtTurn ?? 0;
    return (
      <div key={`${unitInstanceId(unit) || templateId}-${index}`} className="grave-card-row">
        <AssetImage
          imageKey={cardImageKeyForTemplate(templateId)}
          alt={resolveAssetLabel(templateId)}
          fallbackSrc={resolveCardFallbackSrc()}
          className="grave-card-thumb"
        />
        <div className="grave-card-copy">
          <strong>{meta?.name || resolveAssetLabel(templateId)}</strong>
          <span>{`HP ${unitHP(unit)} | ATK ${unitAttack(unit)} | CD ${unitCooldown(unit)}`}</span>
          <span className="grave-card-turn">{`Turn ${diedAtTurn}`}</span>
        </div>
      </div>
    );
  }

  async function handlePlaySelectedCard(slot: number) {
    if (!selectedCard) {
      if (selectedHandCardId || selectedOwnUnitId || selectedEnemyUnitId || selectedSkillCasterId || heroSpellArmed || heroAttackArmed) {
        clearSelections();
        setActionStatus("Selection cleared");
        return;
      }
      pushToast("Select a card from hand first", "error");
      return;
    }
    const type = cardKind(selectedCard) === "buff" ? "play_buff_card" : "play_battle_card";
    const payload =
      type === "play_buff_card"
        ? {
            type,
            card_instance_id: cardInstanceId(selectedCard),
            target_instance_id: selectedOwnUnitId,
          }
        : {
            type,
            card_instance_id: cardInstanceId(selectedCard),
            target_slot: slot,
          };
    await applyAction(payload, `Action sent: ${type}`);
  }

  async function handleEndTurn() {
    await applyAction({ type: "end_turn" }, "Turn ended");
  }

  async function handleHeroSpell() {
    if (!selectedMatch || !myPlayer) {
      pushToast("No battle selected", "error");
      return;
    }
    if (!isMyTurn) {
      pushToast("Wait for your turn", "error");
      return;
    }
    if (myPlayer.hero_ability_cooldown > 0) {
      pushToast(`Hero skill on cooldown (${myPlayer.hero_ability_cooldown})`, "error");
      return;
    }
    if (heroSpellArmed) {
      setHeroSpellArmed(false);
      setActionStatus("Hero skill selection canceled");
      return;
    }
    const mode = heroSpellTargetMode(myPlayer.hero_code);
    setHeroSpellArmed(true);
    setHeroAttackArmed(false);
    setDragAttack(null);
    setSelectedSkillCasterId("");
    setSelectedHandCardId("");
    setSelectedOwnUnitId("");
    setSelectedEnemyUnitId("");
    const hint =
      mode === "own-unit"
        ? "Hero skill armed: pick your unit"
        : mode === "enemy-unit"
          ? "Hero skill armed: pick enemy unit"
          : mode === "enemy-any"
            ? "Hero skill armed: pick enemy unit or enemy hero"
            : "Hero skill armed: pick enemy hero";
    setActionStatus(hint);
  }

  function handleHeroAttackToggle() {
    if (!selectedMatch || !myPlayer) {
      pushToast("No battle selected", "error");
      return;
    }
    if (!isMyTurn) {
      pushToast("Wait for your turn", "error");
      return;
    }
    if (myPlayer.hero_attack_cooldown > 0) {
      pushToast(`Hero attack on cooldown (${myPlayer.hero_attack_cooldown})`, "error");
      return;
    }
    if (myPlayer.hero_attack_power <= 0) {
      pushToast("Hero attack is not available", "error");
      return;
    }
    const next = !heroAttackArmed;
    setHeroAttackArmed(next);
    setHeroSpellArmed(false);
    setDragAttack(null);
    setSelectedSkillCasterId("");
    setSelectedHandCardId("");
    setSelectedOwnUnitId("");
    setSelectedEnemyUnitId("");
    setActionStatus(next ? "Hero attack armed: pick enemy unit or hero" : "Hero attack canceled");
  }

  async function handleLeaveMatch() {
    await applyAction({ type: "leave_match" }, "You left the battle");
  }

  async function handleOwnUnitClick(unit: UnitState) {
    if (heroSpellArmed && canSelectUnitAsHeroSpellTarget("own")) {
      await applyAction(
        {
          type: "hero_spell",
          target_instance_id: unitInstanceId(unit),
        },
        "Hero ability activated",
      );
      return;
    }

    if (selectedSkillCaster) {
      if (!canSelectUnitAsSkillTarget(selectedSkillCaster, "own", unit)) {
        setActionStatus(`Invalid target for ${targetLabel(unitSkillTarget(selectedSkillCaster))}`);
        return;
      }
      await castSkill(unitInstanceId(selectedSkillCaster), unitInstanceId(unit));
      return;
    }

    if (selectedCard && cardKind(selectedCard) === "buff") {
      await applyAction(
        {
          type: "play_buff_card",
          card_instance_id: cardInstanceId(selectedCard),
          target_instance_id: unitInstanceId(unit),
        },
        "Buff card cast",
      );
      return;
    }

    setSelectedOwnUnitId(unitInstanceId(unit));
    setActionStatus(`Selected allied unit: ${resolveAssetLabel(unitTemplateId(unit))}`);
  }

  async function handleEnemyUnitClick(unit: UnitState) {
    if (heroAttackArmed) {
      await applyAction(
        {
          type: "hero_attack",
          target_instance_id: unitInstanceId(unit),
        },
        "Hero attacked enemy unit",
      );
      return;
    }

    if (heroSpellArmed && canSelectUnitAsHeroSpellTarget("enemy")) {
      await applyAction(
        {
          type: "hero_spell",
          target_instance_id: unitInstanceId(unit),
        },
        "Hero ability activated",
      );
      return;
    }

    if (selectedSkillCaster) {
      if (!canSelectUnitAsSkillTarget(selectedSkillCaster, "enemy", unit)) {
        setActionStatus(`Invalid target for ${targetLabel(unitSkillTarget(selectedSkillCaster))}`);
        return;
      }
      await castSkill(unitInstanceId(selectedSkillCaster), unitInstanceId(unit));
      return;
    }

    if (selectedCard) {
      setSelectedEnemyUnitId(unitInstanceId(unit));
      setActionStatus("Enemy unit marked");
      return;
    }

    if (selectedOwnUnitId) {
      await applyAction(
        {
          type: "card_attack",
          card_instance_id: selectedOwnUnitId,
          target_instance_id: unitInstanceId(unit),
        },
        "Unit attacked enemy unit",
      );
      return;
    }

    setSelectedEnemyUnitId(unitInstanceId(unit));
    setActionStatus(`Selected enemy unit: ${resolveAssetLabel(unitTemplateId(unit))}`);
  }

  async function handleEnemyHeroClick() {
    if (heroAttackArmed) {
      await applyAction(
        {
          type: "hero_attack",
          attack_hero: true,
        },
        "Hero attacked enemy hero",
      );
      return;
    }

    if (heroSpellArmed && canSelectEnemyHeroAsHeroSpellTarget()) {
      await applyAction(
        {
          type: "hero_spell",
          attack_hero: true,
        },
        "Hero ability activated",
      );
      return;
    }

    if (selectedSkillCaster) {
      if (!canCardSkillTargetEnemyHero(selectedSkillCaster)) {
        pushToast("This card skill cannot target hero", "error");
        return;
      }
      await castSkill(unitInstanceId(selectedSkillCaster), "", true);
      return;
    }

    if (selectedOwnUnitId) {
      await applyAction(
        {
          type: "card_attack",
          card_instance_id: selectedOwnUnitId,
          attack_hero: true,
        },
        "Unit attacked enemy hero",
      );
      return;
    }

    setSelectedEnemyUnitId("");
    setActionStatus("Enemy hero selected");
  }

  function startUnitDrag(unit: UnitState, sourceEl: HTMLElement | null, clientX: number, clientY: number) {
    if (selectedCard || selectedSkillCaster || heroSpellArmed || heroAttackArmed || !selectedMatch || !myPlayer) {
      return;
    }
    setSelectedOwnUnitId(unitInstanceId(unit));
    const sourcePoint = sourceEl ? getElementCenterInBoard(sourceEl) : getBoardRelativePoint(clientX, clientY);
    if (!sourcePoint) {
      return;
    }
    setDragAttack({
      sourceId: unitInstanceId(unit),
      sourceX: sourcePoint.x,
      sourceY: sourcePoint.y,
      currentX: sourcePoint.x,
      currentY: sourcePoint.y,
    });
    setActionStatus(`Attack vector: ${resolveAssetLabel(unitTemplateId(unit))}`);
  }

  function centerPointInBoard(selector: string): { x: number; y: number } | null {
    if (!battleBoardRef.current) {
      return null;
    }
    const node = battleBoardRef.current.querySelector<HTMLElement>(selector);
    if (!node) {
      return null;
    }
    const boardRect = battleBoardRef.current.getBoundingClientRect();
    const rect = node.getBoundingClientRect();
    return {
      x: rect.left + rect.width / 2 - boardRect.left,
      y: rect.top + rect.height / 2 - boardRect.top,
    };
  }

  function renderHeroGlyph(heroCode: string, imageKey: string | undefined, size: "small" | "large") {
    const tone = getAssetTone(heroCode);
    const label = resolveAssetLabel(imageKey || heroCode || "hero");
    return (
      <div className={`hero-glyph ${size} tone-${tone}`}>
        <AssetImage
          imageKey={imageKey || resolveHeroImageKey(heroCode)}
          alt={label}
          fallbackSrc={resolveHeroFallbackSrc()}
          className="hero-glyph-media"
        />
      </div>
    );
  }

  function renderUnitSlot(unit: UnitState | null, side: "own" | "enemy", slot: number) {
    if (!unit) {
      return <button className="slot empty" onClick={() => handlePlaySelectedCard(slot)}>+</button>;
    }

    const selectedByClicks = side === "own" ? selectedOwnUnitId === unitInstanceId(unit) : selectedEnemyUnitId === unitInstanceId(unit);
    const selectedBySkill = side === "own" && selectedSkillCasterId === unitInstanceId(unit);
    const selected = selectedByClicks || selectedBySkill;
    const skillTargetable = Boolean(
      (selectedSkillCaster && canSelectUnitAsSkillTarget(selectedSkillCaster, side, unit)) ||
        (heroSpellArmed && canSelectUnitAsHeroSpellTarget(side)),
    );
    const tone = getAssetTone(unitTemplateId(unit));
    const meta = cardCatalogEntry(unitTemplateId(unit));
    const fallbackSkill = skillFallbackByTemplate[unitTemplateId(unit)];
    const cooldownLeft = unitCooldown(unit);
    const templateCooldown = meta && "cooldown" in meta && typeof meta.cooldown === "number" ? meta.cooldown : 0;
    const baseCooldown = Math.max(unitBaseCooldown(unit), templateCooldown, cooldownLeft);
    const isOnCooldown = cooldownLeft > 0;
    const isTank = unitIsTank(unit);
    const ownerTurns = side === "own" ? (myPlayer?.turns ?? -1) : (enemyPlayer?.turns ?? -1);
    const isDeployed = ownerTurns >= 0 && unitSummonedInTurn(unit) === ownerTurns;
    const skillCode = unitSkillCode(unit) || meta?.skill_code || fallbackSkill?.code || "";
    const skillTrigger = unitSkillTrigger(unit) || meta?.skill_trigger || fallbackSkill?.trigger || "";
    const skillTarget = unitSkillTarget(unit) || meta?.skill_target || fallbackSkill?.target || "";
    const skillName = unitSkillName(unit) || meta?.skill_name || fallbackSkill?.name || skillCode;
    const hasSkill = skillCode !== "";
    const hasActiveSkill = hasSkill && skillTrigger === SKILL_TRIGGER_ACTIVE;
    const skillLeft = unitSkillCooldownLeft(unit);
    const skillBase = unitSkillCooldown(unit) || meta?.skill_cooldown || fallbackSkill?.cooldown || 0;
    const skillDisabled = !hasActiveSkill || skillLeft > 0 || !isMyTurn || Boolean(selectedCard);
    const isAttackSource =
      boardAttackAnimation?.sourceKind === "unit" &&
      boardAttackAnimation.sourceInstanceId === unitInstanceId(unit);
    const isHitTarget = boardAttackAnimation?.targetIds.includes(unitInstanceId(unit)) ?? false;
    const slotStyle = {
      transform:
        `${selected ? "translateY(-2px) " : ""}` +
        `${isAttackSource ? `translate(${boardAttackAnimation?.dx ?? 0}px, ${boardAttackAnimation?.dy ?? 0}px)` : ""}`,
    } as CSSProperties;

    return (
      <button
        className={`slot tone-${tone} ${selected ? "selected" : ""} ${skillTargetable ? "skill-targetable" : ""} ${selectedBySkill ? "skill-caster" : ""} ${isAttackSource ? "attack-source" : ""} ${isHitTarget ? "hit-target" : ""}`}
        data-unit-id={unitInstanceId(unit)}
        data-slot-side={side}
        data-attack-target={side === "enemy" ? "enemy-unit" : undefined}
        style={slotStyle}
        onClick={() =>
          void (side === "own" ? handleOwnUnitClick(unit) : handleEnemyUnitClick(unit))
        }
        onPointerDown={
          side === "own"
            ? (event) => {
                event.preventDefault();
                startUnitDrag(unit, event.currentTarget as HTMLElement, event.clientX, event.clientY);
              }
            : undefined
        }
      >
        <AssetImage
          imageKey={cardImageKeyForTemplate(unitTemplateId(unit))}
          alt={resolveAssetLabel(unitTemplateId(unit))}
          fallbackSrc={resolveCardFallbackSrc()}
          className="slot-media"
        />
        {side === "own" && hasSkill && (
          <span
            className={`slot-skill-btn ${selectedBySkill ? "armed" : ""} ${skillDisabled ? "disabled" : ""}`}
            onClick={(event) => {
              event.preventDefault();
              event.stopPropagation();
              if (skillDisabled) {
                if (!hasActiveSkill) {
                  pushToast(`Passive skill: ${skillName || skillCode}`);
                }
                return;
              }
              void runTask(() => startSkillCast(unit));
            }}
            title={`${skillName || "Skill"}${skillTarget ? ` • ${targetLabel(skillTarget)}` : ""}`}
            role="button"
            tabIndex={0}
            onKeyDown={(event) => {
              if (event.key !== "Enter" && event.key !== " ") {
                return;
              }
              event.preventDefault();
              if (skillDisabled) {
                return;
              }
              void runTask(() => startSkillCast(unit));
            }}
          >
            {skillName || "Skill"}
            {skillBase > 0 && <span className="slot-skill-cd">{` ${skillLeft}/${skillBase}`}</span>}
          </span>
        )}
        {isTank && (
          <div className="slot-topline">
            <span className="card-chip tank">TANK</span>
          </div>
        )}
        {isDeployed && <div className="slot-deployed-label">-Deployed-</div>}
        {isOnCooldown && <div className="slot-cooldown-label">-CoolDown-</div>}
        <div className="slot-stats">
          <span className="slot-stat hp">{unitHP(unit)}</span>
          <span className="slot-stat atk">{unitAttack(unit)}</span>
          <span className="slot-stat cd">{cooldownLeft}/{baseCooldown}</span>
        </div>
      </button>
    );
  }

  function handleBattleBoardEmptyClick(event: MouseEvent<HTMLElement>) {
    const target = event.target as HTMLElement;
    if (
      target.closest(
        ".slot, .hand-card, .hero-orb-button, .hero-skill-mini, .hero-attack-mini, .end-turn-floating, .ghost-button, .slot-skill-btn, .battle-deck-anchor, .grave-trigger, .grave-panel, .deck-trigger",
      )
    ) {
      return;
    }
    if (selectedHandCardId || selectedOwnUnitId || selectedEnemyUnitId || selectedSkillCasterId || heroSpellArmed || heroAttackArmed) {
      clearSelections();
      setActionStatus("Selection cleared");
    }
  }

  function formatElapsedTimer(totalSeconds: number): string {
    const safe = Math.max(0, totalSeconds);
    const minutes = Math.floor(safe / 60);
    const seconds = safe % 60;
    return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
  }

  function penaltySecondsLeft(isoTime?: string): number {
    if (!isoTime) {
      return 0;
    }
    const untilMs = Date.parse(isoTime);
    if (Number.isNaN(untilMs)) {
      return 0;
    }
    return Math.max(0, Math.ceil((untilMs - Date.now()) / 1000));
  }

  function polarToCartesian(cx: number, cy: number, r: number, angleDeg: number) {
    const rad = (angleDeg * Math.PI) / 180;
    return {
      x: cx + r * Math.cos(rad),
      y: cy + r * Math.sin(rad),
    };
  }

  function describeArc(cx: number, cy: number, r: number, startAngle: number, endAngle: number) {
    const start = polarToCartesian(cx, cy, r, startAngle);
    const end = polarToCartesian(cx, cy, r, endAngle);
    let delta = endAngle - startAngle;
    if (delta < 0) {
      delta += 360;
    }
    const largeArcFlag = delta > 180 ? 1 : 0;
    return `M ${start.x.toFixed(2)} ${start.y.toFixed(2)} A ${r} ${r} 0 ${largeArcFlag} 1 ${end.x.toFixed(2)} ${end.y.toFixed(2)}`;
  }

  function renderHeroHud(player: MatchPlayer, hpPeak: number, hpOnBottom = false) {
    const hpMax = Math.max(1, hpPeak || player.hero_hp || 1);
    const hpRatio = Math.max(0, Math.min(1, player.hero_hp / hpMax));
    const ringRadius = 58;
    const topStart = 184;
    const topEnd = 356;
    const bottomStart = 4;
    const bottomEnd = 176;
    const hpStart = hpOnBottom ? bottomStart : topStart;
    const hpEndMax = hpOnBottom ? bottomEnd : topEnd;
    const hpEnd = hpStart + (hpEndMax - hpStart) * hpRatio;
    const manaStart = hpOnBottom ? topStart : bottomStart;
    const manaEnd = hpOnBottom ? topEnd : bottomEnd;
    const manaCells = Math.max(0, Math.min(10, player.mana));
    const manaDividerAngles =
      manaCells > 1
        ? Array.from(
            { length: manaCells - 1 },
            (_, index) => manaStart + ((index + 1) * (manaEnd - manaStart)) / manaCells,
          )
        : [];
    return (
      <div className={`hero-orb ${selectedMatch?.active_player === player.player_id ? "active-turn" : ""}`}>
        <svg className="hero-orb-rings" viewBox="0 0 120 120" aria-hidden="true">
          <path className="hero-hp-track" d={describeArc(60, 60, ringRadius, hpStart, hpEndMax)} />
          <path className="hero-hp-value" d={describeArc(60, 60, ringRadius, hpStart, hpEnd)} />
          <path className="hero-mana-track" d={describeArc(60, 60, ringRadius, manaStart, manaEnd)} />
          {manaCells > 0 && <path className="hero-mana-value" d={describeArc(60, 60, ringRadius, manaStart, manaEnd)} />}
          {[0, 180].map((angle) => {
            const outer = polarToCartesian(60, 60, 62, angle);
            const inner = polarToCartesian(60, 60, 54, angle);
            return (
              <line
                key={`ring-sep-${angle}`}
                className="hero-ring-joint-separator"
                x1={inner.x}
                y1={inner.y}
                x2={outer.x}
                y2={outer.y}
              />
            );
          })}
          {manaDividerAngles.map((angle) => {
            const outer = polarToCartesian(60, 60, 62, angle);
            const inner = polarToCartesian(60, 60, 54, angle);
            return (
              <line
                key={`mana-divider-${angle}`}
                className="hero-mana-divider"
                x1={inner.x}
                y1={inner.y}
                x2={outer.x}
                y2={outer.y}
              />
            );
          })}
        </svg>
        <div className="hero-orb-avatar">
          {renderHeroGlyph(
            player.hero_code,
            `heroes/${player.hero_code}/image`,
            "large",
          )}
        </div>
      </div>
    );
  }

  const selectedHeroImageKey =
    heroes.find((hero) => hero.hero_code === me?.selected_hero_code)?.image_key ||
    resolveHeroImageKey(me?.selected_hero_code || "unassigned");
  const deckTotal = totalDeck(deckEntries);
  const deckReady = deckTotal === 20;
  const deckGroups = useMemo(() => {
    return deckEntries.map((entry) => {
      const meta = cardCatalog.get(entry.template_id);
      const imageKey = meta?.image_key
        ? meta.image_key
        : entry.kind === "buff"
          ? resolveBuffCardImageKey(entry.template_id)
          : resolveBattleCardImageKey(entry.template_id);
      return {
        key: `${entry.kind}:${entry.template_id}`,
        kind: entry.kind,
        templateId: entry.template_id,
        name: meta?.name || resolveAssetLabel(entry.template_id),
        count: entry.count,
        imageKey,
        mana: meta?.mana_cost ?? 0,
      };
    });
  }, [cardCatalog, deckEntries]);
  const inspectedDeckGroup = deckGroups.find((group) => group.key === deckInspectorKey) ?? null;
  const inspectedDeckMeta = inspectedDeckGroup ? cardCatalog.get(inspectedDeckGroup.templateId) : undefined;
  const deckPreviewGroups = deckGroups.slice(0, 4);
  const hiddenDeckGroups = Math.max(0, deckGroups.length - deckPreviewGroups.length);
  const deckCountMap = useMemo(() => {
    const next = new Map<string, number>();
    for (const entry of deckEntries) {
      next.set(`${entry.kind}:${entry.template_id}`, entry.count);
    }
    return next;
  }, [deckEntries]);
  const catalogCards = useMemo(() => {
    if (catalogKind === "battle") {
      const battle = [...(cards?.battle ?? [])];
      battle.sort((a, b) => {
        switch (catalogSort) {
          case "attack":
            return b.attack - a.attack || a.mana_cost - b.mana_cost;
          case "hp":
            return b.health_points - a.health_points || a.mana_cost - b.mana_cost;
          case "tank":
            return Number(b.is_tank) - Number(a.is_tank) || a.mana_cost - b.mana_cost;
          case "mana":
          default:
            return a.mana_cost - b.mana_cost || b.attack - a.attack;
        }
      });
      return battle;
    }
    const buff = [...(cards?.buff ?? [])];
    buff.sort((a, b) => a.mana_cost - b.mana_cost || a.name.localeCompare(b.name));
    return buff;
  }, [cards, catalogKind, catalogSort]);
  const catalogPages = Math.max(1, Math.ceil(catalogCards.length / 6));
  useEffect(() => {
    setCatalogPage((prev) => Math.min(prev, catalogPages - 1));
  }, [catalogPages]);
  const catalogPageItems = useMemo(() => {
    const from = catalogPage * 6;
    return catalogCards.slice(from, from + 6);
  }, [catalogCards, catalogPage]);
  const matchmakingTimerLabel = useMemo(() => {
    if (queueStatus.state === "searching") {
      return formatElapsedTimer(queueStatus.search_duration_sec ?? 0);
    }
    if (queueStatus.state === "penalty") {
      return formatElapsedTimer(penaltySecondsLeft(queueStatus.penalty_until));
    }
    return "00:00";
  }, [queueStatus]);
  const pendingAcceptSecondsLeft = useMemo(() => {
    if (queueStatus.state !== "pending_match") {
      return 0;
    }
    return penaltySecondsLeft(queueStatus.accept_deadline_at);
  }, [queueStatus]);
  const pendingAcceptTimerLabel = useMemo(
    () => formatElapsedTimer(pendingAcceptSecondsLeft),
    [pendingAcceptSecondsLeft],
  );
  const pendingAcceptProgress = useMemo(() => {
    if (queueStatus.state !== "pending_match") {
      return 0;
    }
    const deadline = queueStatus.accept_deadline_at ? Date.parse(queueStatus.accept_deadline_at) : Number.NaN;
    if (!Number.isFinite(deadline)) {
      return 0;
    }
    const total = 10;
    return Math.max(0, Math.min(1, pendingAcceptSecondsLeft / total));
  }, [pendingAcceptSecondsLeft, queueStatus]);
  const canOpenQueuePanel = queueStatus.state === "idle" || queueStatus.state === "penalty";

  return (
    <div className={`war-shell ${activeBattle ? "battle-mode" : ""}`}>
      <main className="view-frame">
        {!activeBattle && tab === "home" && (
          <section className="screen-grid home-grid">
            {(queueStatus.state === "searching" || queueStatus.state === "penalty") && (
              <div className={`queue-status-banner ${queueStatus.state === "penalty" ? "danger" : ""}`}>
                <div className="queue-status-copy">
                  <span className="queue-status-kicker">
                    {queueStatus.state === "searching" ? "ПОИСК МАТЧА" : "ПОИСК НЕДОСТУПЕН"}
                  </span>
                  <strong>{matchmakingTimerLabel}</strong>
                </div>
                <button
                  className={`queue-status-action ${queueStatus.state === "penalty" ? "disabled" : ""}`}
                  onClick={() => void runTask(leaveMatchmakingQueue)}
                  disabled={queueStatus.state === "penalty"}
                >
                  Отмена
                </button>
              </div>
            )}
            {queueStatus.state === "pending_match" && (
              <div className="match-found-overlay">
                <section className="match-found-panel">
                  <span className="match-found-kicker">Matchmaking</span>
                  <strong className="match-found-title">МАТЧ НАЙДЕН</strong>
                  <span className="match-found-subtitle">
                    {queueStatus.accepted_by_me || queueStatus.accepted_by_opponent
                      ? "Ожидание остальных"
                      : "Подтвердите готовность к бою"}
                  </span>

                  <div className="match-found-players">
                    <div className={`match-found-player ${queueStatus.accepted_by_me ? "accepted" : ""}`}>
                      <span className="match-found-player-label">ВЫ</span>
                      <span className="match-found-player-status">
                        {queueStatus.accepted_by_me ? "Принято" : "Ожидание"}
                      </span>
                    </div>
                    <div className={`match-found-player ${queueStatus.accepted_by_opponent ? "accepted" : ""}`}>
                      <span className="match-found-player-label">ПРОТИВНИК</span>
                      <span className="match-found-player-status">
                        {queueStatus.accepted_by_opponent ? "Принято" : "Ожидание"}
                      </span>
                    </div>
                  </div>

                  <div className="match-found-deadline">
                    <div className="match-found-deadline-copy">
                      <span className="match-found-deadline-label">ДЕДЛАЙН</span>
                      <strong>{pendingAcceptTimerLabel}</strong>
                    </div>
                    <div className={`match-found-timer-line ${pendingAcceptSecondsLeft <= 3 ? "danger" : ""}`}>
                      <span
                        className="match-found-timer-fill left"
                        style={{ width: `${pendingAcceptProgress * 50}%` }}
                      />
                      <span
                        className="match-found-timer-fill right"
                        style={{ width: `${pendingAcceptProgress * 50}%` }}
                      />
                    </div>
                  </div>

                  <div className="match-found-actions">
                    <button
                      className={`match-found-action primary ${queueStatus.accepted_by_me ? "accepted" : ""}`}
                      onClick={() => void runTask(acceptMatchmakingReady)}
                      disabled={Boolean(queueStatus.accepted_by_me)}
                    >
                      {queueStatus.accepted_by_me ? "ОЖИДАНИЕ" : "В БОЙ"}
                    </button>
                    <button className="match-found-action danger" onClick={() => void runTask(declineMatchmakingReady)}>
                      ОТКАЗ
                    </button>
                  </div>
                </section>
              </div>
            )}
            <div className="panel command-panel">
              <div
                className="hero-banner hero-banner-top"
                style={
                  {
                    "--hero-panel-image": `url(${resolveImageSrc(selectedHeroImageKey)})`,
                  } as CSSProperties
                }
              >
                <button
                  className="ghost-button miniapp-fullscreen-button"
                  onClick={() => void runTask(handleToggleMiniAppFullscreen)}
                  title={miniAppFullscreen ? "Exit fullscreen" : "Open fullscreen"}
                  aria-label={miniAppFullscreen ? "Exit fullscreen" : "Open fullscreen"}
                >
                  {miniAppFullscreen ? "WIN" : "FULL"}
                </button>
                <button
                  className={`avatar-trigger hero-profile-mini ${loading ? "busy" : ""}`}
                  onClick={() => setShowProfile(true)}
                >
                  <span className="avatar-core">
                    {(me?.first_name?.[0] || me?.username?.[0] || "?").toUpperCase()}
                  </span>
                </button>
                <button className="hero-portrait-stage hero-select-trigger" onClick={() => setHeroPickerOpen(true)}>
                  {renderHeroGlyph(
                    me?.selected_hero_code || "unassigned",
                    selectedHeroImageKey,
                    "large",
                  )}
                  <div className="hero-banner-copy overlay">
                    <h3>{me?.first_name || me?.username || "No profile loaded"}</h3>
                    <p className="hero-banner-role">{me?.selected_hero_name || "No Hero Assigned"}</p>
                    <p className="muted">Rating {me?.rating ?? "-"}</p>
                  </div>
                </button>
              </div>
              {heroPickerOpen && (
                <div
                  className="hero-picker"
                  onClick={() => {
                    setHeroPickerOpen(false);
                    setHeldHero(null);
                  }}
                >
                  <div className="hero-picker-body" onClick={(event) => event.stopPropagation()}>
                    <button
                      className="hero-picker-close"
                      onClick={() => {
                        setHeroPickerOpen(false);
                        setHeldHero(null);
                      }}
                    >
                      X
                    </button>
                    {heldHero && (
                      <div className="hero-preview">
                        <div className="hero-preview-media">
                          <AssetImage
                            imageKey={heldHero.image_key || resolveHeroImageKey(heldHero.hero_code)}
                            alt={heldHero.name}
                            fallbackSrc={resolveHeroFallbackSrc()}
                            className="hero-preview-image"
                          />
                        </div>
                        <div className="hero-preview-info">
                          <strong>{heldHero.name}</strong>
                          <span>HP {heldHero.health_points}</span>
                          <span>ATK {heldHero.attack_power}</span>
                          <span>CD {heldHero.attack_cooldown}</span>
                          <span>{heldHero.description}</span>
                        </div>
                      </div>
                    )}
                    <div className="hero-picker-head">
                      <strong>Выбери героя</strong>
                    </div>
                    <div className="hero-picker-list">
                      {heroes.map((hero) => {
                        const selected = hero.hero_code === me?.selected_hero_code;
                        return (
                          <article
                            key={hero.hero_code}
                            className={`hero-tile ${selected ? "selected" : ""}`}
                            onClick={() => setHeldHero(hero)}
                          >
                            {renderHeroGlyph(hero.hero_code, hero.image_key, "small")}
                            <span className="hero-tile-name">{hero.name}</span>
                            <button
                              className="hero-tile-pick"
                              onClick={(event) => {
                                event.stopPropagation();
                                void runTask(async () => {
                                  await selectHero(hero.hero_code);
                                  setHeroPickerOpen(false);
                                  setHeldHero(null);
                                });
                              }}
                            >
                              Выбрать
                            </button>
                          </article>
                        );
                      })}
                    </div>
                  </div>
                </div>
              )}
              {queuePanelOpen && (
                <div
                  className="queue-mode-overlay"
                  onClick={() => {
                    setQueuePanelOpen(false);
                    setSelectedQueueMode(null);
                  }}
                >
                  <aside className="queue-mode-drawer" onClick={(event) => event.stopPropagation()}>
                    <button
                      className="queue-mode-close"
                      onClick={() => {
                        setQueuePanelOpen(false);
                        setSelectedQueueMode(null);
                      }}
                    >
                      X
                    </button>
                    <div className="queue-mode-art">
                      <div className="queue-mode-art-placeholder">
                        <span>ART</span>
                      </div>
                    </div>
                    <div className="queue-mode-copy">
                      <span className="panel-kicker">Matchmaking</span>
                      <strong>Choose a queue</strong>
                    </div>
                    <button
                      className={`queue-mode-option ${selectedQueueMode === "ranked" ? "selected" : ""}`}
                      onClick={() => setSelectedQueueMode((prev) => (prev === "ranked" ? null : "ranked"))}
                    >
                      <span className={`queue-mode-check ${selectedQueueMode === "ranked" ? "active" : ""}`} />
                      <span className="queue-mode-option-copy">
                        <strong>Рейтинговая игра</strong>
                        <span>Бой за рейтинг и честный матчмейкинг</span>
                      </span>
                    </button>
                    <button
                      className="queue-mode-confirm"
                      disabled={!selectedQueueMode}
                      onClick={() => void runTask(joinMatchmakingQueue)}
                    >
                      НАЙТИ
                    </button>
                  </aside>
                </div>
              )}
              {!me && <button onClick={() => void runTask(retryTelegramAuth)}>Retry Auth</button>}
              <button
                className={`matchmaking-launch ${queueStatus.state === "penalty" ? "danger" : ""}`}
                onClick={() => {
                  if (!canOpenQueuePanel) {
                    return;
                  }
                  setQueuePanelOpen(true);
                }}
                disabled={!canOpenQueuePanel}
              >
                {queueStatus.state === "penalty"
                  ? "ПОИСК НЕДОСТУПЕН"
                  : queueStatus.state === "pending_match"
                    ? "МАТЧ НАЙДЕН"
                    : queueStatus.state === "searching"
                      ? "ПОИСК ИДЕТ"
                      : "НАЙТИ МАТЧ"}
              </button>
              <button className="open-inventory" onClick={() => setTab("inventory")}>
                Inventory
              </button>
              <div className="home-placeholder-card">
                <span className="panel-kicker">Store</span>
                <strong>Soon</strong>
                <p>Здесь появится магазин, паки и дополнительные предложения.</p>
              </div>
            </div>

          </section>
        )}

        {!activeBattle && tab === "inventory" && (
          <section className="screen-grid">
            <div className="panel inventory-panel">
              <div className="section-head inventory-deck-head">
                <button className="ghost-button inventory-inline-back" onClick={() => setTab("home")}>
                  {"<"} Back
                </button>
                <h2>Deck Doctrine</h2>
              </div>
              <div className="deck-summary">
                <span>Total cards</span>
                <strong>{deckTotal}</strong>
              </div>
              {!deckReady && <p className="deck-warning">Дека не собрана (нужно 20 карт)</p>}
              <button
                className={`deck-preview-shell ${deckGroups.length === 0 ? "empty" : ""}`}
                onClick={() => setDeckOverviewOpen(true)}
              >
                <div className="deck-grid deck-grid-preview">
                {deckPreviewGroups.map((group) => (
                  <button
                    key={group.key}
                    className="deck-slot filled interactive"
                    onClick={(event) => {
                      event.stopPropagation();
                      setDeckOverviewOpen(true);
                    }}
                  >
                    <GameCard
                      mode="deck"
                      data={{
                        kind: group.kind,
                        name: group.name,
                        description: cardCatalog.get(group.templateId)?.description || "",
                        imageKey: group.imageKey,
                        race:
                          group.kind === "battle"
                            ? cardRaceLabel(cardCatalog.get(group.templateId)?.card_type as string)
                            : "ЭФФЕКТ",
                        mana: group.mana,
                        attack:
                          group.kind === "battle"
                            ? (cardCatalog.get(group.templateId)?.attack as number | undefined)
                            : undefined,
                        hp:
                          group.kind === "battle"
                            ? (cardCatalog.get(group.templateId)?.health_points as number | undefined)
                            : undefined,
                        cooldown:
                          group.kind === "battle"
                            ? (cardCatalog.get(group.templateId)?.cooldown as number | undefined)
                            : undefined,
                        skillCooldown:
                          group.kind === "battle"
                            ? (cardCatalog.get(group.templateId)?.skill_cooldown as number | undefined)
                            : undefined,
                        buffType:
                          group.kind === "buff"
                            ? (cardCatalog.get(group.templateId)?.buff_type as string | undefined)
                            : undefined,
                        buffValue:
                          group.kind === "buff"
                            ? (cardCatalog.get(group.templateId)?.buff_value as number | undefined)
                            : undefined,
                        duration:
                          group.kind === "buff"
                            ? (cardCatalog.get(group.templateId)?.duration as number | undefined)
                            : undefined,
                      }}
                    />
                    <span className="deck-slot-count">x{group.count}</span>
                  </button>
                ))}
                {hiddenDeckGroups > 0 && (
                  <article className="deck-slot deck-slot-more">
                    <span>+{hiddenDeckGroups}</span>
                  </article>
                )}
                {deckGroups.length === 0 && (
                  <article className="deck-slot empty deck-slot-empty-wide">
                    <span>Deck is empty</span>
                  </article>
                )}
                </div>
              </button>
              {deckOverviewOpen && (
                <div className="deck-overview-overlay" onClick={() => setDeckOverviewOpen(false)}>
                  <div className="deck-overview-window" onClick={(event) => event.stopPropagation()}>
                    <button className="deck-overview-close" onClick={() => setDeckOverviewOpen(false)}>
                      X
                    </button>
                    <div className="deck-overview-head">
                      <strong>Deck Doctrine</strong>
                      <span>{deckTotal} cards</span>
                    </div>
                    <div className="deck-grid deck-grid-overview">
                      {deckGroups.map((group) => (
                        <button
                          key={`${group.key}:overview`}
                          className="deck-slot filled interactive"
                          onClick={() => setDeckInspectorKey(group.key)}
                        >
                          <GameCard
                            mode="deck"
                            data={{
                              kind: group.kind,
                              name: group.name,
                              description: cardCatalog.get(group.templateId)?.description || "",
                              imageKey: group.imageKey,
                              race:
                                group.kind === "battle"
                                  ? cardRaceLabel(cardCatalog.get(group.templateId)?.card_type as string)
                                  : "ЭФФЕКТ",
                              mana: group.mana,
                              attack:
                                group.kind === "battle"
                                  ? (cardCatalog.get(group.templateId)?.attack as number | undefined)
                                  : undefined,
                              hp:
                                group.kind === "battle"
                                  ? (cardCatalog.get(group.templateId)?.health_points as number | undefined)
                                  : undefined,
                              cooldown:
                                group.kind === "battle"
                                  ? (cardCatalog.get(group.templateId)?.cooldown as number | undefined)
                                  : undefined,
                              skillCooldown:
                                group.kind === "battle"
                                  ? (cardCatalog.get(group.templateId)?.skill_cooldown as number | undefined)
                                  : undefined,
                              buffType:
                                group.kind === "buff"
                                  ? (cardCatalog.get(group.templateId)?.buff_type as string | undefined)
                                  : undefined,
                              buffValue:
                                group.kind === "buff"
                                  ? (cardCatalog.get(group.templateId)?.buff_value as number | undefined)
                                  : undefined,
                              duration:
                                group.kind === "buff"
                                  ? (cardCatalog.get(group.templateId)?.duration as number | undefined)
                                  : undefined,
                            }}
                          />
                          <span className="deck-slot-count">x{group.count}</span>
                        </button>
                      ))}
                    </div>
                  </div>
                </div>
              )}
              {inspectedDeckGroup && (
                <div className="deck-fan-overlay" onClick={() => setDeckInspectorKey(null)}>
                  <div className="deck-fan-window" onClick={(event) => event.stopPropagation()}>
                    <button className="deck-fan-close" onClick={() => setDeckInspectorKey(null)}>
                      X
                    </button>
                    <div className="deck-fan-head">
                      <strong>{inspectedDeckGroup.name}</strong>
                      <span>x{inspectedDeckGroup.count}</span>
                    </div>
                    <div className="deck-fan-row">
                      <div
                        className="deck-fan-dense"
                        style={{ "--fan-count": `${inspectedDeckGroup.count}` } as CSSProperties}
                      >
                        {Array.from({ length: inspectedDeckGroup.count }).map((_, index, array) => (
                          <article
                            key={`${inspectedDeckGroup.key}:fan:${index}`}
                            className="deck-fan-card"
                            style={
                              {
                                "--fan-offset": `${index - (array.length - 1) / 2}`,
                              } as CSSProperties
                            }
                          >
                            <GameCard
                              mode="deck"
                              data={{
                                kind: inspectedDeckGroup.kind,
                                name: inspectedDeckGroup.name,
                                description: inspectedDeckMeta?.description || "",
                                imageKey: inspectedDeckGroup.imageKey,
                                race:
                                  inspectedDeckGroup.kind === "battle"
                                    ? cardRaceLabel(inspectedDeckMeta?.card_type as string)
                                    : "ЭФФЕКТ",
                                mana: inspectedDeckMeta?.mana_cost ?? 0,
                                attack:
                                  inspectedDeckGroup.kind === "battle"
                                    ? inspectedDeckMeta?.attack
                                    : undefined,
                                hp:
                                  inspectedDeckGroup.kind === "battle"
                                    ? inspectedDeckMeta?.health_points
                                    : undefined,
                                cooldown:
                                  inspectedDeckGroup.kind === "battle"
                                    ? inspectedDeckMeta?.cooldown
                                    : undefined,
                                skillCooldown:
                                  inspectedDeckGroup.kind === "battle"
                                    ? inspectedDeckMeta?.skill_cooldown
                                    : undefined,
                                buffType:
                                  inspectedDeckGroup.kind === "buff"
                                    ? inspectedDeckMeta?.buff_type
                                    : undefined,
                                buffValue:
                                  inspectedDeckGroup.kind === "buff"
                                    ? inspectedDeckMeta?.buff_value
                                    : undefined,
                                duration:
                                  inspectedDeckGroup.kind === "buff"
                                    ? inspectedDeckMeta?.duration
                                    : undefined,
                              }}
                            />
                            <button
                              className="deck-fan-remove"
                              onClick={() => void runTask(() => removeCardFromDeck(inspectedDeckGroup.kind, inspectedDeckGroup.templateId))}
                            >
                              X
                            </button>
                          </article>
                        ))}
                      </div>
                    </div>
                    <div className="deck-fan-info">
                      <span>DECK COPIES {inspectedDeckGroup.count}</span>
                      <span>MANA {inspectedDeckMeta?.mana_cost ?? 0}</span>
                      {inspectedDeckGroup.kind === "battle" ? (
                        <span>
                          HP {inspectedDeckMeta?.health_points ?? 0} | ATK {inspectedDeckMeta?.attack ?? 0} | CD {inspectedDeckMeta?.cooldown ?? 0} | MAX {inspectedDeckMeta?.max_copies ?? 0}
                        </span>
                      ) : (
                        <span>
                          {inspectedDeckMeta?.buff_type || "Buff"} {inspectedDeckMeta?.buff_value ?? 0} | DUR {inspectedDeckMeta?.duration ?? 0} | MAX {inspectedDeckMeta?.max_copies ?? 0}
                        </span>
                      )}
                      <span>{inspectedDeckMeta?.description || "-"}</span>
                    </div>
                  </div>
                </div>
              )}
            </div>

            <div className="panel inventory-panel catalog-panel">
              <div className="catalog-toolbar">
                <div className="catalog-kind-switch">
                  <button
                    className={catalogKind === "battle" ? "nav-pill active" : "nav-pill"}
                    onClick={() => {
                      setCatalogKind("battle");
                      setCatalogPage(0);
                    }}
                  >
                    Battle Cards
                  </button>
                  <button
                    className={catalogKind === "buff" ? "nav-pill active" : "nav-pill"}
                    onClick={() => {
                      setCatalogKind("buff");
                      setCatalogPage(0);
                    }}
                  >
                    Buff Cards
                  </button>
                </div>
                <label className="catalog-sort">
                  <span>Sort</span>
                  <select
                    value={catalogSort}
                    onChange={(event) => setCatalogSort(event.target.value as CatalogSort)}
                  >
                    <option value="mana">Mana</option>
                    <option value="attack">Attack</option>
                    <option value="hp">HP</option>
                    <option value="tank">Tank / Non-tank</option>
                  </select>
                </label>
              </div>
              <div className="catalog-grid">
                {catalogPageItems.map((card) => {
                  const imageKey =
                    card.kind === "battle"
                      ? card.image_key || resolveBattleCardImageKey(card.template_id)
                      : card.image_key || resolveBuffCardImageKey(card.template_id);
                  const templateKey = `${card.kind}:${card.template_id}`;
                  const deckCount = deckCountMap.get(templateKey) ?? 0;
                  const addLimit = Math.min(card.max_copies, card.copies);
                  const exhausted = deckCount >= addLimit;
                  return (
                    <article
                      key={templateKey}
                      className={`asset-card tone-${getAssetTone(card.asset_base_key)} clickable ${exhausted ? "exhausted" : ""}`}
                      onClick={() =>
                        setCardPreview(
                          card.kind === "battle"
                            ? {
                                kind: "battle",
                                name: card.name,
                                description: card.description,
                                imageKey,
                                race: cardRaceLabel(card.card_type),
                                mana: card.mana_cost,
                                hp: card.health_points,
                                attack: card.attack,
                                cooldown: card.cooldown,
                                skillCooldown: card.skill_cooldown,
                              }
                            : {
                                kind: "buff",
                                name: card.name,
                                description: card.description,
                                imageKey,
                                mana: card.mana_cost,
                                buffType: card.buff_type,
                                buffValue: card.buff_value,
                                duration: card.duration,
                              },
                        )
                      }
                    >
                      <div className="asset-frame">
                        <GameCard
                          mode="catalog"
                          data={{
                            kind: card.kind,
                            name: card.name,
                            description: card.description,
                            imageKey,
                            race: card.kind === "battle" ? cardRaceLabel(card.card_type) : "ЭФФЕКТ",
                            mana: card.mana_cost,
                            attack: card.kind === "battle" ? card.attack : undefined,
                            hp: card.kind === "battle" ? card.health_points : undefined,
                            cooldown: card.kind === "battle" ? card.cooldown : undefined,
                            skillCooldown: card.kind === "battle" ? card.skill_cooldown : undefined,
                            buffType: card.kind === "buff" ? card.buff_type : undefined,
                            buffValue: card.kind === "buff" ? card.buff_value : undefined,
                            duration: card.kind === "buff" ? card.duration : undefined,
                          }}
                        />
                        <button
                          className="asset-add"
                          disabled={exhausted}
                          onClick={(event) => {
                            event.stopPropagation();
                            if (exhausted) {
                              return;
                            }
                            void runTask(() => addCardToDeck(card.kind, card.template_id));
                          }}
                        >
                          +
                        </button>
                      </div>
                    </article>
                  );
                })}
              </div>
              <div className="catalog-pager">
                <button
                  className="ghost-button"
                  onClick={() => setCatalogPage((prev) => Math.max(0, prev - 1))}
                  disabled={catalogPage === 0}
                >
                  {"<"}
                </button>
                <span>
                  {catalogPage + 1} / {catalogPages}
                </span>
                <button
                  className="ghost-button"
                  onClick={() => setCatalogPage((prev) => Math.min(catalogPages - 1, prev + 1))}
                  disabled={catalogPage >= catalogPages - 1}
                >
                  {">"}
                </button>
              </div>
            </div>
            <div className="panel inventory-panel inventory-placeholder-card">
              <span className="panel-kicker">Store</span>
              <strong>Soon</strong>
              <p>Скоро здесь появится магазин, новые паки карт и дополнительные предложения.</p>
            </div>
          </section>
        )}

        {activeBattle && (
          <section className="battle-screen">
            <section
              className="battle-board panel"
              ref={battleBoardRef}
              onClick={handleBattleBoardEmptyClick}
            >
              <div
                className="battle-board-background"
                style={{ backgroundImage: `url(${resolveBoardBackgroundSrc()})` }}
              />
              {!selectedMatch || !myPlayer || !enemyPlayer ? (
                <div className="empty-battle">
                  <h2>No active battle selected</h2>
                  <p className="muted">Create a match from Start Game and enter the theatre.</p>
                </div>
              ) : (
                <>
                  {activeEvent && (
                    <aside className="battle-cinematic" aria-live="polite">
                      <div className="battle-cinematic-media">
                        <AssetImage
                          imageKey={eventSourceImageKey(activeEvent)}
                          alt={eventSourceLabel(activeEvent)}
                          fallbackSrc={resolveCardFallbackSrc()}
                          className="battle-cinematic-image"
                        />
                      </div>
                      <div className="battle-cinematic-copy">
                        <span className="battle-cinematic-kicker">{eventTitle(activeEvent)}</span>
                        <strong>{eventSourceLabel(activeEvent)}</strong>
                        <span>{`${activeEvent.targets?.length ?? 0} target${(activeEvent.targets?.length ?? 0) === 1 ? "" : "s"}`}</span>
                      </div>
                    </aside>
                  )}
                  <div className="battle-scene">
                    <div className="battle-top-utility">
                      <div className="battle-utility-stack">
                        <button className="ghost-button leave-inline in-board" onClick={() => void runTask(handleLeaveMatch)}>
                          Leave Match
                        </button>
                        <button className="ghost-button expand-inline in-board" onClick={() => void runTask(handleToggleMiniAppFullscreen)}>
                          {miniAppFullscreen ? "Window" : "Fullscreen"}
                        </button>
                      </div>
                    </div>

                    <div className="battle-enemy-hand-layer">
                      <div className="enemy-hand">
                        {Array.from({ length: enemyPlayer.hand_count ?? 0 }).map((_, index, array) => {
                          const offset = index - (array.length - 1) / 2;
                          const fanStyle = {
                            "--fan-offset": `${offset}`,
                            "--fan-depth": `${Math.abs(offset) * 2}px`,
                            zIndex: array.length - index,
                          } as CSSProperties;
                          return <div key={`back-${index}`} className="card-back" style={fanStyle} />;
                        })}
                      </div>
                    </div>

                    <div className="battle-side-info left enemy">
                      <button
                        className={`grave-trigger grave-static ${openedGraveSide === "enemy" ? "active" : ""}`}
                        onClick={(event) => {
                          event.stopPropagation();
                          setOpenedGraveSide((prev) => (prev === "enemy" ? null : "enemy"));
                        }}
                        title="Enemy graveyard"
                      >
                        GY
                        <span>{enemyGraveyard.length}</span>
                      </button>
                    </div>

                    <div className="battle-side-info right enemy">
                      <div className="deck-trigger deck-static" aria-label="Enemy deck">
                        <span className="deck-trigger-stack" aria-hidden="true">
                          <span />
                          <span />
                          <span />
                        </span>
                        <span className="deck-trigger-count">{enemyDisplayedDeckCount}</span>
                      </div>
                    </div>

                    <div className="battle-enemy-hero-layer">
                      <div className="hero-anchor top hero-anchor-scene">
                        <button
                          className={`hero-orb-button ${canSelectEnemyHeroAsHeroSpellTarget() || heroAttackArmed || selectedOwnUnitId ? "hero-targetable" : ""} ${boardAttackAnimation?.sourceKind === "hero" && boardAttackAnimation.sourceSide === "enemy" ? "hero-motion-source" : ""} ${boardAttackAnimation?.targetIds.includes(enemyHeroEventId) ? "hero-hit" : ""}`}
                          onClick={() => void runTask(handleEnemyHeroClick)}
                          data-attack-target="enemy-hero"
                          data-hero-side="enemy"
                          style={
                            boardAttackAnimation?.sourceKind === "hero" && boardAttackAnimation.sourceSide === "enemy"
                              ? ({
                                  transform: `translate(${boardAttackAnimation.dx}px, ${boardAttackAnimation.dy}px)`,
                                } as CSSProperties)
                              : undefined
                          }
                        >
                          {renderHeroHud(enemyPlayer, enemyHeroHpPeak, true)}
                        </button>
                      </div>
                    </div>

                    <div className="battle-field-layer">
                      <div className="battlefield-core">
                        <div className="battlefield-row enemy">
                          <div className="table-line">
                            {enemyPlayer.table.map((unit, index) => renderUnitSlot(unit, "enemy", index))}
                          </div>
                        </div>
                        <div className="battlefield-midline">
                          <div
                            className={`turn-timer-line ${turnSecondsLeft <= 10 ? "danger" : ""} ${isMyTurn ? "my-turn" : "enemy-turn"}`}
                            aria-label={isMyTurn ? "Your turn timer" : "Enemy turn timer"}
                          >
                            <span
                              className="turn-timer-line-fill left"
                              style={{ width: `${Math.max(0, Math.min(50, turnProgress * 50))}%` }}
                            />
                            <span
                              className="turn-timer-line-fill right"
                              style={{ width: `${Math.max(0, Math.min(50, turnProgress * 50))}%` }}
                            />
                          </div>
                        </div>
                        <div className="battlefield-row own">
                          <div className="table-line">
                            {myPlayer.table.map((unit, index) => renderUnitSlot(unit, "own", index))}
                          </div>
                        </div>
                      </div>
                    </div>

                    <div className="battle-turn-control-layer">
                      <button className="end-turn-floating" onClick={() => void runTask(handleEndTurn)}>
                        End Turn
                      </button>
                    </div>

                    <div className="battle-side-info left own">
                      <button
                        className={`grave-trigger grave-static ${openedGraveSide === "own" ? "active" : ""}`}
                        onClick={(event) => {
                          event.stopPropagation();
                          setOpenedGraveSide((prev) => (prev === "own" ? null : "own"));
                        }}
                        title="Your graveyard"
                      >
                        GY
                        <span>{ownGraveyard.length}</span>
                      </button>
                    </div>

                    <div className="battle-side-info right own">
                      <div className="deck-trigger deck-static" aria-label="Your deck">
                        <span className="deck-trigger-stack" aria-hidden="true">
                          <span />
                          <span />
                          <span />
                        </span>
                        <span className="deck-trigger-count">{displayedDeckCount}</span>
                      </div>
                    </div>

                    <div className="battle-player-hero-layer">
                      <div className="hero-anchor bottom hero-anchor-scene">
                        <button
                          className={`hero-attack-mini ${heroAttackArmed ? "armed" : ""}`}
                          onClick={handleHeroAttackToggle}
                          title={`Hero attack (${myPlayer.hero_attack_power})`}
                        >
                          HA
                        </button>
                        <button className="hero-skill-mini" onClick={() => void runTask(handleHeroSpell)}>
                          HS
                          <span className="hero-skill-mana">{heroAbilityManaCost(myPlayer)}</span>
                        </button>
                        <div
                          className={`hero-center-wrap ${boardAttackAnimation?.sourceKind === "hero" && boardAttackAnimation.sourceSide === "own" ? "hero-motion-source" : ""} ${boardAttackAnimation?.targetIds.includes(ownHeroEventId) ? "hero-hit" : ""}`}
                          data-hero-side="own"
                          style={
                            boardAttackAnimation?.sourceKind === "hero" && boardAttackAnimation.sourceSide === "own"
                              ? ({
                                  transform: `translate(${boardAttackAnimation.dx}px, ${boardAttackAnimation.dy}px)`,
                                } as CSSProperties)
                              : undefined
                          }
                        >
                          {renderHeroHud(myPlayer, ownHeroHpPeak, false)}
                        </div>
                      </div>
                    </div>

                    <div className="battle-player-hand-layer">
                      <div
                        className={`hand-row ${myHand.length >= 9 ? "ultra-compact" : myHand.length >= 7 ? "compact" : ""}`}
                      >
                        {myHand.map((card, index) => {
                          const selected = selectedHandCardId === cardInstanceId(card);
                          const templateId = cardTemplateId(card);
                          const tone = getAssetTone(templateId);
                          const meta = cardCatalogEntry(templateId);
                          const offset = index - (myHand.length - 1) / 2;
                          const depthScale = myHand.length >= 9 ? 1.8 : myHand.length >= 7 ? 2.4 : 3;
                          const fanStyle = {
                            "--fan-offset": `${offset}`,
                            "--fan-depth": `${Math.abs(offset) * depthScale}px`,
                            zIndex: selected ? 30 : myHand.length - index,
                          } as CSSProperties;
                          return (
                            <button
                              key={cardInstanceId(card)}
                              className={`hand-card tone-${tone} ${selected ? "selected" : ""}`}
                              style={fanStyle}
                              onClick={() => setSelectedHandCardId(cardInstanceId(card))}
                            >
                              <GameCard
                                mode="hand"
                                data={{
                                  kind:
                                    "attack" in (meta ?? {}) && meta?.attack !== undefined
                                      ? "battle"
                                      : "buff",
                                  name: meta?.name || resolveAssetLabel(templateId),
                                  description: meta?.description || "",
                                  imageKey: cardImageKeyForTemplate(templateId),
                                  race: undefined,
                                  mana: meta?.mana_cost ?? 0,
                                  attack:
                                    "attack" in (meta ?? {}) && meta?.attack !== undefined
                                      ? meta.attack
                                      : undefined,
                                  hp:
                                    "health_points" in (meta ?? {}) &&
                                    meta?.health_points !== undefined
                                      ? meta.health_points
                                      : undefined,
                                  cooldown:
                                    "cooldown" in (meta ?? {}) && meta?.cooldown !== undefined
                                      ? meta.cooldown
                                      : undefined,
                                  skillCooldown:
                                    "skill_cooldown" in (meta ?? {}) &&
                                    meta?.skill_cooldown !== undefined
                                      ? meta.skill_cooldown
                                      : undefined,
                                  buffType:
                                    "buff_type" in (meta ?? {}) && meta?.buff_type !== undefined
                                      ? String(meta.buff_type)
                                      : undefined,
                                  buffValue:
                                    "buff_value" in (meta ?? {}) &&
                                    meta?.buff_value !== undefined
                                      ? Number(meta.buff_value)
                                      : undefined,
                                  duration:
                                    "duration" in (meta ?? {}) && meta?.duration !== undefined
                                      ? Number(meta.duration)
                                      : undefined,
                                }}
                              />
                            </button>
                          );
                        })}
                      </div>
                    </div>

                    {openedGraveSide && (
                      <div
                        className={`grave-panel ${openedGraveSide === "enemy" ? "top" : "bottom"}`}
                        onClick={(event) => event.stopPropagation()}
                      >
                        <div className="grave-panel-head">
                          <strong>{openedGraveSide === "own" ? "Your Graveyard" : "Enemy Graveyard"}</strong>
                          <button
                            className="grave-panel-close"
                            onClick={(event) => {
                              event.stopPropagation();
                              setOpenedGraveSide(null);
                            }}
                          >
                            x
                          </button>
                        </div>
                        <div className="grave-panel-list">
                          {openedGraveyard.length === 0 ? (
                            <p className="grave-panel-empty">No dead cards yet</p>
                          ) : (
                            openedGraveyard.map(renderGraveEntry)
                          )}
                        </div>
                      </div>
                    )}
                  </div>
                  {drawFxTick > 0 && <span key={drawFxTick} className="draw-card-fx" aria-hidden="true" />}
                </>
              )}
            </section>
          </section>
        )}
      </main>

      {showProfile && (
        <ProfilePanel me={me} matches={matches} onClose={() => setShowProfile(false)} />
      )}
      {cardPreview && (
        <div className="card-viewer-overlay" onClick={() => setCardPreview(null)}>
          <div className="card-viewer-window" onClick={(event) => event.stopPropagation()}>
            <button className="card-viewer-close" onClick={() => setCardPreview(null)}>
              X
            </button>
            <GameCard data={toGameCardData(cardPreview)} mode="viewer" />
            </div>
        </div>
      )}
      <div className="toast-stack" aria-live="polite" aria-atomic="true">
        {toasts.map((toast) => (
          <div key={toast.id} className={`toast ${toast.tone}`}>
            {toast.message}
          </div>
        ))}
      </div>
    </div>
  );
}

