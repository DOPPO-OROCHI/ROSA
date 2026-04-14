import { useEffect, useState } from "react";
import { GameModePanel } from "./components/GameModePanel";
import { InventoryScreen } from "./components/InventoryScreen";
import { MainMenu } from "./components/MainMenu";
import { request } from "./lib/api";
import type {
  BattleCard,
  BuffCard,
  CardsResponse,
  DeckEntry,
  DeckResponse,
  Hero,
  HeroesResponse,
  JoinQueueResponse,
  MeResponse,
  QueueStatusResponse,
} from "./types";

const QUICK_USERS = ["dev", "roman", "test", "rosa"];

type Screen = "menu" | "inventory";

export function App() {
  const [screen, setScreen] = useState<Screen>("menu");
  const [username, setUsername] = useState("dev");
  const [userId, setUserId] = useState("");
  const [me, setMe] = useState<MeResponse | null>(null);
  const [heroes, setHeroes] = useState<Hero[]>([]);
  const [battleCards, setBattleCards] = useState<BattleCard[]>([]);
  const [buffCards, setBuffCards] = useState<BuffCard[]>([]);
  const [deckEntries, setDeckEntries] = useState<DeckEntry[]>([]);
  const [draftDeckEntries, setDraftDeckEntries] = useState<DeckEntry[]>([]);
  const [busy, setBusy] = useState(true);
  const [error, setError] = useState("");
  const [selectedHeroCode, setSelectedHeroCode] = useState("");
  const [heroPickerOpen, setHeroPickerOpen] = useState(false);
  const [gameModeOpen, setGameModeOpen] = useState(false);
  const [queueBusy, setQueueBusy] = useState(false);
  const [queueError, setQueueError] = useState("");
  const [queueStatus, setQueueStatus] = useState<QueueStatusResponse>({ state: "idle" });

  const selectedHero = heroes.find((hero) => hero.hero_code === selectedHeroCode) ?? heroes[0] ?? null;
  const deckCardCount = deckEntries.reduce((total, entry) => total + entry.count, 0);
  const canJoinQueue = Boolean(selectedHeroCode) && deckCardCount === 20;
  const queueSearching = queueStatus.state === "searching" || queueStatus.state === "pending_match";

  async function loadQueueStatus() {
    const status = await request<QueueStatusResponse>("/queue/status");
    setQueueStatus(status);
    setQueueError("");
  }

  async function loadSession() {
    const [nextMe, nextHeroes] = await Promise.all([
      request<MeResponse>("/me"),
      request<HeroesResponse>("/heroes"),
    ]);
    setMe(nextMe);
    setHeroes(nextHeroes.heroes);
    setSelectedHeroCode(nextMe.selected_hero_code || nextHeroes.heroes[0]?.hero_code || "");
    setError("");
  }

  async function loadInventory() {
    const [cards, deck] = await Promise.all([
      request<CardsResponse>("/cards"),
      request<DeckResponse>("/deck"),
    ]);
    setBattleCards(cards.battle);
    setBuffCards(cards.buff);
    setDeckEntries(deck.entries);
    setDraftDeckEntries(deck.entries);
  }

  useEffect(() => {
    loadSession()
      .then(async () => {
        await loadInventory();
        await loadQueueStatus();
      })
      .catch(() => {
        setMe(null);
        setHeroes([]);
        setBattleCards([]);
        setBuffCards([]);
        setDeckEntries([]);
        setDraftDeckEntries([]);
        setSelectedHeroCode("");
        setQueueStatus({ state: "idle" });
      })
      .finally(() => setBusy(false));
  }, []);

  useEffect(() => {
    if (!queueSearching) {
      return;
    }

    const id = window.setInterval(() => {
      void loadQueueStatus().catch(() => {
        setQueueStatus((current) => ({ ...current, state: "idle" }));
      });
    }, 1000);

    return () => window.clearInterval(id);
  }, [queueSearching]);

  async function login(nextUsername: string) {
    setBusy(true);
    setError("");
    try {
      await request("/auth/dev", {
        method: "POST",
        body: JSON.stringify({ username: nextUsername }),
      });
      await loadSession();
      await loadInventory();
      await loadQueueStatus();
      setUsername(nextUsername);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setBusy(false);
    }
  }

  async function loginAsUserId() {
    const parsed = Number(userId);
    if (!Number.isInteger(parsed) || parsed <= 0) {
      setError("Введите корректный user ID");
      return;
    }

    setBusy(true);
    setError("");
    try {
      await request("/auth/dev", {
        method: "POST",
        body: JSON.stringify({ user_id: parsed }),
      });
      await loadSession();
      await loadInventory();
      await loadQueueStatus();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setBusy(false);
    }
  }

  async function chooseHero(hero: Hero) {
    setError("");
    try {
      await request("/heroes/select", {
        method: "POST",
        body: JSON.stringify({ hero_code: hero.hero_code }),
      });
      const nextMe = await request<MeResponse>("/me");
      setSelectedHeroCode(hero.hero_code);
      setMe(nextMe);
      setHeroPickerOpen(false);
      setScreen("menu");
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to select hero";
      setError(message);
      throw new Error(message);
    }
  }

  async function joinQueue() {
    if (!selectedHeroCode) {
      setQueueError("СНАЧАЛА ВЫБЕРИТЕ ПЕРСОНАЖА");
      return;
    }

    if (deckCardCount !== 20) {
      setQueueError("КОЛОДА ДОЛЖНА БЫТЬ ПОЛНОЙ");
      return;
    }

    setQueueBusy(true);
    setQueueError("");

    try {
      const response = await request<JoinQueueResponse>("/queue/join", {
        method: "POST",
      });
      setQueueStatus((current) => ({
        ...current,
        state: response.state,
        opponent_user_id: response.opponent_user_id,
        search_duration_sec: 0,
      }));
      setGameModeOpen(false);
      await loadQueueStatus();
    } catch (err) {
      setQueueError(err instanceof Error ? err.message : "НЕ УДАЛОСЬ НАЧАТЬ ПОИСК");
    } finally {
      setQueueBusy(false);
    }
  }

  async function leaveQueue() {
    setQueueBusy(true);

    try {
      const response = await request("/queue/leave", {
        method: "POST",
      });
      void response;
      setQueueStatus({ state: "idle", search_duration_sec: 0 });
      setQueueError("");
    } catch (err) {
      setQueueError(err instanceof Error ? err.message : "НЕ УДАЛОСЬ ОТМЕНИТЬ ПОИСК");
    } finally {
      setQueueBusy(false);
    }
  }

  return (
    <main className="app-shell">
      {screen === "menu" ? (
        <MainMenu
          me={me}
          selectedHero={selectedHero}
          heroes={heroes}
          heroPickerOpen={heroPickerOpen}
          setHeroPickerOpen={setHeroPickerOpen}
          chooseHero={chooseHero}
          onStartMatch={() => {
            setQueueError("");
            setGameModeOpen(true);
          }}
          inventoryHidden={queueSearching}
          startMatchDisabled={queueSearching}
          onInventory={() => setScreen("inventory")}
        />
      ) : (
        <InventoryScreen
          draftDeckEntries={draftDeckEntries}
          savedDeckEntries={deckEntries}
          battleCards={battleCards}
          buffCards={buffCards}
          onBack={() => setScreen("menu")}
          onDraftDeckChange={setDraftDeckEntries}
          onDeckSaved={setDeckEntries}
        />
      )}

      <section className="card surface dev-panel">
        <div className="section-head">
          <div>
            <p className="eyebrow">Dev Auth</p>
            <h2>Быстрый вход</h2>
          </div>
          <span className={`status-pill ${me ? "status-pill--ok" : "status-pill--idle"}`}>
            {busy ? "Loading" : me ? "Authorized" : "Guest"}
          </span>
        </div>

        <label className="field">
          <span>Username</span>
          <input
            value={username}
            onChange={(event) => setUsername(event.target.value)}
            placeholder="dev"
            autoComplete="off"
          />
        </label>

        <label className="field">
          <span>User ID</span>
          <input
            value={userId}
            onChange={(event) => setUserId(event.target.value)}
            placeholder="4"
            inputMode="numeric"
            autoComplete="off"
          />
        </label>

        <div className="quick-row">
          {QUICK_USERS.map((item) => (
            <button key={item} type="button" className="chip" onClick={() => login(item)} disabled={busy}>
              {item}
            </button>
          ))}
        </div>

        <button type="button" className="primary-button" onClick={() => login(username)} disabled={busy}>
          {busy ? "Подключаем..." : "Войти как dev user"}
        </button>

        <button type="button" className="secondary-button" onClick={loginAsUserId} disabled={busy}>
          {busy ? "Подключаем..." : "Войти по user ID"}
        </button>

        {error ? <p className="error-text">{error}</p> : null}
      </section>

      <GameModePanel
        open={gameModeOpen}
        searching={queueSearching}
        queueState={queueStatus.state}
        busy={queueBusy}
        error={queueError}
        searchDurationSec={queueStatus.search_duration_sec ?? 0}
        canQueue={canJoinQueue}
        onClose={() => {
          setGameModeOpen(false);
          setQueueError("");
        }}
        onFindMatch={joinQueue}
        onCancelSearch={leaveQueue}
      />
    </main>
  );
}
