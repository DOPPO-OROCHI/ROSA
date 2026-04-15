import { useEffect, useState } from "react";
import { GameModePanel } from "./components/GameModePanel";
import { InventoryScreen } from "./components/InventoryScreen";
import { MainMenu } from "./components/MainMenu";
import { MatchFoundPanel } from "./components/MatchFoundPanel";
import { BattleScreen } from "./components/battle/BattleScreen";
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
import type { MaskedBattleMatchState } from "./components/battle/types";

const QUICK_USERS = ["dev", "roman", "test", "rosa"];

type Screen = "menu" | "inventory" | "battle";

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
  const [activeMatchId, setActiveMatchId] = useState<number | null>(null);
  const [matchFoundVerdict, setMatchFoundVerdict] = useState<"idle" | "accepted" | "declined_self" | "declined_opponent" | "countdown">("idle");
  const [matchFoundCountdown, setMatchFoundCountdown] = useState(3);
  const [acceptedByMeLatch, setAcceptedByMeLatch] = useState(false);

  const selectedHero = heroes.find((hero) => hero.hero_code === selectedHeroCode) ?? heroes[0] ?? null;
  const deckCardCount = deckEntries.reduce((total, entry) => total + entry.count, 0);
  const canJoinQueue = Boolean(selectedHeroCode) && deckCardCount === 20;
  const queueSearching = queueStatus.state === "searching" || queueStatus.state === "pending_match";
  const queueDeckCards = deckEntries
    .filter((entry) => entry.kind === "battle" && entry.count > 0)
    .map((entry) => {
      const card = battleCards.find((item) => item.template_id === entry.template_id);
      if (!card) {
        return null;
      }

      return {
        templateId: card.template_id,
        name: card.name,
        count: entry.count,
      };
    })
    .filter((entry): entry is { templateId: string; name: string; count: number } => entry !== null);

  async function loadQueueStatus() {
    const status = await request<QueueStatusResponse>("/queue/status");
    setQueueStatus(status);
    setQueueError("");
  }

  async function loadActiveMatch() {
    const matches = await request<MaskedBattleMatchState[]>("/matches");
    const active = matches.find((match) => !match.finished);
    if (active) {
      setActiveMatchId(active.match_id);
      if (queueStatus.state === "pending_match" || matchFoundVerdict === "countdown") {
        setMatchFoundVerdict("countdown");
        return active.match_id;
      }
      setScreen("battle");
      return active.match_id;
    }
    setActiveMatchId(null);
    return null;
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
        await loadActiveMatch();
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
        setActiveMatchId(null);
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
      void loadActiveMatch().catch(() => undefined);
    }, 1000);

    return () => window.clearInterval(id);
  }, [queueSearching]);

  useEffect(() => {
    if (queueStatus.state === "pending_match") {
      if (queueStatus.accepted_by_me) {
        setAcceptedByMeLatch(true);
        setMatchFoundVerdict("accepted");
      } else {
        setMatchFoundVerdict("idle");
      }
      return;
    }

    if (queueStatus.state === "searching" && acceptedByMeLatch) {
      setMatchFoundVerdict("declined_opponent");
      const id = window.setTimeout(() => {
        setMatchFoundVerdict("idle");
        setAcceptedByMeLatch(false);
      }, 3000);
      return () => window.clearTimeout(id);
    }

    if (queueStatus.state === "penalty") {
      const id = window.setTimeout(() => {
        setMatchFoundVerdict("idle");
        setAcceptedByMeLatch(false);
      }, 3000);
      return () => window.clearTimeout(id);
    }
  }, [acceptedByMeLatch, queueStatus.state, queueStatus.accepted_by_me]);

  useEffect(() => {
    if (matchFoundVerdict !== "countdown" || !activeMatchId) {
      setMatchFoundCountdown(3);
      return;
    }

    const id = window.setInterval(() => {
      setMatchFoundCountdown((current) => {
        if (current <= 1) {
          window.clearInterval(id);
          setScreen("battle");
          setMatchFoundVerdict("idle");
          setAcceptedByMeLatch(false);
          return 0;
        }
        return current - 1;
      });
    }, 1000);

    return () => window.clearInterval(id);
  }, [activeMatchId, matchFoundVerdict]);

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
      await loadActiveMatch();
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
      await loadActiveMatch();
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
      setQueueError("НЕТ ГЕРОЯ");
      return;
    }

    if (deckCardCount !== 20) {
      setQueueError("НЕПОЛНАЯ ДЕКА");
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
      await request("/queue/leave", {
        method: "POST",
      });
      setQueueStatus({ state: "idle", search_duration_sec: 0 });
      setQueueError("");
    } catch (err) {
      setQueueError(err instanceof Error ? err.message : "НЕ УДАЛОСЬ ОТМЕНИТЬ ПОИСК");
    } finally {
      setQueueBusy(false);
    }
  }

  async function acceptFoundMatch() {
    setQueueBusy(true);
    setQueueError("");
    try {
      const response = await request<QueueStatusResponse | MaskedBattleMatchState>("/queue/accept", {
        method: "POST",
      });
      if ("match_id" in response) {
        setActiveMatchId(response.match_id);
        setMatchFoundVerdict("countdown");
        return;
      }
      setQueueStatus((current) => ({
        ...current,
        state: response.state,
      }));
      setAcceptedByMeLatch(true);
      setMatchFoundVerdict("accepted");
      await loadQueueStatus();
    } catch (err) {
      setQueueError(err instanceof Error ? err.message : "НЕ УДАЛОСЬ ПРИНЯТЬ МАТЧ");
    } finally {
      setQueueBusy(false);
    }
  }

  async function declineFoundMatch() {
    setQueueBusy(true);
    setQueueError("");
    try {
      await request("/queue/decline", {
        method: "POST",
      });
      setMatchFoundVerdict("declined_self");
      setQueueStatus((current) => ({ ...current, state: "penalty" }));
      window.setTimeout(() => {
        setScreen("menu");
        setAcceptedByMeLatch(false);
      }, 3000);
    } catch (err) {
      setQueueError(err instanceof Error ? err.message : "НЕ УДАЛОСЬ ОТКАЗАТЬСЯ");
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
      ) : screen === "inventory" ? (
        <InventoryScreen
          draftDeckEntries={draftDeckEntries}
          savedDeckEntries={deckEntries}
          battleCards={battleCards}
          buffCards={buffCards}
          onBack={() => setScreen("menu")}
          onDraftDeckChange={setDraftDeckEntries}
          onDeckSaved={setDeckEntries}
        />
      ) : activeMatchId && me ? (
        <BattleScreen
          currentUserId={me.user_id}
          matchId={activeMatchId}
          onLeaveToMenu={() => {
            setActiveMatchId(null);
            setScreen("menu");
          }}
        />
      ) : null}

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
        selectedHero={selectedHero}
        deckCards={queueDeckCards}
        onClose={() => {
          setGameModeOpen(false);
          setQueueError("");
        }}
        onFindMatch={joinQueue}
        onCancelSearch={leaveQueue}
      />

      <MatchFoundPanel
        open={queueStatus.state === "pending_match" || matchFoundVerdict === "declined_self" || matchFoundVerdict === "declined_opponent" || matchFoundVerdict === "countdown"}
        busy={queueBusy}
        error={queueError}
        acceptedByMe={Boolean(queueStatus.accepted_by_me) || acceptedByMeLatch}
        acceptedByOpponent={Boolean(queueStatus.accepted_by_opponent)}
        deadlineSec={Math.max(0, Math.ceil(((queueStatus.accept_deadline_at ? Date.parse(queueStatus.accept_deadline_at) : 0) - Date.now()) / 1000))}
        verdict={matchFoundVerdict}
        countdownSec={matchFoundCountdown}
        playerRating={me?.rating}
        opponentRating={queueStatus.opponent_user_id ?? 0}
        onAccept={acceptFoundMatch}
        onDecline={declineFoundMatch}
      />
    </main>
  );
}
