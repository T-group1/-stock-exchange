import { useNavigate } from "react-router-dom";
import { useState } from "react";

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
    const token = localStorage.getItem("token");
    
    try {
      const response = await fetch(`/api/subscriptions/${id}`, {
        method: "DELETE",
        headers: {
          "Authorization": `Bearer ${token}`
        }
      });
      
      if (!response.ok) {
        console.error("Failed to delete subscription on backend");
      }
    } catch (err) {
      console.error("Error deleting subscription:", err);
    }
    
    // Удаляем из локального состояния в любом случае
    setNotifications(notifications.filter((n: any) => n.id !== id));
  };

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

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "20px" }}>
        <h2 style={{ margin: 0 }}>Мои уведомления </h2>
      </div>

      {notifications.length === 0 ? (
        <div style={{ textAlign: "center", padding: "40px", color: "#94a3b8", border: "2px dashed #e2e8f0", borderRadius: "12px" }}>
          У вас пока нет активных уведомлений.
        </div>
      ) : (
        <ul style={{ listStyle: "none", padding: 0 }}>
          {notifications.map((n: any) => (
            <li key={n.id} style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "15px", marginBottom: "10px", background: "#fff", border: "1px solid #e2e8f0", borderRadius: "8px" }}>
              <div>
                <strong style={{ display: "block" }}>{n.from} / {n.to}</strong>
                <span style={{ fontSize: "14px", color: "#64748b" }}>
                  {n.condition === "above" ? "Повысится до..." : "Понизится до..."} {n.value}
                </span>
              </div>
              
              <button 
                onClick={() => removeNotification(n.id)}
                style={{ background: "#fee2e2", border: "none", color: "#ef4444", padding: "8px 12px", borderRadius: "6px", cursor: "pointer" }}
              >
                Удалить
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}