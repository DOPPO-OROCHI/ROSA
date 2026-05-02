import { useEffect, useRef, useState } from "react";
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
  QueueState,
  QueueStatusResponse,
} from "./types";
import type { MaskedBattleMatchState } from "./components/battle/types";

const QUICK_USERS = ["dev", "roman", "test", "rosa"];
const MENU_MUSIC_VOLUME = 0.35;
const MENU_MUSIC_MATCH_FOUND_VOLUME = MENU_MUSIC_VOLUME * 0.5;
const BATTLE_MUSIC_VOLUME = 0.28;
const MUSIC_FADE_MS = 420;

type Screen = "menu" | "inventory" | "battle";

export function App() {
  const clickAudioRef = useRef<HTMLAudioElement | null>(null);
  const matchFoundAudioRef = useRef<HTMLAudioElement | null>(null);
  const menuMusicRef = useRef<HTMLAudioElement | null>(null);
  const battleMusicRef = useRef<HTMLAudioElement | null>(null);
  const menuMusicFadeFrameRef = useRef<number | null>(null);
  const battleMusicFadeFrameRef = useRef<number | null>(null);
  const battleMusicUnlockedRef = useRef(false);
  const screenRef = useRef<Screen>("menu");
  const musicEnabledRef = useRef(true);
  const prevQueueStateRef = useRef<QueueState>("idle");
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
  const [musicEnabled, setMusicEnabled] = useState(true);

  const selectedHero = heroes.find((hero) => hero.hero_code === selectedHeroCode) ?? null;
  const deckCardCount = draftDeckEntries.reduce((total, entry) => total + entry.count, 0);
  const canJoinQueue = Boolean(selectedHeroCode) && deckCardCount === 20;
  const queueHint = !selectedHeroCode && deckCardCount !== 20
    ? "НЕТ ГЕРОЯ | НЕПОЛНАЯ ДЕКА"
    : !selectedHeroCode
      ? "НЕТ ГЕРОЯ"
      : deckCardCount !== 20
        ? "НЕПОЛНАЯ ДЕКА"
        : "";
  const queueSearching = queueStatus.state === "searching" || queueStatus.state === "pending_match";
  const queuePenalty = queueStatus.state === "penalty";
  const matchTransitionLocked = Boolean(activeMatchId) && (acceptedByMeLatch || matchFoundVerdict === "countdown");
  const startMatchLabel = !selectedHeroCode && deckCardCount !== 20
    ? "НЕТ ГЕРОЯ / НЕПОЛНАЯ ДЕКА"
    : !selectedHeroCode
      ? "НЕ ВЫБРАН ПЕРСОНАЖ"
      : deckCardCount !== 20
        ? "НЕПОЛНАЯ ДЕКА"
        : "START MATCH";
  function fadeAudioTo(
    audio: HTMLAudioElement | null,
    frameRef: React.MutableRefObject<number | null>,
    targetVolume: number,
    durationMs = MUSIC_FADE_MS,
  ) {
    if (!audio) {
      return;
    }

    if (frameRef.current !== null) {
      window.cancelAnimationFrame(frameRef.current);
      frameRef.current = null;
    }

    const startVolume = audio.volume;
    if (Math.abs(startVolume - targetVolume) < 0.005) {
      audio.volume = targetVolume;
      return;
    }

    const startedAt = performance.now();
    const tick = (now: number) => {
      const progress = Math.min((now - startedAt) / durationMs, 1);
      const eased = 1 - Math.pow(1 - progress, 3);
      audio.volume = startVolume + (targetVolume - startVolume) * eased;

      if (progress < 1) {
        frameRef.current = window.requestAnimationFrame(tick);
        return;
      }

      audio.volume = targetVolume;
      frameRef.current = null;
    };

    frameRef.current = window.requestAnimationFrame(tick);
  }

  function getMenuMusicTargetVolume() {
    return queueStatus.state === "pending_match"
      ? MENU_MUSIC_MATCH_FOUND_VOLUME
      : MENU_MUSIC_VOLUME;
  }

  function fadeMenuMusicTo(targetVolume: number, durationMs = MUSIC_FADE_MS) {
    fadeAudioTo(menuMusicRef.current, menuMusicFadeFrameRef, targetVolume, durationMs);
  }

  function fadeBattleMusicTo(targetVolume: number, durationMs = MUSIC_FADE_MS) {
    fadeAudioTo(battleMusicRef.current, battleMusicFadeFrameRef, targetVolume, durationMs);
  }

  function syncMusicPlayback(nextScreen: Screen, nextMusicEnabled: boolean) {
    const menuAudio = menuMusicRef.current;
    const battleAudio = battleMusicRef.current;
    if (!menuAudio || !battleAudio) {
      return;
    }

    if (!nextMusicEnabled) {
      menuAudio.pause();
      battleAudio.pause();
      menuAudio.volume = getMenuMusicTargetVolume();
      battleAudio.volume = BATTLE_MUSIC_VOLUME;
      return;
    }

    if (nextScreen === "menu" || nextScreen === "inventory") {
      if (menuAudio.paused) {
        void menuAudio.play().catch(() => undefined);
      }
      if (battleMusicUnlockedRef.current && battleAudio.paused) {
        battleAudio.volume = 0;
        void battleAudio.play().catch(() => undefined);
      }
      fadeBattleMusicTo(0);
      fadeMenuMusicTo(getMenuMusicTargetVolume());
      return;
    }

    if (nextScreen === "battle") {
      if (battleAudio.paused) {
        battleAudio.volume = 0;
        void battleAudio.play().catch(() => undefined);
      }
      fadeMenuMusicTo(0);
      fadeBattleMusicTo(BATTLE_MUSIC_VOLUME);
    }
  }

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
      if (queueStatus.state === "pending_match" || acceptedByMeLatch || matchFoundVerdict === "countdown") {
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
    setSelectedHeroCode(nextMe.selected_hero_code || "");
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
    screenRef.current = screen;
  }, [screen]);

  useEffect(() => {
    musicEnabledRef.current = musicEnabled;
  }, [musicEnabled]);

  useEffect(() => {
    const audio = new Audio("/assets/ui/sounds/ui/click.mp3");
    audio.preload = "auto";
    clickAudioRef.current = audio;

    function handleUiClick(event: PointerEvent) {
      const target = event.target;
      if (!(target instanceof Element)) {
        return;
      }

      const interactive = target.closest(
        'button, a, input, select, textarea, summary, [role="button"], [data-ui-click]',
      );

      if (!(interactive instanceof HTMLElement)) {
        return;
      }

      if (
        interactive.matches(":disabled") ||
        interactive.getAttribute("aria-disabled") === "true"
      ) {
        return;
      }

      const baseAudio = clickAudioRef.current;
      if (!baseAudio) {
        return;
      }

      const clickAudio = baseAudio.cloneNode() as HTMLAudioElement;
      clickAudio.volume = 0.1;
      void clickAudio.play().catch(() => undefined);

      const menuAudio = menuMusicRef.current;
      const battleAudio = battleMusicRef.current;

      if (!musicEnabledRef.current) {
        return;
      }

      if (
        menuAudio &&
        (screenRef.current === "menu" || screenRef.current === "inventory") &&
        menuAudio.paused
      ) {
        void menuAudio.play().catch(() => undefined);
      }

      if (battleAudio && screenRef.current === "battle" && battleAudio.paused) {
        void battleAudio.play().catch(() => undefined);
      }

      if (battleAudio && !battleMusicUnlockedRef.current) {
        const previousVolume = battleAudio.volume;
        battleAudio.volume = 0;
        void battleAudio.play()
          .then(() => {
            battleAudio.pause();
            battleAudio.currentTime = 0;
            battleAudio.volume = previousVolume;
            battleMusicUnlockedRef.current = true;
          })
          .catch(() => {
            battleAudio.volume = previousVolume;
          });
      }
    }

    window.addEventListener("pointerdown", handleUiClick, { passive: true });
    return () => window.removeEventListener("pointerdown", handleUiClick);
  }, []);

  useEffect(() => {
    const audio = new Audio("/assets/ui/sounds/ui/match_found.mp3");
    audio.preload = "auto";
    audio.volume = 0.62;
    matchFoundAudioRef.current = audio;

    return () => {
      audio.pause();
      audio.currentTime = 0;
      if (matchFoundAudioRef.current === audio) {
        matchFoundAudioRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    const audio = new Audio("/assets/ui/sounds/music/menu_music.mp3");
    audio.preload = "auto";
    audio.loop = true;
    audio.volume = MENU_MUSIC_VOLUME;
    menuMusicRef.current = audio;
    syncMusicPlayback(screenRef.current, musicEnabledRef.current);

    return () => {
      if (menuMusicFadeFrameRef.current !== null) {
        window.cancelAnimationFrame(menuMusicFadeFrameRef.current);
        menuMusicFadeFrameRef.current = null;
      }
      audio.pause();
      audio.currentTime = 0;
      if (menuMusicRef.current === audio) {
        menuMusicRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    const audio = new Audio("/assets/ui/sounds/music/battle.mp3");
    audio.preload = "auto";
    audio.loop = true;
    audio.volume = 0.28;
    battleMusicRef.current = audio;
    syncMusicPlayback(screenRef.current, musicEnabledRef.current);

    return () => {
      audio.pause();
      audio.currentTime = 0;
      battleMusicUnlockedRef.current = false;
      if (battleMusicRef.current === audio) {
        battleMusicRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    syncMusicPlayback(screen, musicEnabled);
  }, [musicEnabled, screen]);

  useEffect(() => {
    if (!musicEnabled) {
      return;
    }

    fadeMenuMusicTo(
      queueStatus.state === "pending_match"
        ? MENU_MUSIC_MATCH_FOUND_VOLUME
        : MENU_MUSIC_VOLUME,
    );
  }, [musicEnabled, queueStatus.state]);

  useEffect(() => {
    const previousState = prevQueueStateRef.current;
    if (queueStatus.state === "pending_match" && previousState !== "pending_match") {
      const baseAudio = matchFoundAudioRef.current;
      if (baseAudio) {
        const matchFoundAudio = baseAudio.cloneNode() as HTMLAudioElement;
        matchFoundAudio.volume = baseAudio.volume;
        void matchFoundAudio.play().catch(() => undefined);
      }
    }
    prevQueueStateRef.current = queueStatus.state;
  }, [queueStatus.state]);

  useEffect(() => {
    if (!queueSearching && !queuePenalty) {
      return;
    }

    const id = window.setInterval(() => {
      void loadQueueStatus().catch(() => {
        setQueueStatus((current) => ({ ...current, state: "idle" }));
      });
      if (queueSearching) {
        void loadActiveMatch().catch(() => undefined);
      }
    }, 1000);

    return () => window.clearInterval(id);
  }, [queuePenalty, queueSearching]);

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
    if (!activeMatchId || !acceptedByMeLatch || matchFoundVerdict === "countdown") {
      return;
    }
    setMatchFoundVerdict("countdown");
  }, [acceptedByMeLatch, activeMatchId, matchFoundVerdict]);

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
            if (!canJoinQueue) {
              setQueueError(queueHint || "РќР•Р›Р¬Р—РЇ РќРђР§РђРўР¬ РџРћРРЎРљ");
              setGameModeOpen(true);
              return;
            }
            setQueueError("");
            setGameModeOpen(true);
          }}
          inventoryHidden={queueSearching}
          startMatchDisabled={queueSearching || !canJoinQueue}
          startMatchLabel={startMatchLabel}
          onInventory={() => setScreen("inventory")}
          musicEnabled={musicEnabled}
          onToggleMusic={() => {
            const nextEnabled = !musicEnabled;
            setMusicEnabled(nextEnabled);

            const menuAudio = menuMusicRef.current;
            const battleAudio = battleMusicRef.current;
            if (!menuAudio || !battleAudio) {
              return;
            }

            if (!nextEnabled) {
              menuAudio.pause();
              battleAudio.pause();
              return;
            }

            if (screen === "menu" || screen === "inventory") {
              void menuAudio.play().catch(() => undefined);
              return;
            }

            void battleAudio.play().catch(() => undefined);
          }}
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
      ) : activeMatchId && me && !matchTransitionLocked ? (
        <BattleScreen
          currentUserId={me.user_id}
          matchId={activeMatchId}
          heroes={heroes}
          deckEntries={deckEntries}
          onLeaveToMenu={() => {
            setActiveMatchId(null);
            setQueueStatus({ state: "idle", search_duration_sec: 0 });
            setQueueError("");
            setMatchFoundVerdict("idle");
            setMatchFoundCountdown(3);
            setAcceptedByMeLatch(false);
            setScreen("menu");
            void loadSession().catch(() => undefined);
            void loadQueueStatus().catch(() => undefined);
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
        penaltyUntil={queueStatus.penalty_until}
        busy={queueBusy}
        error={queueError}
        queueHint={queueHint}
        searchDurationSec={queueStatus.search_duration_sec ?? 0}
        canQueue={canJoinQueue}
        onClose={() => {
          setGameModeOpen(false);
          setQueueError("");
        }}
        onFindMatch={joinQueue}
        onCancelSearch={leaveQueue}
      />

      <MatchFoundPanel
        open={
          queueStatus.state === "pending_match" ||
          matchFoundVerdict === "declined_self" ||
          matchFoundVerdict === "declined_opponent" ||
          matchFoundVerdict === "countdown" ||
          (Boolean(activeMatchId) && acceptedByMeLatch)
        }
        busy={queueBusy}
        error={queueError}
        acceptedByMe={Boolean(queueStatus.accepted_by_me) || acceptedByMeLatch}
        acceptedByOpponent={Boolean(queueStatus.accepted_by_opponent)}
        deadlineSec={Math.max(0, Math.ceil(((queueStatus.accept_deadline_at ? Date.parse(queueStatus.accept_deadline_at) : 0) - Date.now()) / 1000))}
        verdict={matchFoundVerdict}
        countdownSec={matchFoundCountdown}
        playerRating={me?.rating}
        opponentRating={queueStatus.opponent_rating ?? 0}
        onAccept={acceptFoundMatch}
        onDecline={declineFoundMatch}
      />
    </main>
  );
}
