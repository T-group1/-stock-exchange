import { useState, useEffect } from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";

import { fetchRates } from "./Api/currencyApi";

import Header from "./components/Header";
import AuthPage from "./pages/AuthPage";
import VerifyEmailPage from "./pages/VerifyEmailPage";
import NotificationManagerPage from "./pages/NotificationManagerPage";
import CreateNotificationPage from "./pages/CreateNotificationPage";
import CurrencyConverter from "./components/CurrencyConverter";
import FavoritesList from "./components/FavoritesList";
import CurrencyDetailPage from "./pages/CurrencyDetailPage";

export default function App() {
  const [from, setFrom] = useState("USD");
  const [to, setTo] = useState("RUB");
  const [user, setUser] = useState<any>(null);
  const [favorites, setFavorites] = useState<any[]>([]);
  const [notifications, setNotifications] = useState<any[]>([]);
  const [rates, setRates] = useState<any>({pair: "USD_RUB", rate: 73.3591 }); 

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
          <Route path="/" element={
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
          } />

          <Route path="/auth" element={<AuthPage setUser={setUser} />} />
          
          <Route path="/verify-email" element={<VerifyEmailPage setUser={setUser} />} />
          
          <Route path="/profile" element={
            <NotificationManagerPage 
              user={user} 
              setUser={setUser} 
              notifications={notifications} 
              setNotifications={setNotifications} 
              setFavorites={setFavorites}
            />
          } />

          <Route path="/create-notification" element={
            <CreateNotificationPage 
              user={user} 
              notifications={notifications} 
              setNotifications={setNotifications} 
              rates={rates} />           
              } />
          <Route path="/currency/:pair" element={<CurrencyDetailPage rates={rates} />} />
        </Routes>
      </div>
    </Router>
  );
}