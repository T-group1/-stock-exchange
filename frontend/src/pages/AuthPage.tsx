import { useState } from "react";
import { useNavigate } from "react-router-dom";

export default function AuthPage({ setUser }: any) {
  const navigate = useNavigate();
  const [isLoginMode, setIsLoginMode] = useState<boolean>(true);
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [name, setName] = useState<string>("");
  const [error, setError] = useState<string>("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (isLoginMode) {
      try {
        const response = await fetch('/api/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password })
        });

        if (response.ok) {
          const userData = await response.json();
          setUser(userData);
          navigate("/profile"); 
        } else {
          setError("Неверный email или пароль");
        }
      } catch (err) {
        console.error(err);
        setError("Ошибка при подключении к серверу");
      }
    } else {
      if (!name) {
        setError("Пожалуйста, введите ваше имя для регистрации");
        return;
      }

      if (password.length < 8) {
        setError("Пароль должен быть не менее 8 символов");
        return;
      }
      
      try {
        const response = await fetch('/api/auth/register', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password, name })
        });

        if (response.ok) {
          navigate("/verify-email", { state: { email } }); 
        } else {
          setError("Ошибка регистрации. Возможно, такой email уже используется.");
        }
      } catch (err) {
        console.error(err);
        setError("Ошибка при подключении к серверу");
      }
    }
  };

  return (
    <div style={{ maxWidth: "400px", margin: "50px auto", padding: "30px", background: "#fff", borderRadius: "12px", border: "1px solid #e2e8f0", fontFamily: "sans-serif" }}>
      <h2 style={{ textAlign: "center", color: "#1e293b", marginBottom: "20px" }}>
        {isLoginMode ? "Войти в аккаунт" : "Регистрация"}
      </h2>

      {error && (
        <div style={{ background: "#fee2e2", color: "#ef4444", padding: "10px", borderRadius: "8px", fontSize: "14px", marginBottom: "15px", border: "1px solid #fca5a5" }}>
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
        {!isLoginMode && (
          <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
            <label style={{ fontSize: "14px", color: "#475569", fontWeight: "600" }}>Ваше имя</label>
            <input 
              type="text" 
              value={name} 
              onChange={(e) => setName(e.target.value)} 
              placeholder="Как Вас зовут?"
              style={{ padding: "10px", borderRadius: "8px", border: "1px solid #cbd5e1", fontSize: "15px" }}
            />
          </div>
        )}

        <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
          <label style={{ fontSize: "14px", color: "#475569", fontWeight: "600" }}>Email</label>
          <input 
            type="email" 
            required 
            value={email} 
            onChange={(e) => setEmail(e.target.value)} 
            placeholder="example@mail.ru"
            style={{ padding: "10px", borderRadius: "8px", border: "1px solid #cbd5e1", fontSize: "15px" }}
          />
        </div>

        <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
          <label style={{ fontSize: "14px", color: "#475569", fontWeight: "600" }}>Пароль</label>
          <input 
            type="password" 
            required 
            value={password} 
            onChange={(e) => setPassword(e.target.value)} 
            placeholder="••••••••"
            style={{ padding: "10px", borderRadius: "8px", border: "1px solid #cbd5e1", fontSize: "15px" }}
          />
        </div>

        <button 
          type="submit" 
          style={{ padding: "12px", background: "#7c3aed", color: "#fff", border: "none", borderRadius: "8px", cursor: "pointer", fontSize: "16px", fontWeight: "600", marginTop: "10px" }}
        >
          {isLoginMode ? "Войти" : "Зарегистрироваться"}
        </button>
      </form>

      <div style={{ textAlign: "center", marginTop: "20px" }}>
        <button 
          type="button"
          onClick={() => { setError(""); setIsLoginMode(!isLoginMode); }} 
          style={{ background: "none", border: "none", color: "#7c3aed", cursor: "pointer", fontSize: "14px", fontWeight: "600" }}
        >
          {isLoginMode ? "Ещё нет аккаунта? Создать" : "Уже есть аккаунт? Войти"}
        </button>
      </div>
    </div>
  );
}
