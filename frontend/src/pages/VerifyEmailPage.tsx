import { useLocation, useNavigate } from "react-router-dom";

export default function VerifyEmailPage({ setUser }: any) {
  const location = useLocation();
  const navigate = useNavigate();
  const email = location.state?.email || "вашу почту";

  const handleSimulateEmailClick = () => {
    const savedUserString = localStorage.getItem("pending_user");
    
    if (savedUserString) {
      const savedUser = JSON.parse(savedUserString);
      setUser({ name: savedUser.name, email: savedUser.email });
      localStorage.removeItem("pending_user");
    }

    navigate("/");
  };

  return (
    <div style={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "100vh", background: "#f8fafc", fontFamily: "sans-serif" }}>
      <div style={{ background: "#fff", padding: "40px", borderRadius: "12px", boxShadow: "0 4px 6px -1px rgba(0, 0, 0, 0.1)", textAlign: "center", maxWidth: "450px" }}>
        <h2 style={{ color: "#0f172a", marginBottom: "15px" }}>Подтвердите email</h2>
        <p style={{ color: "#475569", lineHeight: "1.5", marginBottom: "30px" }}>
          Мы отправили письмо с подтверждением на адрес <br />
          <strong style={{ color: "#0f172a" }}>{email}</strong>
        </p>
        
        <button 
          onClick={handleSimulateEmailClick} 
          style={{ width: "100%", padding: "12px", background: "#22c55e", color: "#fff", border: "none", borderRadius: "6px", cursor: "pointer", fontSize: "16px", fontWeight: "bold" }}
        >
          Я перешел по ссылке в письме (Демо)
        </button>
      </div>
    </div>
  );
}