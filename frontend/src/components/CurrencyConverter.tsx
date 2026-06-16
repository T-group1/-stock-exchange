import { useState } from "react";
import { useNavigate } from "react-router-dom";
import CurrencyChart from "./CurrencyChart";

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

  const currencies = ["USD", "RUB", "EUR", "CNY", "GBP", "JPY", "AUD", "CAD", "SGD", "CHF", "HKD"];

  const fromRate = rates[from] || 1;
  const toRate = rates[to] || 1;

  const result = amount === "" 
    ? "0.00" 
    : from === to 
      ? Number(amount).toFixed(2) 
      : ((Number(amount) * fromRate) / toRate).toFixed(2);

  const currentPairRate = (fromRate / toRate).toFixed(4);

  const isFavorite = favorites?.some((f: any) => f.from === from && f.to === to) || false;

  const toggleFavorite = () => {
    if (isFavorite) {
      setFavorites(favorites.filter((f: any) => !(f.from === from && f.to === to)));
    } else {
      if (favorites.length >= 10) {
        alert("Достигнут лимит! Максимум 10 пар в избранном.");
        return;
      }
      setFavorites([...favorites, { id: Date.now(), from, to }]);
    }
  };

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "25px", fontFamily: "sans-serif" }}>
      
      <div style={{ 
        display: "flex", 
        gap: "25px", 
        alignItems: "stretch", 
        flexWrap: "wrap" 
      }}>
        
        <div style={{ 
          flex: "2", 
          minWidth: "320px", 
          display: "flex", 
          flexDirection: "column", 
          gap: "16px" 
        }}>
          
          <div style={{ display: "flex", gap: "12px", width: "100%" }}>
            <input 
              type="number" 
              value={amount} 
              onChange={(e) => {
                const val = e.target.value;
                setAmount(val === "" ? "" : Number(val));
              }} 
              placeholder="0"
              style={{ 
                flex: "2", 
                padding: "16px 20px", 
                fontSize: "18px", 
                fontWeight: "600",
                border: "2px solid #ddd6fe", 
                borderRadius: "12px",
                outline: "none",
                color: "#1e293b"
              }}
            />
            <select 
              value={from} 
              onChange={(e) => setFrom(e.target.value)} 
              style={{ 
                flex: "1", 
                padding: "16px", 
                fontSize: "18px", 
                fontWeight: "700",
                border: "2px solid #ddd6fe",
                borderRadius: "12px",
                background: "#f8fafc",
                cursor: "pointer",
                outline: "none",
                color: "#1e293b"
              }}
            >
              {currencies.map(c => <option key={c} value={c}>{c}</option>)}
            </select>
          </div>

          <div style={{ display: "flex", gap: "12px", width: "100%" }}>
            <div style={{ 
              flex: "2", 
              padding: "16px 20px", 
              fontSize: "22px", 
              fontWeight: "800",
              border: "2px solid #e2e8f0",
              borderRadius: "12px",
              background: "#f8fafc",
              color: "#7c3aed",
              display: "flex",
              alignItems: "center"
            }}>
              {result}
            </div>
            <select 
              value={to} 
              onChange={(e) => setTo(e.target.value)} 
              style={{ 
                flex: "1", 
                padding: "16px", 
                fontSize: "18px", 
                fontWeight: "700",
                border: "2px solid #e2e8f0",
                borderRadius: "12px",
                background: "#f8fafc",
                cursor: "pointer",
                outline: "none",
                color: "#1e293b"
              }}
            >
              {currencies.map(c => <option key={c} value={c}>{c}</option>)}
            </select>
          </div>

        </div>

        <div style={{
          flex: "1",
          minWidth: "240px",
          background: "linear-gradient(135deg, #f5f3ff 0%, #ede9fe 100%)",
          border: "2px dashed #c084fc",
          borderRadius: "12px",
          padding: "20px",
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          alignItems: "center",
          textAlign: "center"
        }}>
          <span style={{ fontSize: "12px", color: "#7c3aed", fontWeight: "700", letterSpacing: "0.5px", textTransform: "uppercase" }}>
            Рыночный курс пары
          </span>
          <div style={{ fontSize: "32px", fontWeight: "800", color: "#5b21b6", margin: "6px 0" }}>
            {currentPairRate}
          </div>
          <span style={{ fontSize: "14px", color: "#64748b", fontWeight: "500" }}>
            1 {from} = {currentPairRate} {to}
          </span>
        </div>

      </div>

      <div style={{ width: "100%" }}>
        {user ? (
          <button 
            onClick={toggleFavorite} 
            onMouseEnter={() => setHoverFav(true)}
            onMouseLeave={() => setHoverFav(false)}
            style={{ 
              width: "100%", 
              padding: "14px", 
              cursor: "pointer", 
              background: isFavorite ? "#f5f3ff" : (hoverFav ? "#f3e8ff" : "#fff"),
              color: isFavorite ? "#7c3aed" : "#64748b",
              border: isFavorite ? "2px solid #7c3aed" : "2px solid #cbd5e1",
              borderRadius: "12px",
              fontWeight: "600",
              fontSize: "15px",
              transition: "all 0.2s ease",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              gap: "8px"
            }}
          >
            {isFavorite ? "⭐ Пара добавлена в избранное" : "☆ Добавить эту пару в Избранное"}
          </button>
        ) : (
          <div style={{ 
            padding: "20px", 
            background: "#f5f3ff", 
            borderRadius: "12px", 
            border: "1px solid #ddd6fe", 
            textAlign: "center" 
          }}>
            <p style={{ margin: "0 0 12px 0", fontSize: "14px", color: "#6d28d9", fontWeight: "500" }}>
              Чтобы добавлять валютные пары в избранное, пожалуйста, войдите в систему
            </p>
            <button 
              onClick={() => navigate("/auth")} 
              style={{ 
                padding: "10px 24px", 
                background: "#7c3aed", 
                color: "#fff", 
                border: "none", 
                borderRadius: "8px", 
                cursor: "pointer",
                fontSize: "14px",
                fontWeight: "600",
                boxShadow: "0 2px 4px rgba(124, 58, 237, 0.2)",
                transition: "all 0.2s"
              }}
              onMouseEnter={(e) => e.currentTarget.style.background = "#6d28d9"}
              onMouseLeave={(e) => e.currentTarget.style.background = "#7c3aed"}
            >
              Войти в личный кабинет
            </button>
          </div>
        )}
      </div>

      <div style={{ background: "#fff", padding: "20px 0 0 0", borderTop: "1px solid #f1f5f9" }}>
        <CurrencyChart baseCurrency={from} quoteCurrency={to} />  
      </div>

      <div style={{ marginTop: "15px" }}>
        <button
          disabled={!user} 
          onClick={() => navigate("/create-notification", { state: { from, to, openNotificationModal: true } })}
          style={{
            width: "100%",
            padding: "12px",
            background: user ? "#7c3aed" : "#cbd5e1",
            color: user ? "#fff" : "#94a3b8",
            border: "none",
            borderRadius: "12px",
            cursor: user ? "pointer" : "not-allowed",
            fontSize: "14px",
            fontWeight: "600",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            gap: "6px",
            transition: "all 0.2s"
          }}
          onMouseEnter={(e) => {
            if (user) e.currentTarget.style.background = "#5b21b6";
          }}
          onMouseLeave={(e) => {
            if (user) e.currentTarget.style.background = "#7c3aed";
          }}
        >
          {user ? "Создать уведомление об этой паре" : "Войдите, чтобы настроить уведомления"}
        </button>
      </div>

    </div>
  );
}