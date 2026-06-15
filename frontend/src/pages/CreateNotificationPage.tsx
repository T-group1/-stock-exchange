import { useState } from "react";
import { useNavigate } from "react-router-dom";

export default function CreateNotificationPage({ user, notifications, setNotifications, rates }: any) {
  const navigate = useNavigate();

  const [from, setFrom] = useState("USD");
  const [to, setTo] = useState("RUB");
  const [condition, setCondition] = useState("above");
  const [value, setValue] = useState("");

  const [hoverSave, setHoverSave] = useState(false);
  const [hoverBack, setHoverBack] = useState(false);
  const [error, setError] = useState("");

  const currencies = ["USD", "RUB", "EUR", "CNY", "GBP", "JPY", "AUD", "CAD", "SGD", "CHF"];

  if (!user) {
    return (
      <div style={{ padding: "30px 20px", display: "flex", justifyContent: "center" }}>
        <div style={{ maxWidth: "400px", width: "100%", background: "#f5f3ff", border: "1px solid #c084fc", borderRadius: "12px", padding: "25px", textAlign: "center" }}>
          <h3 style={{ margin: "0 0 15px 0", color: "#4b5563", fontSize: "18px" }}>Вы не вошли в систему 🔒</h3>
          <p style={{ color: "#6b7280", fontSize: "14px", margin: "0 0 20px 0" }}>Для создания уведомлений нужно авторизоваться.</p>
          <button onClick={() => navigate("/auth")} style={{ padding: "10px 24px", background: "#7c3aed", color: "#fff", border: "none", borderRadius: "8px", cursor: "pointer", fontWeight: "600" }}>Войти в аккаунт</button>
        </div>
      </div>
    );
  }

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    if (!value || Number(value) <= 0) {
      setError("Пожалуйста, введите корректное целевое значение курса");
      return;
    }

    const newNotification = {
      id: Date.now(),
      from,
      to,
      condition,
      value: Number(value)
    };

    setNotifications([...notifications, newNotification]);
    navigate("/profile"); 
  };

  return (
    <div style={{ 
      maxWidth: "520px", 
      margin: "40px auto", 
      padding: "30px", 
      background: "#f5f3ff", 
      borderRadius: "16px", 
      border: "1px solid #ddd6fe",
      boxShadow: "0 4px 10px rgba(124, 58, 237, 0.05)",
      fontFamily: "sans-serif"
    }}>
      <h2 style={{ margin: "0 0 8px 0", textAlign: "center", color: "#1e293b", fontSize: "24px" }}>
        Создать уведомление
      </h2>
      <p style={{ margin: "0 0 25px 0", textAlign: "center", color: "#64748b", fontSize: "15px" }}>
        Мы сообщим вам, когда курс достигнет указанной отметки
      </p>

      {error && (
        <div style={{ background: "#fee2e2", color: "#ef4444", padding: "10px", borderRadius: "8px", fontSize: "14px", marginBottom: "15px", border: "1px solid #fca5a5", textAlign: "center" }}>
          {error}
        </div>
      )}

      <form onSubmit={handleCreate} style={{ display: "flex", flexDirection: "column", gap: "22px" }}>

        <div style={{ display: "flex", gap: "15px", alignItems: "center", background: "#fff", padding: "18px", borderRadius: "12px", border: "1px solid #e2e8f0" }}>
          <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: "6px" }}>
            <span style={{ fontSize: "12px", color: "#94a3b8", fontWeight: "700" }}>ИЗ ВАЛЮТЫ</span>
            <select 
              value={from} 
              onChange={(e) => setFrom(e.target.value)} 
              style={{ padding: "10px", borderRadius: "8px", border: "1px solid #cbd5e1", fontSize: "16px", fontWeight: "600", color: "#1e293b", outline: "none", cursor: "pointer" }}
            >
              {currencies.map(c => <option key={c} value={c}>{c}</option>)}
            </select>
          </div>

          <span style={{ fontWeight: "bold", color: "#cbd5e1", marginTop: "20px", fontSize: "18px" }}>➔</span>

          <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: "6px" }}>
            <span style={{ fontSize: "12px", color: "#94a3b8", fontWeight: "700" }}>В ВАЛЮТУ</span>
            <select 
              value={to} 
              onChange={(e) => setTo(e.target.value)} 
              style={{ padding: "10px", borderRadius: "8px", border: "1px solid #cbd5e1", fontSize: "16px", fontWeight: "600", color: "#1e293b", outline: "none", cursor: "pointer" }}
            >
              {currencies.map(c => <option key={c} value={c}>{c}</option>)}
            </select>
          </div>
        </div>

        <div style={{
          background: "#fff",
          padding: "12px 18px",
          borderRadius: "10px",
          border: "1px dashed #ddd6fe",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center"
        }}>
          <span style={{ fontSize: "14px", color: "#64748b", fontWeight: "500" }}>Текущий рыночный курс:</span>
          <span style={{ fontSize: "16px", color: "#7c3aed", fontWeight: "700" }}>
            {rates && rates[from] && rates[to] ? (
              `1 ${from} = ${(rates[to] / rates[from]).toFixed(4)} ${to}`
            ) : (
              "Загрузка курсов..."
            )}
          </span>
        </div>

       <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
          <label style={{ fontSize: "15px", color: "#475569", fontWeight: "600" }}>
            Критерий:
          </label>

          <div style={{ display: "flex", gap: "12px", width: "100%" }}>
            
            <button
              type="button"
              onClick={() => setCondition("above")}
              style={{
                flex: 1, 
                borderRadius: "10px",
                fontSize: "15px",
                fontWeight: "600",
                cursor: "pointer",
                transition: "all 0.2s ease",
                background: condition === "above" ? "#22c55e" : "#fff",
                color: condition === "above" ? "#fff" : "#64748b",
                border: condition === "above" ? "2px solid #22c55e" : "2px solid #cbd5e1",
                boxShadow: condition === "above" ? "0 4px 12px rgba(34, 197, 94, 0.2)" : "none",
                transform: condition === "above" ? "scale(1.02)" : "scale(1)"
              }}
            >
                Повысится до...
            </button>

            <button
              type="button"
              onClick={() => setCondition("below")}
              style={{
                flex: 1,
                padding: "14px",
                borderRadius: "10px",
                fontSize: "15px",
                fontWeight: "600",
                cursor: "pointer",
                transition: "all 0.2s ease",
                background: condition === "below" ? "#ef4444" : "#fff",
                color: condition === "below" ? "#fff" : "#64748b",
                border: condition === "below" ? "2px solid #ef4444" : "2px solid #cbd5e1",
                boxShadow: condition === "below" ? "0 4px 12px rgba(239, 68, 68, 0.2)" : "none",
                transform: condition === "below" ? "scale(1.02)" : "scale(1)"
              }}
            >
              Понизится до...
            </button>

          </div>
        </div>

        <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
          <label style={{ fontSize: "15px", color: "#475569", fontWeight: "600" }}>Целевой курс валюты:</label>
          <input 
            type="number" 
            step="0.01"
            required
            value={value} 
            onChange={(e) => { setError(""); setValue(e.target.value); }} 
            placeholder="0.00"
            style={{ padding: "12px", borderRadius: "8px", border: "1px solid #cbd5e1", fontSize: "16px", fontWeight: "500", color: "#1e293b", outline: "none" }}
          />
        </div>

        <div style={{ display: "flex", gap: "12px", marginTop: "10px" }}>
          <button 
            type="button"
            onClick={() => navigate("/profile")}
            onMouseEnter={() => setHoverBack(true)}
            onMouseLeave={() => setHoverBack(false)}
            style={{ 
              flex: 1, padding: "14px", background: hoverBack ? "#e2e8f0" : "#fff", color: "#475569", 
              border: "1px solid #cbd5e1", borderRadius: "10px", cursor: "pointer", fontWeight: "600", fontSize: "15px", transition: "all 0.2s" 
            }}
          >
            Отмена
          </button>
          
          <button 
            type="submit"
            onMouseEnter={() => setHoverSave(true)}
            onMouseLeave={() => setHoverSave(false)}
            style={{ 
              flex: 1, padding: "14px", background: hoverSave ? "#6d28d9" : "#7c3aed", color: "#fff", 
              border: "none", borderRadius: "10px", cursor: "pointer", fontWeight: "600", fontSize: "15px", transition: "all 0.2s",
              boxShadow: "0 4px 6px rgba(124, 58, 237, 0.15)"
            }}
          >
            Создать уведомление
          </button>
        </div>

      </form>
    </div>
  );
}