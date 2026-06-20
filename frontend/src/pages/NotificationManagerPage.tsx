import { useNavigate } from "react-router-dom";
import { useState } from "react";
import { apiFetch } from "../Api/apiClient";

export default function NotificationManagerPage({
  user,
  setUser,
  notifications,
  setNotifications,
  setFavorites
}: any) {
  const navigate = useNavigate();
  const [hoverCreate, setHoverCreate] = useState(false);
  const [hoverBtnLogout, setHoverBtnLogout] = useState(false);

  if (!user) {
    return (
      <div style={{ padding: "30px 20px", display: "flex", justifyContent: "center" }}>
        <div style={{
          maxWidth: "400px",
          width: "100%",
          background: "#f5f3ff",
          border: "1px solid #c084fc",
          borderRadius: "12px",
          padding: "25px",
          textAlign: "center",
          boxShadow: "0 4px 6px -1px rgba(0,0,0,0.05)"
        }}>
          <h3 style={{ margin: "0 0 15px 0", color: "#4b5563", fontSize: "18px", fontWeight: "600" }}>
            Вы не вошли в систему
          </h3>
          <p style={{ color: "#6b7280", fontSize: "14px", margin: "0 0 20px 0", lineHeight: "1.5" }}>
            Для доступа к личному кабинету и управлению уведомлениями необходимо авторизоваться.
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
              fontWeight: "600",
              fontSize: "14px",
              boxShadow: "0 2px 4px rgba(124, 58, 237, 0.2)"
            }}
          >
            Войти или зарегистрироваться
          </button>
        </div>
      </div>
    );
  }

  const handleLogout = () => {
    setUser(null);
    setFavorites([]);
    setNotifications([]);

    localStorage.removeItem("user");
    localStorage.removeItem("notifications");
    localStorage.removeItem("favorites");
    localStorage.removeItem("token");
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");

    navigate("/");
  };

  // ИСПРАВЛЕНО: удаляем подписку и на бэкенде тоже
  const removeNotification = async (id: string) => {
    try {
      const response = await apiFetch(`/api/subscriptions/${id}`, {
        method: "DELETE"
      });

      if (!response.ok) {
        console.error("Failed to delete subscription on backend");
      }
    } catch (err) {
      console.error("Error deleting subscription:", err);
    }

    // Удаляем из локального состояния
    setNotifications(notifications.filter((n: any) => n.id !== id));
  };

  const activeNotifications = notifications.filter((n: any) => n.is_active !== false);
  const inactiveNotifications = notifications.filter((n: any) => n.is_active === false);

  return (
    <div style={{ maxWidth: "600px", margin: "0 auto", padding: "20px" }}>

      <div style={{ marginBottom: "30px", background: "#f8fafc", padding: "20px", borderRadius: "12px", border: "1px solid #e2e8f0" }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          <div>
            <h2 style={{ margin: 0, fontSize: "20px", color: "#0f172a" }}>{user.name ? `Здравствуйте, ${user.name}!` : "Личный кабинет"} </h2>
            <p style={{ color: "#64748b", fontSize: "13px", margin: "4px 0 0 0" }}>{user.email}</p>
          </div>

          <button
            onClick={handleLogout}
            onMouseEnter={() => setHoverBtnLogout(true)}
            onMouseLeave={() => setHoverBtnLogout(false)}
            style={{
              padding: "6px 12px",
              background: hoverBtnLogout ? "#fee2e2" : "#fff",
              border: hoverBtnLogout ? "1px solid #fca5a5" : "1px solid #cbd5e1",
              borderRadius: "6px",
              cursor: "pointer",
              color: hoverBtnLogout ? "#ef4444" : "#64748b",
              fontSize: "13px",
              fontWeight: "500",
              transition: "all 0.15s ease"
            }}
          >
            Выйти
          </button>
        </div>

        <button
          onClick={() => navigate("/create-notification")}
          onMouseEnter={() => setHoverCreate(true)}
          onMouseLeave={() => setHoverCreate(false)}
          style={{
            marginTop: "20px",
            width: "100%",
            padding: "11px",
            background: hoverCreate ? "#6d28d9" : "#7c3aed",
            color: "#fff",
            border: "none",
            borderRadius: "8px",
            cursor: "pointer",
            fontWeight: "600",
            fontSize: "14px",
            transition: "all 0.2s ease",
            boxShadow: "0 2px 4px rgba(124, 58, 237, 0.15)"
          }}
        >
          + Создать новое уведомление
        </button>
      </div>

      <div>
        <h3 style={{ margin: "30px 0 15px 0", color: "#1e293b", fontSize: "20px" }}>Активные уведомления</h3>
        {activeNotifications.length === 0 ? (
          <div style={{ padding: "20px", textAlign: "center", color: "#94a3b8", background: "#f8fafc", borderRadius: "10px", border: "2px dashed #e2e8f0" }}>
            Нет активных уведомлений.
          </div>
        ) : (
          <div style={{ display: "flex", flexDirection: "column", gap: "10px" }}>
            {activeNotifications.map((n: any) => (
              <div key={n.id} style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "15px", background: "#fff", border: "1px solid #e2e8f0", borderRadius: "10px", boxShadow: "0 2px 4px rgba(0,0,0,0.02)" }}>
                <div>
                  <div style={{ fontWeight: "bold", color: "#0f172a", fontSize: "16px" }}>{n.from} / {n.to}</div>
                  <div style={{ fontSize: "14px", color: "#64748b", marginTop: "4px" }}>
                    {n.condition === "above" ? "Повысится до..." : "Понизится до..."} {n.value}
                  </div>
                </div>
                <button
                  onClick={() => removeNotification(n.id)} // УБЕДИСЬ, ЧТО ФУНКЦИЯ УДАЛЕНИЯ НАЗЫВАЕТСЯ ТАК
                  style={{ padding: "8px 16px", background: "#fee2e2", color: "#ef4444", border: "none", borderRadius: "6px", cursor: "pointer", fontWeight: "600", transition: "0.2s" }}
                  onMouseEnter={(e) => e.currentTarget.style.background = "#fca5a5"}
                  onMouseLeave={(e) => e.currentTarget.style.background = "#fee2e2"}
                >
                  Удалить
                </button>
              </div>
            ))}
          </div>
        )}

        {/* Блок неактивных (показываем только если они есть) */}
        {inactiveNotifications.length > 0 && (
          <>
            <h3 style={{ margin: "40px 0 15px 0", color: "#94a3b8", fontSize: "18px" }}>Исполненные уведомления</h3>
            <div style={{ display: "flex", flexDirection: "column", gap: "10px" }}>
              {inactiveNotifications.map((n: any) => (
                <div key={n.id} style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "15px", background: "#f1f5f9", border: "1px solid #e2e8f0", borderRadius: "10px", opacity: 0.6 }}>
                  <div>
                    <div style={{ fontWeight: "bold", color: "#64748b", fontSize: "16px", textDecoration: "line-through" }}>{n.from} / {n.to}</div>
                    <div style={{ fontSize: "14px", color: "#94a3b8", marginTop: "4px" }}>
                      Сработало на отметке {n.value}
                    </div>
                  </div>
                  <button
                    onClick={() => removeNotification(n.id)}
                    style={{ padding: "8px 16px", background: "#e2e8f0", color: "#64748b", border: "none", borderRadius: "6px", cursor: "pointer", fontWeight: "600", transition: "0.2s" }}
                    onMouseEnter={(e) => e.currentTarget.style.background = "#cbd5e1"}
                    onMouseLeave={(e) => e.currentTarget.style.background = "#e2e8f0"}
                  >
                    Удалить
                  </button>
                </div>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  );
}