import { useState } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { apiFetch } from "../Api/apiClient";

export default function CreateNotificationPage({
  user,
  notifications,
  setNotifications,
  rates // ДОБАВИЛИ rates
}: any) {
  const navigate = useNavigate();
  const location = useLocation(); // Достаем данные от роутера

  // Берем валюты из кнопки конвертера, если они есть. Иначе ставим USD и RUB
  const [from, setFrom] = useState(location.state?.from || "USD");
  const [to, setTo] = useState(location.state?.to || "RUB");

  const [condition, setCondition] = useState("above");
  const [value, setValue] = useState("");
  const [error, setError] = useState("");
  const [hoverBack, setHoverBack] = useState(false);
  const [hoverSave, setHoverSave] = useState(false);
  const [loading, setLoading] = useState(false);

  // Рассчитываем текущий курс
  const fromRate = rates?.[from] || 1;
  const toRate = rates?.[to] || 1;
  const currentRate = rates?.pair === `${from}_${to}` && rates?.rate
    ? Number(rates.rate).toFixed(4)
    : (fromRate / toRate).toFixed(4);

  // Генерируем полный список доступных валют
  const availableCurrencies = rates ? Object.keys(rates).filter(key => key !== 'date' && key !== 'pair' && key !== 'rate') : ["USD", "EUR", "RUB", "CNY", "GBP", "JPY"];

  if (!user) {
    return (
      <div style={{ padding: "30px 20px", textAlign: "center" }}>
        <h3>Вы не вошли в систему</h3>
        <button
          onClick={() => navigate("/auth")}
          style={{ padding: "10px 24px", background: "#7c3aed", color: "#fff", border: "none", borderRadius: "8px", cursor: "pointer" }}
        >
          Войти
        </button>
      </div>
    );
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!value || Number(value) <= 0) {
      setError("Пожалуйста, введите корректное целевое значение курса");
      return;
    }

    setLoading(true);
    setError("");

    try {
      const subscriptionData = {
        currency_code: from,
        rate_value: Number(value),
        condition: condition
      };

      const response = await apiFetch("/api/subscriptions", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify(subscriptionData)
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || "Failed to create subscription");
      }

      const newSubscription = await response.json();

      const newNotification = {
        id: newSubscription.id || Date.now(),
        from,
        to,
        condition,
        value: Number(value)
      };

      setNotifications([...notifications, newNotification]);
      navigate("/profile");
    } catch (err: any) {
      setError(err.message || "Не удалось создать уведомление. Попробуйте снова.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ maxWidth: "600px", margin: "0 auto", padding: "20px" }}>
      <h2 style={{ marginBottom: "30px", color: "#0f172a" }}>Создание нового уведомления</h2>

      <div style={{ background: "#fff", padding: "30px", borderRadius: "12px", border: "1px solid #e2e8f0" }}>
        <div style={{ marginBottom: "25px" }}>
          <label style={{ display: "block", marginBottom: "8px", fontWeight: "600", color: "#475569" }}>
            Из валюты:
          </label>
          <select
            value={from}
            onChange={(e) => setFrom(e.target.value)}
            style={{ width: "100%", padding: "12px", borderRadius: "8px", border: "1px solid #cbd5e1", fontSize: "16px", outline: "none" }}
          >
            {availableCurrencies.map((currency) => (
              <option key={`from-${currency}`} value={currency}>{currency}</option>
            ))}
          </select>
        </div>

        <div style={{ marginBottom: "25px" }}>
          <label style={{ display: "block", marginBottom: "8px", fontWeight: "600", color: "#475569" }}>
            В валюту:
          </label>
          <select
            value={to}
            onChange={(e) => setTo(e.target.value)}
            style={{ width: "100%", padding: "12px", borderRadius: "8px", border: "1px solid #cbd5e1", fontSize: "16px", outline: "none" }}
          >
            {availableCurrencies.map((currency) => (
              <option key={`to-${currency}`} value={currency}>{currency}</option>
            ))}
          </select>
        </div>

        <div style={{ marginBottom: "25px" }}>
          <label style={{ display: "block", marginBottom: "8px", fontWeight: "600", color: "#475569" }}>
            Условие:
          </label>
          <select
            value={condition}
            onChange={(e) => setCondition(e.target.value)}
            style={{ width: "100%", padding: "12px", borderRadius: "8px", border: "1px solid #cbd5e1", fontSize: "16px", outline: "none" }}
          >
            <option value="above">Повысится до...</option>
            <option value="below">Понизится до...</option>
          </select>
        </div>

        <div style={{ marginBottom: "25px" }}>
          <label style={{ display: "flex", justifyContent: "space-between", marginBottom: "8px", fontWeight: "600", color: "#475569" }}>
            <span>Целевой курс валюты:</span>
            <span style={{ fontSize: "13px", color: "#64748b", fontWeight: "normal" }}>
              Текущий курс: 1 {from} = <strong>{currentRate}</strong> {to}
            </span>
          </label>
          <input
            type="number"
            step="0.01"
            value={value}
            onChange={(e) => {
              setError("");
              setValue(e.target.value);
            }}
            placeholder="0.00"
            style={{ padding: "12px", borderRadius: "8px", border: "1px solid #cbd5e1", fontSize: "16px", fontWeight: "500", color: "#1e293b", outline: "none", width: "100%", boxSizing: "border-box" }}
          />
        </div>

        {error && (
          <div style={{ marginBottom: "20px", padding: "12px", background: "#fee2e2", color: "#ef4444", borderRadius: "8px", fontSize: "14px" }}>
            {error}
          </div>
        )}

        <div style={{ display: "flex", gap: "12px" }}>
          <button
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
            onClick={handleCreate}
            onMouseEnter={() => setHoverSave(true)}
            onMouseLeave={() => setHoverSave(false)}
            disabled={loading}
            style={{
              flex: 1, padding: "14px", background: loading ? "#a78bfa" : hoverSave ? "#6d28d9" : "#7c3aed",
              color: "#fff", border: "none", borderRadius: "10px", cursor: loading ? "not-allowed" : "pointer",
              fontWeight: "600", fontSize: "15px", transition: "all 0.2s", boxShadow: "0 4px 6px rgba(124, 58, 237, 0.15)"
            }}
          >
            {loading ? "Создание..." : "Создать уведомление"}
          </button>
        </div>
      </div>
    </div>
  );
}