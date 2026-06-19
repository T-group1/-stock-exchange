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
        localStorage.setItem("access_token", data.access_token);
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
        
        // ИСПРАВЛЕНО: Показываем сообщение о необходимости подтвердить email
        setSuccessMessage(data.message || "Пользователь успешно создан. Проверьте почту для подтверждения email.");
        setIsRegistrationSuccess(true);
        
        // Очищаем форму
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

  // Если регистрация успешна, показываем только сообщение
  if (isRegistrationSuccess) {
    return (
      <div style={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "100vh", background: "#f8fafc", fontFamily: "sans-serif" }}>
        <div style={{ background: "#fff", padding: "40px", borderRadius: "12px", boxShadow: "0 4px 6px -1px rgba(0, 0, 0, 0.1)", maxWidth: "450px", width: "100%", textAlign: "center" }}>
          <div style={{ background: "#d1fae5", color: "#059669", padding: "20px", borderRadius: "8px", marginBottom: "24px" }}>
            <svg style={{ width: "48px", height: "48px", margin: "0 auto 16px", display: "block" }} fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <h2 style={{ margin: "0 0 12px", fontSize: "20px" }}>Регистрация успешна!</h2>
            <p style={{ margin: "0 0 8px", lineHeight: "1.5" }}>{successMessage}</p>
            <p style={{ margin: "0", fontSize: "14px", opacity: "0.9" }}>После подтверждения email вы сможете войти в систему.</p>
          </div>
          
          <button 
            onClick={handleGoToLogin}
            style={{ width: "100%", padding: "12px", background: "#2563eb", color: "#fff", border: "none", borderRadius: "6px", fontSize: "16px", fontWeight: "600", cursor: "pointer" }}
          >
            Перейти к входу
          </button>
        </div>
      </div>
    );
  }

  // Обычная форма авторизации/регистрации
  return (
    <div style={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "100vh", background: "#f8fafc", fontFamily: "sans-serif" }}>
      <div style={{ background: "#fff", padding: "40px", borderRadius: "12px", boxShadow: "0 4px 6px -1px rgba(0, 0, 0, 0.1)", maxWidth: "450px", width: "100%" }}>
        <h2 style={{ textAlign: "center", marginBottom: "30px", color: "#0f172a" }}>
          {isLoginMode ? "Вход в систему" : "Регистрация"}
        </h2>

        {error && (
          <div style={{ background: "#fee2e2", color: "#dc2626", padding: "12px", borderRadius: "6px", marginBottom: "20px" }}>
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          {!isLoginMode && (
            <div style={{ marginBottom: "20px" }}>
              <label style={{ display: "block", marginBottom: "8px", color: "#475569", fontWeight: "500" }}>
                Имя
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                style={{ width: "100%", padding: "10px", border: "1px solid #cbd5e1", borderRadius: "6px", fontSize: "14px" }}
              />
            </div>
          )}

          <div style={{ marginBottom: "20px" }}>
            <label style={{ display: "block", marginBottom: "8px", color: "#475569", fontWeight: "500" }}>
              Email
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              style={{ width: "100%", padding: "10px", border: "1px solid #cbd5e1", borderRadius: "6px", fontSize: "14px" }}
            />
          </div>

          <div style={{ marginBottom: "30px" }}>
            <label style={{ display: "block", marginBottom: "8px", color: "#475569", fontWeight: "500" }}>
              Пароль
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
              style={{ width: "100%", padding: "10px", border: "1px solid #cbd5e1", borderRadius: "6px", fontSize: "14px" }}
            />
          </div>

          <button
            type="submit"
            style={{ width: "100%", padding: "12px", background: "#2563eb", color: "#fff", border: "none", borderRadius: "6px", fontSize: "16px", fontWeight: "600", cursor: "pointer" }}
          >
            {isLoginMode ? "Войти" : "Зарегистрироваться"}
          </button>
        </form>

        <p style={{ textAlign: "center", marginTop: "20px", color: "#64748b" }}>
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
    </div>
  );
}