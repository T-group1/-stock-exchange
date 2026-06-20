import { useState } from "react";
import { useNavigate } from "react-router-dom";
import CurrencyChart from "./CurrencyChart";
import { addFavorite, removeFavorite } from "../Api/favoritesApi";

export default function CurrencyConverter({
  rates = {},
  user,
  favorites,
  setFavorites,
  from,
  setFrom,
  to,
  setTo
}: any) {
  const navigate = useNavigate();
  const [hoverFav, setHoverFav] = useState(false);
  const [amount, setAmount] = useState<number | "">(1);

  const currencies = rates ? Object.keys(rates).filter(key => key !== 'date' && key !== 'pair' && key !== 'rate') : ["USD", "RUB"]

  const fromRate = rates[from] || (rates.pair === `${from}_${to}` ? rates.rate : 1);
  const toRate = rates[to] || 1;

  const currentPairRate = rates.pair === `${from}_${to}` && rates.rate
    ? Number(rates.rate).toFixed(4)
    : (fromRate / toRate).toFixed(4);

  const result = amount === ""
    ? "0.00"
    : from === to
      ? Number(amount).toFixed(2)
      : rates.pair === `${from}_${to}` && rates.rate
        ? (Number(amount) * Number(rates.rate)).toFixed(2)
        : ((Number(amount) * fromRate) / toRate).toFixed(2);

  console.log("Конвертер видит rates:", rates);
  console.log("Финальный курс пары для рендера:", currentPairRate);

  const isFavorite = favorites?.some((f: any) => f.from === from && f.to === to) || false;

  const toggleFavorite = async () => {
    if (!user) {
      alert("Необходимо войти в аккаунт");
      navigate("/auth");
      return;
    }

    const currencyPair = `${from}_${to}`;

    if (isFavorite) {
      try {
        await removeFavorite(currencyPair);
        setFavorites(favorites.filter((f: any) => !(f.from === from && f.to === to)));
      } catch (err) {
        console.error("Failed to remove favorite:", err);
        alert("Ошибка при удалении из избранного");
      }
    } else {
      if (favorites.length >= 10) {
        alert("Достигнут лимит! Максимум 10 пар в избранном.");
        return;
      }
      try {
        await addFavorite(currencyPair);
        setFavorites([...favorites, { from, to }]);
      } catch (err) {
        console.error("Failed to add favorite:", err);
        alert("Ошибка при добавлении в избранное");
      }
    }
  };

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "25px", fontFamily: "sans-serif" }}>
      <div style={{
        background: "#fff",
        borderRadius: "16px",
        padding: "30px",
        boxShadow: "0 4px 20px rgba(0,0,0,0.08)",
        border: "1px solid #e8e8e8"
      }}>
        <div style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          marginBottom: "25px"
        }}>
          <h2 style={{ margin: 0, fontSize: "24px", color: "#333" }}>Конвертер валют</h2>
          {user && (
            <button
              onClick={toggleFavorite}
              onMouseEnter={() => setHoverFav(true)}
              onMouseLeave={() => setHoverFav(false)}
              style={{
                background: "none",
                border: "none",
                cursor: "pointer",
                fontSize: "32px",
                color: isFavorite ? "#ffd700" : (hoverFav ? "#ffd700" : "#ccc"),
                transition: "color 0.2s ease",
                padding: "5px"
              }}
              title={isFavorite ? "Удалить из избранного" : "Добавить в избранное"}
            >
              ★
            </button>
          )}
        </div>

        <div style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr",
          gap: "20px",
          marginBottom: "25px"
        }}>
          <div>
            <label style={{ display: "block", marginBottom: "8px", color: "#666", fontSize: "14px" }}>
              Из
            </label>
            <select
              value={from}
              onChange={(e) => setFrom(e.target.value)}
              style={{
                width: "100%",
                padding: "12px",
                fontSize: "16px",
                border: "2px solid #e0e0e0",
                borderRadius: "8px",
                background: "#f9f9f9"
              }}
            >
              {currencies.map((c: string) => (
                <option key={c} value={c}>{c}</option>
              ))}
            </select>
          </div>

          <div>
            <label style={{ display: "block", marginBottom: "8px", color: "#666", fontSize: "14px" }}>
              В
            </label>
            <select
              value={to}
              onChange={(e) => setTo(e.target.value)}
              style={{
                width: "100%",
                padding: "12px",
                fontSize: "16px",
                border: "2px solid #e0e0e0",
                borderRadius: "8px",
                background: "#f9f9f9"
              }}
            >
              {currencies.map((c: string) => (
                <option key={c} value={c}>{c}</option>
              ))}
            </select>
          </div>
        </div>

        <div style={{ marginBottom: "25px" }}>
          <label style={{ display: "block", marginBottom: "8px", color: "#666", fontSize: "14px" }}>
            Сумма
          </label>
          <input
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value === "" ? "" : Number(e.target.value))}
            style={{
              width: "100%",
              padding: "12px",
              fontSize: "16px",
              border: "2px solid #e0e0e0",
              borderRadius: "8px",
              background: "#f9f9f9",
              boxSizing: "border-box"
            }}
          />
        </div>

        <div style={{
          background: "#f5f7fa",
          padding: "20px",
          borderRadius: "12px",
          textAlign: "center"
        }}>
          <div style={{ fontSize: "14px", color: "#666", marginBottom: "8px" }}>
            Курс: 1 {from} = {currentPairRate} {to}
          </div>
          <div style={{ fontSize: "32px", fontWeight: "bold", color: "#333" }}>
            {result} {to}
          </div>
        </div>
      </div>

      <CurrencyChart rates={rates} from={from} to={to} />
    </div>
  );
}