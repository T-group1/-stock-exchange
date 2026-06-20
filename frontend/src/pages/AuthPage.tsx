import { useState } from "react";
import { useNavigate } from "react-router-dom";

interface AuthPageProps {
  setUser: (user: any) => void;
}

export default function AuthPage({ setUser }: AuthPageProps) {
  const [isLoginMode, setIsLoginMode] = useState(true);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [error, setError] = useState("");
  const [successMessage, setSuccessMessage] = useState("");
  const [isRegistrationSuccess, setIsRegistrationSuccess] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccessMessage("");

    try {
      if (isLoginMode) {
        // --- ЛОГИКА ВХОДА ---
        const res = await fetch('/api/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password })
        });

        if (!res.ok) {
          const errorData = await res.json();
          throw new Error(errorData.message || "Неверный email или пароль");
        }

        const data = await res.json();
        setUser(data.user);
        localStorage.setItem("user", JSON.stringify(data.user));
        
        // ИСПРАВЛЕНО: сохраняем токен под обоими ключами для совместимости
        localStorage.setItem("access_token", data.access_token);
        localStorage.setItem("token", data.access_token);
        localStorage.setItem("refresh_token", data.refresh_token);
        
        navigate("/");
      } else {
        // --- ЛОГИКА РЕГИСТРАЦИИ ---
        if (!name) {
          setError("Пожалуйста, введите ваше имя для регистрации");
          return;
        }

        if (password.length < 8) {
          setError("Пароль должен быть не менее 8 символов");
          return;
        }

        const res = await fetch('/api/auth/register', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name, email, password })
        });

        if (!res.ok) {
          const errorData = await res.json();
          throw new Error(errorData.message || "Ошибка регистрации. Возможно, такой email уже используется или сервер недоступен.");
        }

        const data = await res.json();

        setSuccessMessage(data.message || "Пользователь успешно создан. Проверьте почту для подтверждения email.");
        setIsRegistrationSuccess(true);

        setName("");
        setEmail("");
        setPassword("");
      }
    } catch (err: any) {
      console.error(err);
      setError(err.message || "Ошибка при подключении к серверу");
    }
  };

  const handleGoToLogin = () => {
    setIsLoginMode(true);
    setIsRegistrationSuccess(false);
    setSuccessMessage("");
    setError("");
  };

  if (isRegistrationSuccess) {
    return (
      <div style={{ maxWidth: "500px", margin: "60px auto", padding: "30px", textAlign: "center", background: "#f0fdf4", border: "1px solid #86efac", borderRadius: "12px" }}>
        <h2 style={{ color: "#166534", marginBottom: "15px" }}>Регистрация успешна!</h2>
        <p style={{ color: "#15803d", fontSize: "16px", lineHeight: "1.6" }}>{successMessage}</p>
        <p style={{ color: "#64748b", fontSize: "14px", marginTop: "15px" }}>После подтверждения email вы сможете войти в систему.</p>
        <button
          onClick={handleGoToLogin}
          style={{ marginTop: "20px", padding: "10px 24px", background: "#2563eb", color: "#fff", border: "none", borderRadius: "8px", cursor: "pointer", fontWeight: "600" }}
        >
          Перейти к входу
        </button>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: "400px", margin: "60px auto", padding: "30px", background: "#fff", border: "1px solid #e2e8f0", borderRadius: "12px", boxShadow: "0 4px 6px -1px rgba(0,0,0,0.05)" }}>
      <h2 style={{ textAlign: "center", marginBottom: "25px", color: "#0f172a" }}>{isLoginMode ? "Вход в систему" : "Регистрация"}</h2>
      
      {error && (
        <div style={{ marginBottom: "20px", padding: "12px", background: "#fee2e2", color: "#ef4444", borderRadius: "8px", fontSize: "14px" }}>
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit}>
        {!isLoginMode && (
          <div style={{ marginBottom: "15px" }}>
            <label style={{ display: "block", marginBottom: "6px", fontWeight: "500", color: "#475569", fontSize: "14px" }}>Имя</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              style={{ width: "100%", padding: "10px", border: "1px solid #cbd5e1", borderRadius: "6px", fontSize: "14px", boxSizing: "border-box" }}
            />
          </div>
        )}

        <div style={{ marginBottom: "15px" }}>
          <label style={{ display: "block", marginBottom: "6px", fontWeight: "500", color: "#475569", fontSize: "14px" }}>Email</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            style={{ width: "100%", padding: "10px", border: "1px solid #cbd5e1", borderRadius: "6px", fontSize: "14px", boxSizing: "border-box" }}
          />
        </div>

        <div style={{ marginBottom: "20px" }}>
          <label style={{ display: "block", marginBottom: "6px", fontWeight: "500", color: "#475569", fontSize: "14px" }}>Пароль</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
            style={{ width: "100%", padding: "10px", border: "1px solid #cbd5e1", borderRadius: "6px", fontSize: "14px", boxSizing: "border-box" }}
          />
        </div>

        <button
          type="submit"
          style={{ width: "100%", padding: "12px", background: "#2563eb", color: "#fff", border: "none", borderRadius: "8px", cursor: "pointer", fontWeight: "600", fontSize: "15px" }}
        >
          {isLoginMode ? "Войти" : "Зарегистрироваться"}
        </button>
      </form>

      <p style={{ textAlign: "center", marginTop: "20px", color: "#64748b", fontSize: "14px" }}>
        {isLoginMode ? "Нет аккаунта?" : "Уже есть аккаунт?"}{" "}
        <button
          onClick={() => {
            setIsLoginMode(!isLoginMode);
            setError("");
            setSuccessMessage("");
            setIsRegistrationSuccess(false);
          }}
          style={{ background: "none", border: "none", color: "#2563eb", cursor: "pointer", fontWeight: "600" }}
        >
          {isLoginMode ? "Зарегистрироваться" : "Войти"}
        </button>
      </p>
    </div>
  );
}