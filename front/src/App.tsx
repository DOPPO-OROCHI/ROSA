import { type CSSProperties, useEffect, useMemo, useRef, useState } from "react";
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
import { bootstrapTelegramWebApp } from "./telegram";

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
  image_key: string;
  attack?: number;
  health_points?: number;
  cooldown?: number;
  buff_value?: number;
  buff_type?: string;
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
  Cooldown?: number;
  IsTank?: boolean;
  Effects?: Array<{ EffectType?: string; TurnsLeft?: number; Value?: number }>;
  instance_id?: string;
  template_id?: string;
  hp?: number;
  attack?: number;
  max_hp?: number;
  cooldown?: number;
  is_tank?: boolean;
  effects?: Array<{ effect_type?: string; turns_left?: number; value?: number }>;
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
  mana: number;
  turns: number;
  table: Array<UnitState | null>;
  hand?: CardsInMatch[];
  deck?: CardsInMatch[];
  discard?: CardsInMatch[];
  hand_count?: number;
  deck_count?: number;
  disc_count?: number;
};

type MatchEvent = {
  type: string;
  vfx_key?: string;
  sfx_key?: string;
  source_template_id?: string;
  source_card_template_id?: string;
};

type MatchState = {
  match_id: number;
  version: number;
  active_player: number;
  phase: string;
  finished: boolean;
  result: string;
  players: [MatchPlayer | null, MatchPlayer | null];
  events?: MatchEvent[];
};

type DragAttackState = {
  sourceId: string;
  currentX: number;
  currentY: number;
};

