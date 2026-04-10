import { useEffect, useState } from "react";

type MeResponse = {
  user_id: number;
  username: string;
  first_name: string;
  rating: number;
  xp: number;
  selected_hero_name?: string;
};

const QUICK_USERS = ["dev", "roman", "test", "rosa"];

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
    ...init,
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Request failed: ${response.status}`);
  }

  return response.json() as Promise<T>;
}

export function App() {
  const [username, setUsername] = useState("dev");
  const [me, setMe] = useState<MeResponse | null>(null);
  const [busy, setBusy] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    request<MeResponse>("/me")
      .then((next) => {
        setMe(next);
        setError("");
      })
      .catch(() => {
        setMe(null);
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
      const nextMe = await request<MeResponse>("/me");
      setMe(nextMe);
      setUsername(nextUsername);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="app-shell">
      <section className="hero-panel">
        <h1>PROJECT ROSE</h1>
      </section>

      <section className="card surface">
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

        {error ? <p className="error-text">{error}</p> : null}
      </section>

      <section className="card surface">
        <div className="section-head">
          <div>
            <p className="eyebrow">Session</p>
            <h2>Текущее состояние</h2>
          </div>
        </div>

        {me ? (
          <div className="profile-grid">
            <div className="metric">
              <span>ID</span>
              <strong>{me.user_id}</strong>
            </div>
            <div className="metric">
              <span>Username</span>
              <strong>{me.username}</strong>
            </div>
            <div className="metric">
              <span>Rating</span>
              <strong>{me.rating}</strong>
            </div>
            <div className="metric">
              <span>XP</span>
              <strong>{me.xp}</strong>
            </div>
          </div>
        ) : (
          <p className="empty-state">Сессии пока нет. Нажми кнопку входа выше.</p>
        )}
      </section>
    </main>
  );
}
