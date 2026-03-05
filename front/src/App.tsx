import { useEffect, useMemo, useRef, useState } from "react";
import { getAssetTone, resolveAssetLabel } from "./assets";

type TabId = "home" | "inventory" | "battle";

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

declare global {
  interface Window {
    Telegram?: {
      WebApp?: {
        ready: () => void;
        expand: () => void;
      };
    };
  }
}

const defaultDeck: DeckEntry[] = [
  { kind: "battle", template_id: "imperial_guardian", count: 5 },
  { kind: "battle", template_id: "mechanical_knight", count: 3 },
  { kind: "battle", template_id: "drones", count: 4 },
  { kind: "buff", template_id: "adrenalin", count: 4 },
  { kind: "buff", template_id: "linear_actuator", count: 2 },
  { kind: "buff", template_id: "processor_update", count: 2 },
];

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
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

function totalDeck(entries: DeckEntry[]): number {
  return entries.reduce((sum, entry) => sum + entry.count, 0);
}

function pretty(value: unknown): string {
  return JSON.stringify(value, null, 2);
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
            <span>XP</span>
            <strong>{props.me?.xp ?? "-"}</strong>
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
  const [status, setStatus] = useState("System ready");
  const [loading, setLoading] = useState(false);
  const [devUserId, setDevUserId] = useState("1");
  const [opponentUserId, setOpponentUserId] = useState("5");
  const [showProfile, setShowProfile] = useState(false);

  const [me, setMe] = useState<MeResponse | null>(null);
  const [heroes, setHeroes] = useState<OwnedHero[]>([]);
  const [cards, setCards] = useState<CardsResponse | null>(null);
  const [deckEntries, setDeckEntries] = useState<DeckEntry[]>(defaultDeck);
  const [matches, setMatches] = useState<MatchState[]>([]);
  const [selectedMatchId, setSelectedMatchId] = useState<number | null>(null);
  const [selectedMatch, setSelectedMatch] = useState<MatchState | null>(null);

  const [actionStatus, setActionStatus] = useState("Select a card, then issue an order.");
  const [selectedHandCardId, setSelectedHandCardId] = useState("");
  const [selectedOwnUnitId, setSelectedOwnUnitId] = useState("");
  const [selectedEnemyUnitId, setSelectedEnemyUnitId] = useState("");

  const streamRef = useRef<EventSource | null>(null);

  useEffect(() => {
    const app = window.Telegram?.WebApp;
    if (!app) {
      return;
    }
    app.ready();
    app.expand();
  }, []);

  useEffect(() => {
    if (!selectedMatchId) {
      streamRef.current?.close();
      streamRef.current = null;
      return;
    }

    const stream = new EventSource(`/matches/${selectedMatchId}/stream`, {
      withCredentials: true,
    });
    stream.addEventListener("state", (event) => {
      try {
        const state = JSON.parse((event as MessageEvent<string>).data) as MatchState;
        setSelectedMatch(state);
      } catch {
        setStatus("Failed to parse SSE state");
      }
    });
    stream.onerror = () => {
      setStatus("Battle feed disconnected. Refresh match list.");
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

  async function runTask(task: () => Promise<void>) {
    setLoading(true);
    try {
      await task();
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Unknown error");
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
    if (selectedMatchId) {
      const current = data.find((match) => match.match_id === selectedMatchId) ?? null;
      setSelectedMatch(current);
    }
  }

  async function refreshMatch(matchId: number) {
    const data = await apiFetch<MatchState>(`/matches/${matchId}`);
    setSelectedMatchId(matchId);
    setSelectedMatch(data);
    setTab("battle");
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

  async function login(userId: string) {
    await apiFetch<void>(`/auth/dev?user_id=${encodeURIComponent(userId)}`, {
      method: "POST",
    });
    setDevUserId(userId);
    await refreshAll();
    setStatus(`Authenticated as user ${userId}`);
  }

  async function selectHero(heroCode: string) {
    await apiFetch("/heroes/select", {
      method: "POST",
      body: JSON.stringify({ hero_code: heroCode }),
    });
    await Promise.all([refreshMe(), refreshHeroes()]);
    setStatus(`Hero selected: ${heroCode}`);
  }

  async function saveDefaultDeck() {
    await apiFetch("/deck", {
      method: "POST",
      body: JSON.stringify({ entries: defaultDeck }),
    });
    setDeckEntries(defaultDeck);
    setStatus("Standard combat deck loaded");
  }

  async function createMatch() {
    const match = await apiFetch<MatchState>("/matches", {
      method: "POST",
      body: JSON.stringify({ opponent_user_id: Number(opponentUserId) }),
    });
    await refreshMatches();
    setSelectedMatchId(match.match_id);
    setSelectedMatch(match);
    setTab("battle");
    setStatus(`Battle #${match.match_id} deployed`);
  }

  function clearSelections() {
    setSelectedHandCardId("");
    setSelectedOwnUnitId("");
    setSelectedEnemyUnitId("");
  }

  async function applyAction(payload: Record<string, unknown>, successText: string) {
    if (!selectedMatch) {
      setStatus("No match selected");
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
  }

  const selectedCard = myPlayer?.hand?.find(
    (card) => cardInstanceId(card) === selectedHandCardId,
  );

  async function handlePlaySelectedCard(slot: number) {
    if (!selectedCard) {
      setStatus("Select a card from hand first");
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
      setStatus("No battle selected");
      return;
    }
    const heroCode = myPlayer.hero_code;
    let payload: Record<string, unknown>;

    if (heroCode === "imperial_commander" || heroCode === "black_cell" || heroCode === "karn" || heroCode === "slavic_priest") {
      if (!selectedOwnUnitId) {
        setStatus("Select your unit for this hero ability");
        return;
      }
      payload = {
        type: "hero_spell",
        target_instance_id: selectedOwnUnitId,
      };
    } else if (heroCode === "the_system") {
      if (!selectedEnemyUnitId) {
        setStatus("Select an enemy unit for this hero ability");
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

  function renderHeroGlyph(heroCode: string, imageKey: string | undefined, size: "small" | "large") {
    const tone = getAssetTone(heroCode);
    const label = resolveAssetLabel(imageKey || heroCode || "hero");
    return (
      <div className={`hero-glyph ${size} tone-${tone}`}>
        <span>{label}</span>
      </div>
    );
  }

  function renderUnitSlot(unit: UnitState | null, side: "own" | "enemy", slot: number) {
    if (!unit) {
      return <button className="slot empty" onClick={() => handlePlaySelectedCard(slot)}>+</button>;
    }

    const selected = side === "own" ? selectedOwnUnitId === unitInstanceId(unit) : selectedEnemyUnitId === unitInstanceId(unit);
    const tone = getAssetTone(unitTemplateId(unit));

    return (
      <button
        className={`slot tone-${tone} ${selected ? "selected" : ""}`}
        onClick={() =>
          void (side === "own" ? handleOwnUnitClick(unit) : handleEnemyUnitClick(unit))
        }
      >
        <strong>{resolveAssetLabel(unitTemplateId(unit))}</strong>
        <span>HP {unitHP(unit)}/{unitMaxHP(unit)}</span>
        <span>ATK {unitAttack(unit)}</span>
        <span>CD {unitCooldown(unit)}</span>
      </button>
    );
  }

  return (
    <div className="war-shell">
      <header className="top-frame">
        <button className="avatar-trigger" onClick={() => setShowProfile(true)}>
          <span className="avatar-core">
            {(me?.first_name?.[0] || me?.username?.[0] || "?").toUpperCase()}
          </span>
        </button>
        <div className="command-title">
          <p className="eyebrow">TheWar / command bridge</p>
          <h1>Alpha Theatre</h1>
          <p className="muted">
            Dark war-console shell for Telegram Mini App flow.
          </p>
        </div>
        <div className="status-rack">
          <span className={`status-pill ${loading ? "hot" : "cold"}`}>
            {loading ? "Operational" : "Standby"}
          </span>
          <span className="status-pill cold">{status}</span>
        </div>
      </header>

      <nav className="battle-nav">
        <button className={tab === "home" ? "nav-pill active" : "nav-pill"} onClick={() => setTab("home")}>
          Start Game
        </button>
        <button className={tab === "inventory" ? "nav-pill active" : "nav-pill"} onClick={() => setTab("inventory")}>
          Inventory
        </button>
        <button className={tab === "battle" ? "nav-pill active" : "nav-pill"} onClick={() => setTab("battle")}>
          Battle
        </button>
      </nav>

      <main className="view-frame">
        {tab === "home" && (
          <section className="screen-grid">
            <div className="panel command-panel">
              <h2>Command Link</h2>
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
              <div className="quick-login">
                <button onClick={() => void runTask(() => login("1"))}>Quick Login</button>
                <button onClick={() => void runTask(refreshAll)}>Refresh Intel</button>
              </div>
              <div className="hero-banner">
                {renderHeroGlyph(
                  me?.selected_hero_code || "unassigned",
                  heroes.find((hero) => hero.hero_code === me?.selected_hero_code)?.image_key,
                  "large",
                )}
                <div>
                  <h3>{me?.selected_hero_name || "No Hero Assigned"}</h3>
                  <p className="muted">
                    {me?.first_name || me?.username || "No profile loaded"}
                  </p>
                  <p className="muted">Rating {me?.rating ?? "-"} / XP {me?.xp ?? "-"}</p>
                </div>
              </div>
            </div>

            <div className="panel command-panel">
              <h2>Deployment</h2>
              <label>
                Opponent user id
                <input
                  value={opponentUserId}
                  onChange={(event) => setOpponentUserId(event.target.value)}
                />
              </label>
              <div className="quick-login">
                <button onClick={() => void runTask(createMatch)}>Start Battle</button>
                <button onClick={() => void runTask(saveDefaultDeck)}>Load Standard Deck</button>
              </div>
              <div className="match-list compact">
                {matches.length === 0 ? (
                  <p className="muted">No battle history yet.</p>
                ) : (
                  matches.map((match) => (
                    <button
                      key={match.match_id}
                      className={selectedMatchId === match.match_id ? "match-pill active" : "match-pill"}
                      onClick={() => void runTask(() => refreshMatch(match.match_id))}
                    >
                      <span>Battle #{match.match_id}</span>
                      <span>{match.result}</span>
                    </button>
                  ))
                )}
              </div>
            </div>

            <div className="panel hero-select">
              <h2>War Council</h2>
              <div className="hero-grid">
                {heroes.map((hero) => (
                  <button
                    key={hero.hero_code}
                    className="hero-card"
                    onClick={() => void runTask(() => selectHero(hero.hero_code))}
                  >
                    {renderHeroGlyph(hero.hero_code, hero.image_key, "small")}
                    <div>
                      <strong>{hero.name}</strong>
                      <span>HP {hero.health_points}</span>
                      <span>ATK {hero.attack_power}</span>
                    </div>
                  </button>
                ))}
              </div>
            </div>
          </section>
        )}

        {tab === "inventory" && (
          <section className="screen-grid">
            <div className="panel inventory-panel">
              <div className="section-head">
                <h2>Deck Doctrine</h2>
                <button onClick={() => void runTask(saveDefaultDeck)}>Load Standard Deck</button>
              </div>
              <div className="deck-summary">
                <span>Total cards</span>
                <strong>{totalDeck(deckEntries)}</strong>
              </div>
              <pre>{pretty(deckEntries)}</pre>
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
                      <span>{resolveAssetLabel(card.image_key)}</span>
                    </div>
                    <strong>{card.name}</strong>
                    <span>{card.template_id}</span>
                    <span>HP {card.health_points} / ATK {card.attack}</span>
                    <span>{card.copies} copies</span>
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
                      <span>{resolveAssetLabel(card.image_key)}</span>
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

        {tab === "battle" && (
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

            <section className="battle-board panel">
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
                      {Array.from({ length: enemyPlayer.hand_count ?? 0 }).map((_, index) => (
                        <div key={`back-${index}`} className="card-back" />
                      ))}
                    </div>
                    <div className="hero-anchor top">
                      <button className="hero-anchor-button" onClick={() => void runTask(handleEnemyHeroClick)}>
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
                      {(myPlayer.hand ?? []).map((card) => {
                        const selected = selectedHandCardId === cardInstanceId(card);
                        const templateId = cardTemplateId(card);
                        const tone = getAssetTone(templateId);
                        return (
                          <button
                            key={cardInstanceId(card)}
                            className={`hand-card tone-${tone} ${selected ? "selected" : ""}`}
                            onClick={() => setSelectedHandCardId(cardInstanceId(card))}
                          >
                            <div className="asset-frame compact">
                              <span>{resolveAssetLabel(templateId)}</span>
                            </div>
                            <strong>{templateId}</strong>
                            <span>{cardKind(card)}</span>
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
    </div>
  );
}
