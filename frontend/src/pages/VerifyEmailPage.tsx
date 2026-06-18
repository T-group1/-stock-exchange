import { useEffect, useState } from "react";
import { useSearchParams, useNavigate } from "react-router-dom";

interface VerifyEmailPageProps {
  setUser: (user: any) => void;
}

export default function VerifyEmailPage({ setUser }: VerifyEmailPageProps) {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  
  const [status, setStatus] = useState<"waiting" | "loading" | "success" | "error">("waiting");
  const [errorMessage, setErrorMessage] = useState("");

  const token = searchParams.get("token");

  useEffect(() => {
    if (!token) {
      setStatus("waiting");
      return;
    }

    async function verifyEmail() {
      try {
        setStatus("loading");
        const response = await fetch(`http://localhost:8080/auth/verify?token=${token}`, {
          method: "POST",
          credentials: "include", 
        });

        const data = await response.json();

        if (!response.ok) {
          throw new Error(data.message || "Ошибка верификации");
        }

        setUser(data.user);
        
        setStatus("success");
        setTimeout(() => {
          navigate("/");
        }, 3000);

      } catch (err: any) {
        setStatus("error");
        setErrorMessage(err.message || "Ссылка недействительна или устарела");
      }
    }

    verifyEmail();
  }, [token, navigate, setUser]);

  return (
    <div style={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "100vh", background: "#f8fafc", fontFamily: "sans-serif" }}>
      <div style={{ background: "#fff", padding: "40px", borderRadius: "12px", boxShadow: "0 4px 6px -1px rgba(0, 0, 0, 0.1)", textAlign: "center", maxWidth: "450px", width: "100%" }}>
        
        {status === "waiting" && (
          <>
            <h2 style={{ color: "#0f172a", marginBottom: "15px" }}>Подтвердите email</h2>
            <p style={{ color: "#475569", lineHeight: "1.5" }}>
              Мы отправили письмо с подтверждением. Пожалуйста, перейдите по ссылке в письме, чтобы активировать аккаунт.
            </p>
          </>
        )}

        {status === "loading" && (
          <h2 style={{ color: "#2563eb" }}>Проверяем вашу ссылку...</h2>
        )}

        {status === "success" && (
          <>
            <h2 style={{ color: "#16a34a", marginBottom: "15px" }}>Почта успешно подтверждена!</h2>
            <p style={{ color: "#475569" }}>Добро пожаловать на биржу. Сейчас вы будете перенаправлены...</p>
          </>
        )}

        {status === "error" && (
          <>
            <h2 style={{ color: "#dc2626", marginBottom: "15px" }}>Ошибка активации</h2>
            <p style={{ color: "#ef4444", marginBottom: "20px" }}>{errorMessage}</p>
            <button onClick={() => navigate("/auth")} style={{ padding: "10px 20px", background: "#2563eb", color: "#fff", border: "none", borderRadius: "6px", cursor: "pointer" }}>
              Вернуться к авторизации
            </button>
          </>
        )}

      </div>
    </div>
  );
}