type ToastEntry = {
  id: number;
  message: string;
  tone: "info" | "error";
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
  const response = await fetch(apiUrl(path), {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
    ...init,
  });

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
  const [tab, setTab] = useState<TabId>("home");
  const [loading, setLoading] = useState(false);
  const [devUserId, setDevUserId] = useState("1");
  const [opponentUserId, setOpponentUserId] = useState("5");
  const [showProfile, setShowProfile] = useState(false);
  const [toasts, setToasts] = useState<ToastEntry[]>([]);

  const [me, setMe] = useState<MeResponse | null>(null);
  const [heroes, setHeroes] = useState<OwnedHero[]>([]);
  const [cards, setCards] = useState<CardsResponse | null>(null);
  const [deckEntries, setDeckEntries] = useState<DeckEntry[]>(defaultDeck);
  const [deckInspectorKey, setDeckInspectorKey] = useState<string | null>(null);
  const [heroPickerOpen, setHeroPickerOpen] = useState(false);
  const [heldHero, setHeldHero] = useState<OwnedHero | null>(null);
  const [matches, setMatches] = useState<MatchState[]>([]);
  const [selectedMatchId, setSelectedMatchId] = useState<number | null>(null);
  const [selectedMatch, setSelectedMatch] = useState<MatchState | null>(null);

  const [actionStatus, setActionStatus] = useState("Select a card, then issue an order.");
  const [selectedHandCardId, setSelectedHandCardId] = useState("");
  const [selectedOwnUnitId, setSelectedOwnUnitId] = useState("");
  const [selectedEnemyUnitId, setSelectedEnemyUnitId] = useState("");

  const streamRef = useRef<EventSource | null>(null);
  const battleBoardRef = useRef<HTMLElement | null>(null);
  const [dragAttack, setDragAttack] = useState<DragAttackState | null>(null);
  const toastIdRef = useRef(1);
  const heroHoldTimerRef = useRef<number | null>(null);
  const heroHoldTriggeredRef = useRef(false);

  function pushToast(message: string, tone: ToastEntry["tone"] = "info") {
    const id = toastIdRef.current++;
    setToasts((prev) => [...prev, { id, message, tone }]);
    window.setTimeout(() => {
      setToasts((prev) => prev.filter((entry) => entry.id !== id));
    }, 1200);
  }

  useEffect(() => {
    bootstrapTelegramWebApp();
  }, []);

  useEffect(() => {
    if (!selectedMatchId) {
      streamRef.current?.close();
      streamRef.current = null;
      return;
    }

    const stream = new EventSource(apiUrl(`/matches/${selectedMatchId}/stream`), {
      withCredentials: true,
    });
    stream.addEventListener("state", (event) => {
      try {
        const state = JSON.parse((event as MessageEvent<string>).data) as MatchState;
        setSelectedMatch(state);
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

  const myPlayer = useMemo(
    () => selectedMatch?.players.find((player) => player?.user_id === me?.user_id) ?? null,
    [selectedMatch, me],
  );
  const enemyPlayer = useMemo(
    () => selectedMatch?.players.find((player) => player?.user_id !== me?.user_id) ?? null,
    [selectedMatch, me],
  );
  const activeBattle = Boolean(selectedMatch && !selectedMatch.finished);
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
      setSelectedMatch(current);
      if (!current || current.finished) {
        setSelectedMatchId(null);
      }
      return;
    }
    if (active) {
      setSelectedMatchId(active.match_id);
      setSelectedMatch(active);
      pushToast(`Battle #${active.match_id} ready`);
    }
  }

  async function refreshMatch(matchId: number) {
    const data = await apiFetch<MatchState>(`/matches/${matchId}`);
    if (data.finished) {
      setSelectedMatchId(null);
      setSelectedMatch(null);
      pushToast(`Battle #${matchId} already finished`);
      return;
    }
    setSelectedMatchId(matchId);
    setSelectedMatch(data);
  }

  async function refreshAll() {
    await Promise.all([
      refreshMe(),
      refreshHeroes(),
      refreshCards(),
      refreshDeck(),
      refreshMatches(),
    ]);
  }

  useEffect(() => {
    if (!me || activeBattle) {
      return;
    }
    const pollId = window.setInterval(() => {
      void refreshMatches();
    }, 3000);
    return () => window.clearInterval(pollId);
  }, [me, activeBattle]);

  async function login(userId: string) {
    await apiFetch<void>(`/auth/dev?user_id=${encodeURIComponent(userId)}`, {
      method: "POST",
    });
    setDevUserId(userId);
    await refreshAll();
    pushToast(`Authenticated as user ${userId}`);
  }

  async function selectHero(heroCode: string) {
    await apiFetch("/heroes/select", {
      method: "POST",
      body: JSON.stringify({ hero_code: heroCode }),
    });
    await Promise.all([refreshMe(), refreshHeroes()]);
    pushToast(`Hero selected: ${heroCode}`);
  }

  function clearHeroHoldTimer() {
    if (heroHoldTimerRef.current !== null) {
      window.clearTimeout(heroHoldTimerRef.current);
      heroHoldTimerRef.current = null;
    }
  }

  function beginHeroHold(hero: OwnedHero) {
    clearHeroHoldTimer();
    heroHoldTriggeredRef.current = false;
    heroHoldTimerRef.current = window.setTimeout(() => {
      heroHoldTriggeredRef.current = true;
      setHeldHero(hero);
    }, 360);
  }

  function endHeroHold() {
    clearHeroHoldTimer();
    setHeldHero(null);
    window.setTimeout(() => {
      heroHoldTriggeredRef.current = false;
    }, 0);
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

  async function createMatch() {
    const match = await apiFetch<MatchState>("/matches", {
      method: "POST",
      body: JSON.stringify({ opponent_user_id: Number(opponentUserId) }),
    });
    await refreshMatches();
    setSelectedMatchId(match.match_id);
    setSelectedMatch(match);
    pushToast(`Battle #${match.match_id} deployed`);
  }

  function clearSelections() {
    setSelectedHandCardId("");
    setSelectedOwnUnitId("");
    setSelectedEnemyUnitId("");
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
    setSelectedMatch(next);
    await refreshMatches();
    setActionStatus(successText);
    clearSelections();
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

  async function handlePlaySelectedCard(slot: number) {
    if (!selectedCard) {
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
    const heroCode = myPlayer.hero_code;
    let payload: Record<string, unknown>;

    if (heroCode === "imperial_commander" || heroCode === "black_cell" || heroCode === "karn" || heroCode === "slavic_priest") {
      if (!selectedOwnUnitId) {
        pushToast("Select your unit for this hero ability", "error");
        return;
      }
      payload = {
        type: "hero_spell",
        target_instance_id: selectedOwnUnitId,
      };
    } else if (heroCode === "the_system") {
      if (!selectedEnemyUnitId) {
        pushToast("Select an enemy unit for this hero ability", "error");
        return;
      }
      payload = {
        type: "hero_spell",
        target_instance_id: selectedEnemyUnitId,
      };
    } else if (heroCode === "suprime_lider") {
      payload = selectedEnemyUnitId
        ? {
            type: "hero_spell",
            target_instance_id: selectedEnemyUnitId,
          }
        : {
            type: "hero_spell",
            attack_hero: true,
          };
    } else {
      payload = { type: "hero_spell", attack_hero: true };
    }

    await applyAction(payload, "Hero ability activated");
  }

  async function handleLeaveMatch() {
    await applyAction({ type: "leave_match" }, "You left the battle");
  }

  async function handleOwnUnitClick(unit: UnitState) {
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

  function startUnitDrag(unit: UnitState, clientX: number, clientY: number) {
    if (selectedCard || !selectedMatch || !myPlayer) {
      return;
    }
    const rect = battleBoardRef.current?.getBoundingClientRect();
    if (!rect) {
      return;
    }
    setSelectedOwnUnitId(unitInstanceId(unit));
    setDragAttack({
      sourceId: unitInstanceId(unit),
      currentX: clientX - rect.left,
      currentY: clientY - rect.top,
    });
    setActionStatus(`Attack vector: ${resolveAssetLabel(unitTemplateId(unit))}`);
  }

  function dragSourcePoint() {
    if (!dragAttack || !battleBoardRef.current) {
      return null;
    }
    const sourceNode = battleBoardRef.current.querySelector<HTMLElement>(
      `[data-unit-id="${dragAttack.sourceId}"][data-slot-side="own"]`,
    );
    if (!sourceNode) {
      return null;
    }
    const boardRect = battleBoardRef.current.getBoundingClientRect();
    const sourceRect = sourceNode.getBoundingClientRect();
    return {
      x: sourceRect.left + sourceRect.width / 2 - boardRect.left,
      y: sourceRect.top + sourceRect.height / 2 - boardRect.top,
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

    const selected = side === "own" ? selectedOwnUnitId === unitInstanceId(unit) : selectedEnemyUnitId === unitInstanceId(unit);
    const tone = getAssetTone(unitTemplateId(unit));
    const meta = cardCatalogEntry(unitTemplateId(unit));

    return (
      <button
        className={`slot tone-${tone} ${selected ? "selected" : ""}`}
        data-unit-id={unitInstanceId(unit)}
        data-slot-side={side}
        data-attack-target={side === "enemy" ? "enemy-unit" : undefined}
        onClick={() =>
          void (side === "own" ? handleOwnUnitClick(unit) : handleEnemyUnitClick(unit))
        }
        onPointerDown={
          side === "own"
            ? (event) => {
                event.preventDefault();
                startUnitDrag(unit, event.clientX, event.clientY);
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
        <div className="slot-topline">
          <span className="card-chip mana">{meta?.mana_cost ?? 0}</span>
          <span className="card-chip side">{side === "own" ? "ALLY" : "HOSTILE"}</span>
        </div>
        <div className="slot-stats">
          <span className="slot-stat hp">{unitHP(unit)}</span>
          <span className="slot-stat atk">{unitAttack(unit)}</span>
          <span className="slot-stat cd">{unitCooldown(unit)}</span>
        </div>
      </button>
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

  return (
    <div className="war-shell">
      {!activeBattle && (
        <nav className="battle-nav two-up">
          <button className={tab === "home" ? "nav-pill active" : "nav-pill"} onClick={() => setTab("home")}>
            Start Game
          </button>
          <button className={tab === "inventory" ? "nav-pill active" : "nav-pill"} onClick={() => setTab("inventory")}>
            Inventory
          </button>
        </nav>
      )}

      <main className="view-frame">
        {!activeBattle && tab === "home" && (
          <section className="screen-grid home-grid">
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
                          <button
                            key={hero.hero_code}
                            className={`hero-tile ${selected ? "selected" : ""}`}
                            onPointerDown={() => beginHeroHold(hero)}
                            onPointerUp={endHeroHold}
                            onPointerLeave={endHeroHold}
                            onPointerCancel={endHeroHold}
                            onClick={() => {
                              if (heroHoldTriggeredRef.current) {
                                return;
                              }
                              void runTask(async () => {
                                await selectHero(hero.hero_code);
                                setHeroPickerOpen(false);
                                setHeldHero(null);
                              });
                            }}
                          >
                            {renderHeroGlyph(hero.hero_code, hero.image_key, "small")}
                          </button>
                        );
                      })}
                    </div>
                  </div>
                </div>
              )}
              <div className="login-row">
                <label>
                  Active user
                  <input
                    value={devUserId}
                    onChange={(event) => setDevUserId(event.target.value)}
                  />
                </label>
                <button onClick={() => void runTask(() => login(devUserId))}>Login</button>
              </div>
              <label>
                Opponent user id
                <input
                  value={opponentUserId}
                  onChange={(event) => setOpponentUserId(event.target.value)}
                />
              </label>
              <button onClick={() => void runTask(createMatch)}>Start Battle</button>
            </div>

          </section>
        )}

        {!activeBattle && tab === "inventory" && (
          <section className="screen-grid">
            <div className="panel inventory-panel">
              <div className="section-head">
                <h2>Deck Doctrine</h2>
                <button onClick={() => void runTask(saveDefaultDeck)}>Load Standard Deck</button>
              </div>
              <div className="deck-summary">
                <span>Total cards</span>
                <strong>{deckTotal}</strong>
              </div>
              {!deckReady && <p className="deck-warning">Дека не собрана (нужно 20 карт)</p>}
              <div className="deck-grid">
                {deckGroups.map((group) => (
                  <button
                    key={group.key}
                    className="deck-slot filled interactive"
                    onClick={() => setDeckInspectorKey(group.key)}
                  >
                    <AssetImage
                      imageKey={group.imageKey}
                      alt={group.name}
                      fallbackSrc={resolveCardFallbackSrc()}
                      className="deck-slot-media"
                    />
                    <div className="deck-slot-meta">
                      <span className="deck-slot-mana">{group.mana}</span>
                      <strong>{group.name}</strong>
                    </div>
                    <span className="deck-slot-count">x{group.count}</span>
                  </button>
                ))}
                {deckGroups.length === 0 && (
                  <article className="deck-slot empty deck-slot-empty-wide">
                    <span>Deck is empty</span>
                  </article>
                )}
              </div>
              {inspectedDeckGroup && (
                <div className="deck-inspector" onClick={() => setDeckInspectorKey(null)}>
                  <div className="deck-inspector-body" onClick={(event) => event.stopPropagation()}>
                    <div className="deck-inspector-head">
                      <strong>{inspectedDeckGroup.name}</strong>
                      <button className="deck-inspector-close" onClick={() => setDeckInspectorKey(null)}>
                        x
                      </button>
                    </div>
                    <div className="deck-inspector-grid">
                      {Array.from({ length: inspectedDeckGroup.count }).map((_, index) => (
                        <article key={`${inspectedDeckGroup.key}:${index}`} className="deck-copy-card">
                          <AssetImage
                            imageKey={inspectedDeckGroup.imageKey}
                            alt={inspectedDeckGroup.name}
                            fallbackSrc={resolveCardFallbackSrc()}
                            className="deck-copy-media"
                          />
                          <button
                            className="deck-copy-remove"
                            onClick={() => void runTask(() => removeCardFromDeck(inspectedDeckGroup.kind, inspectedDeckGroup.templateId))}
                          >
                            x
                          </button>
                        </article>
                      ))}
                    </div>
                  </div>
                </div>
              )}
            </div>

            <div className="panel inventory-panel">
              <div className="section-head">
                <h2>Battle Cards</h2>
                <button onClick={() => void runTask(refreshCards)}>Refresh Cards</button>
              </div>
              <div className="asset-grid">
                {cards?.battle.map((card) => (
                  <article key={card.template_id} className={`asset-card tone-${getAssetTone(card.asset_base_key)}`}>
                    <div className="asset-frame">
                      <AssetImage
                        imageKey={card.image_key || resolveBattleCardImageKey(card.template_id)}
                        alt={card.name}
                        fallbackSrc={resolveCardFallbackSrc()}
                        className="asset-frame-media"
                      />
                      <button
                        className="asset-add"
                        onClick={() => void runTask(() => addCardToDeck("battle", card.template_id))}
                      >
                        +
                      </button>
                    </div>
                    <strong>{card.name}</strong>
                    <span>HP {card.health_points}</span>
                    <span>ATK {card.attack}</span>
                    <span>CD {card.cooldown}</span>
                    <span>Max copies {card.max_copies}</span>
                  </article>
                ))}
              </div>
            </div>

            <div className="panel inventory-panel span-all">
              <h2>Buff Cards</h2>
              <div className="asset-grid">
                {cards?.buff.map((card) => (
                  <article key={card.template_id} className={`asset-card tone-${getAssetTone(card.asset_base_key)}`}>
                    <div className="asset-frame">
                      <AssetImage
                        imageKey={card.image_key || resolveBuffCardImageKey(card.template_id)}
                        alt={card.name}
                        fallbackSrc={resolveCardFallbackSrc()}
                        className="asset-frame-media"
                      />
                      <span>{resolveAssetLabel(card.image_key || card.template_id)}</span>
                      <button
                        className="asset-add"
                        onClick={() => void runTask(() => addCardToDeck("buff", card.template_id))}
                      >
                        +
                      </button>
                    </div>
                    <strong>{card.name}</strong>
                    <span>{card.template_id}</span>
                    <span>{card.buff_type}</span>
                    <span>Value {card.buff_value}</span>
                    <span>{card.copies} copies</span>
                  </article>
                ))}
              </div>
            </div>
          </section>
        )}

        {activeBattle && (
          <section className="battle-screen">
            <aside className="battle-side left">
              <button className="command-button" onClick={() => void runTask(handleEndTurn)}>
                End Turn
              </button>
              <button className="command-button" onClick={() => void runTask(handleHeroSpell)}>
                Hero Skill
              </button>
              <button className="command-button danger" onClick={() => void runTask(handleLeaveMatch)}>
                Leave Match
              </button>
              <div className="debug-box">
                <strong>Orders</strong>
                <span>{actionStatus}</span>
              </div>
            </aside>

            <section
              className="battle-board panel"
              ref={battleBoardRef}
            >
              <div
                className="battle-board-background"
                style={{ backgroundImage: `url(${resolveBoardBackgroundSrc()})` }}
              />
              {dragAttack && dragSourcePoint() && (
                <svg className="attack-drag-layer" viewBox="0 0 100 100" preserveAspectRatio="none">
                  <defs>
                    <marker id="attack-arrowhead" markerWidth="8" markerHeight="8" refX="6" refY="3" orient="auto">
                      <polygon points="0 0, 6 3, 0 6" fill="#79c2d6" />
                    </marker>
                  </defs>
                  {(() => {
                    const source = dragSourcePoint();
                    if (!source) {
                      return null;
                    }
                    return (
                  <line
                    x1={`${(source.x / Math.max(1, battleBoardRef.current?.clientWidth ?? 1)) * 100}`}
                    y1={`${(source.y / Math.max(1, battleBoardRef.current?.clientHeight ?? 1)) * 100}`}
                    x2={`${(dragAttack.currentX / Math.max(1, battleBoardRef.current?.clientWidth ?? 1)) * 100}`}
                    y2={`${(dragAttack.currentY / Math.max(1, battleBoardRef.current?.clientHeight ?? 1)) * 100}`}
                    className="attack-drag-line"
                    markerEnd="url(#attack-arrowhead)"
                  />
                    );
                  })()}
                </svg>
              )}
              {!selectedMatch || !myPlayer || !enemyPlayer ? (
                <div className="empty-battle">
                  <h2>No active battle selected</h2>
                  <p className="muted">Create a match from Start Game and enter the theatre.</p>
                </div>
              ) : (
                <>
                  <div className="enemy-zone">
                    <div className="enemy-stats">
                      <span>Enemy mana {enemyPlayer.mana}</span>
                      <span>Enemy HP {enemyPlayer.hero_hp}</span>
                      <span>Phase {selectedMatch.phase}</span>
                    </div>
                    <div className="enemy-hand">
                      {Array.from({ length: enemyPlayer.hand_count ?? 0 }).map((_, index, array) => {
                        const offset = index - (array.length - 1) / 2;
                        const fanStyle = {
                          "--fan-offset": `${offset}`,
                          "--fan-depth": `${Math.abs(offset) * 2}px`,
                          zIndex: array.length - index,
                        } as CSSProperties;
                        return (
                          <div key={`back-${index}`} className="card-back" style={fanStyle} />
                        );
                      })}
                    </div>
                    <div className="hero-anchor top">
                      <button
                        className="hero-anchor-button"
                        onClick={() => void runTask(handleEnemyHeroClick)}
                        data-attack-target="enemy-hero"
                      >
                        {renderHeroGlyph(
                          enemyPlayer.hero_code,
                          `heroes/${enemyPlayer.hero_code}/image`,
                          "large",
                        )}
                      </button>
                    </div>
                    <div className="table-line">
                      {enemyPlayer.table.map((unit, index) => renderUnitSlot(unit, "enemy", index))}
                    </div>
                  </div>

                  <div className="battle-midline">
                    <div className="turn-indicator">
                      <span>Battle #{selectedMatch.match_id}</span>
                      <strong>
                        {selectedMatch.active_player === myPlayer.player_id
                          ? "Your Turn"
                          : "Enemy Turn"}
                      </strong>
                    </div>
                    <div className="event-feed">
                      {(selectedMatch.events ?? []).slice(-3).map((event, index) => (
                        <div key={`${event.type}-${index}`} className="event-line">
                          <span>{event.type}</span>
                          <span>{event.vfx_key || "no vfx"}</span>
                        </div>
                      ))}
                    </div>
                  </div>

                  <div className="ally-zone">
                    <div className="table-line">
                      {myPlayer.table.map((unit, index) => renderUnitSlot(unit, "own", index))}
                    </div>
                    <div className="hero-anchor bottom">
                      <div className="hero-anchor-button passive">
                        {renderHeroGlyph(
                          myPlayer.hero_code,
                          `heroes/${myPlayer.hero_code}/image`,
                          "large",
                        )}
                      </div>
                    </div>
                    <div className="ally-stats">
                      <span>Mana {myPlayer.mana}</span>
                      <span>HP {myPlayer.hero_hp}</span>
                      <span>Deck {myPlayer.deck?.length ?? myPlayer.deck_count ?? 0}</span>
                    </div>
                    <div className="hand-row">
                      {myHand.map((card, index) => {
                        const selected = selectedHandCardId === cardInstanceId(card);
                        const templateId = cardTemplateId(card);
                        const tone = getAssetTone(templateId);
                        const meta = cardCatalogEntry(templateId);
                        const offset = index - (myHand.length - 1) / 2;
                        const fanStyle = {
                          "--fan-offset": `${offset}`,
                          "--fan-depth": `${Math.abs(offset) * 3}px`,
                          zIndex: selected ? 30 : myHand.length - index,
                        } as CSSProperties;
                        return (
                          <button
                            key={cardInstanceId(card)}
                            className={`hand-card tone-${tone} ${selected ? "selected" : ""}`}
                            style={fanStyle}
                            onClick={() => setSelectedHandCardId(cardInstanceId(card))}
                          >
                              <div className="hand-card-topline">
                                <span className="card-chip mana">{meta?.mana_cost ?? "?"}</span>
                                <span className="card-chip kind">{cardKind(card)}</span>
                              </div>
                              <div className="asset-frame compact">
                                <AssetImage
                                  imageKey={cardImageKeyForTemplate(templateId)}
                                  alt={templateId}
                                  fallbackSrc={resolveCardFallbackSrc()}
                                  className="asset-frame-media"
                                />
                                <span>{resolveAssetLabel(templateId)}</span>
                              </div>
                              <div className="hand-card-stats">
                                {"attack" in (meta ?? {}) && meta?.attack !== undefined ? <span>ATK {meta.attack}</span> : <span>{meta?.buff_type || cardKind(card)}</span>}
                                {"health_points" in (meta ?? {}) && meta?.health_points !== undefined ? <span>HP {meta.health_points}</span> : <span>VAL {meta?.buff_value ?? "-"}</span>}
                                {"cooldown" in (meta ?? {}) && meta?.cooldown !== undefined ? <span>CD {meta.cooldown}</span> : <span>LVL {card.card_level ?? card.CardLevel ?? 1}</span>}
                              </div>
                            </button>
                          );
                        })}
                    </div>
                  </div>
                </>
              )}
            </section>
          </section>
        )}
      </main>

      {showProfile && (
        <ProfilePanel me={me} matches={matches} onClose={() => setShowProfile(false)} />
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

