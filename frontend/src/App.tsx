import { useState, useEffect } from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import { fetchRates } from "./Api/currencyApi";
import { fetchFavorites } from "./Api/favoritesApi";
import type { FavoritePair } from "./Api/favoritesApi";
import { apiFetch } from "./Api/apiClient";
import Header from "./components/Header";
import AuthPage from "./pages/AuthPage";
import VerifyEmailPage from "./pages/VerifyEmailPage";
import NotificationManagerPage from "./pages/NotificationManagerPage";
import CreateNotificationPage from "./pages/CreateNotificationPage";
import CurrencyConverter from "./components/CurrencyConverter";
import FavoritesList from "./components/FavoritesList";
import CurrencyDetailPage from "./pages/CurrencyDetailPage";


// ИСПРАВЛЕНО: загружаем ПОДПИСКИ (subscriptions), а не уведомления!
async function fetchSubscriptions() {
  const token = localStorage.getItem("token");
  const response = await apiFetch("/api/subscriptions", {
    headers: {
      "Authorization": `Bearer ${token}`,
    },
  });
  if (!response.ok) throw new Error("Failed to fetch subscriptions");
  return response.json();
}

export default function App() {
  const [from, setFrom] = useState("USD");
  const [to, setTo] = useState("RUB");
  const [user, setUser] = useState<any>(() => {
    const saved = localStorage.getItem("user");
    return saved ? JSON.parse(saved) : null;
  });
  const [favorites, setFavorites] = useState<FavoritePair[]>([]);
  // ИСПРАВЛЕНО: теперь это список подписок (алёртов), а не уведомлений
  const [notifications, setNotifications] = useState<any[]>([]);
  const [rates, setRates] = useState<any>({ pair: "USD_RUB", rate: 73.3591 });

  useEffect(() => {
    if (user) {
      localStorage.setItem("user", JSON.stringify(user));
    } else {
      localStorage.removeItem("user");
    }
  }, [user]);

  // Загрузка favorites и subscriptions с бэкенда при логине
  useEffect(() => {
    if (user) {
      fetchFavorites()
        .then((data) => {
          const pairs: FavoritePair[] = data.favorite_pairs.map((pair: string) => {
            const [from, to] = pair.split("_");
            return { from, to };
          });
          setFavorites(pairs);
        })
        .catch((err) => {
          console.error("Failed to load favorites:", err);
          setFavorites([]);
        });

      // ИСПРАВЛЕНО: Загружаем ПОДПИСКИ (subscriptions), а не notifications
      fetchSubscriptions()
        .then((data) => {
          // Маппим подписки в формат, который ожидает NotificationManagerPage
          const subs = (data.subscriptions || []).map((s: any) => ({
            id: s.id,
            from: s.currency_code,
            to: "RUB", // Бэкенд хранит курс к RUB
            condition: s.condition,
            value: s.rate_value,
            is_active: s.is_active
          }));
          setNotifications(subs);
        })
        .catch((err) => {
          console.error("Failed to load subscriptions:", err);
          setNotifications([]);
        });
    } else {
      setFavorites([]);
      setNotifications([]);
    }
  }, [user]);

  useEffect(() => {
    console.log(`=== НАЧАЛО ЗАПРОСА ДЛЯ ПАРЫ: ${from}_${to} ===`);
    fetchRates()
      .then((data: any) => {
        console.log("Данные, которые пришли в App.tsx из API:", data);
        setRates(data);
      })
      .catch((err: any) => {
        console.error("Ошибка при получении курсов", err);
      });
  }, []);

  return (
    <Router>
      <div style={{ maxWidth: "1200px", margin: "0 auto", padding: "20px", fontFamily: "sans-serif" }}>
        <Header user={user} />
        <Routes>
          <Route
            path="/"
            element={
              <div>
                <CurrencyConverter
                  rates={rates}
                  user={user}
                  favorites={favorites}
                  setFavorites={setFavorites}
                  from={from}
                  setFrom={setFrom}
                  to={to}
                  setTo={setTo}
                />
                {user && (
                  <FavoritesList
                    favorites={favorites}
                    setFavorites={setFavorites}
                    setFrom={setFrom}
                    setTo={setTo}
                  />
                )}
              </div>
            }
          />
          <Route path="/auth" element={<AuthPage setUser={setUser} />} />
          <Route path="/verify-email" element={<VerifyEmailPage setUser={setUser} />} />
          <Route
            path="/profile"
            element={
              <NotificationManagerPage
                user={user}
                setUser={setUser}
                notifications={notifications}
                setNotifications={setNotifications}
                setFavorites={setFavorites}
              />
            }
          />
          <Route
            path="/create-notification"
            element={
              <CreateNotificationPage
                user={user}
                notifications={notifications}
                setNotifications={setNotifications}
              />
            }
          />
          <Route path="/currency/:pair" element={<CurrencyDetailPage rates={rates} />} />
        </Routes>
      </div>
    </Router>
  );
}