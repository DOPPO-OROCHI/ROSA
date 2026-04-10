import { useEffect, useState } from "react";
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
  MeResponse,
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
  const [busy, setBusy] = useState(true);
  const [error, setError] = useState("");
  const [selectedHeroCode, setSelectedHeroCode] = useState("");
  const [heroPickerOpen, setHeroPickerOpen] = useState(false);

  const selectedHero = heroes.find((hero) => hero.hero_code === selectedHeroCode) ?? heroes[0] ?? null;

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
  }

  useEffect(() => {
    loadSession()
      .then(() => loadInventory())
      .catch(() => {
        setMe(null);
        setHeroes([]);
        setBattleCards([]);
        setBuffCards([]);
        setDeckEntries([]);
        setSelectedHeroCode("");
      })
      .finally(() => setBusy(false));
  }, []);

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
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setBusy(false);
    }
  }

  async function chooseHero(hero: Hero) {
    setSelectedHeroCode(hero.hero_code);
    try {
      await request("/heroes/select", {
        method: "POST",
        body: JSON.stringify({ hero_code: hero.hero_code }),
      });
      const nextMe = await request<MeResponse>("/me");
      setMe(nextMe);
      setHeroPickerOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to select hero");
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
          onInventory={() => setScreen("inventory")}
        />
      ) : (
        <InventoryScreen
          deckEntries={deckEntries}
          battleCards={battleCards}
          buffCards={buffCards}
          onBack={() => setScreen("menu")}
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
    </main>
  );
}
