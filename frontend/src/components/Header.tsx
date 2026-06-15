import { useState } from "react";
import { useNavigate } from "react-router-dom";

export default function Header({ user }: any) {
  const navigate = useNavigate();
  
  const [hoverBtn, setHoverBtn] = useState(false);
  const [hoverProfile, setHoverProfile] = useState(false);

  return (
    <header style={{ 
      background: "linear-gradient(135deg, #7c3aed 0%, #5b21b6 100%)", 
      padding: "30px 40px", 
      borderRadius: "16px",
      marginBottom: "30px",
      display: "flex",
      justifyContent: "space-between",
      alignItems: "center",
      boxShadow: "0 10px 15px -3px rgba(124, 58, 237, 0.2)", 
      border: "1px solid #6d28d9",
      fontFamily: "sans-serif"
    }}>
      
      <div>
        <h1 
          onClick={() => navigate("/")}
          style={{ 
            margin: 0, 
            color: "#fff", 
            fontSize: "26px", 
            fontWeight: "800", 
            letterSpacing: "-0.5px",
            cursor: "pointer"
          }}
        >
          Мониторинг валютных курсов 
        </h1>
        <p style={{ margin: "6px 0 0 0", color: "#ddd6fe", fontSize: "14px", fontWeight: "400" }}>
          Интерактивный терминал конвертации и трендового анализа
        </p>
      </div>

      {user ? (
    
        <div 
          onClick={() => navigate("/profile")} 
          onMouseEnter={() => setHoverProfile(true)}
          onMouseLeave={() => setHoverProfile(false)}
          style={{
            display: "flex",
            alignItems: "center",
            gap: "8px",

            background: hoverProfile ? "#ffffff" : "rgba(255, 255, 255, 0.15)",
            color: hoverProfile ? "#5b21b6" : "#ffffff", 
            padding: "10px 20px",
            borderRadius: "20px",
            fontWeight: "600",
            fontSize: "14px",
            cursor: "pointer",
            border: hoverProfile ? "1px solid #fff" : "1px solid rgba(255, 255, 255, 0.25)",
            transform: hoverProfile ? "scale(1.03)" : "scale(1)",
            transition: "all 0.2s cubic-bezier(0.4, 0, 0.2, 1)",
            boxShadow: hoverProfile ? "0 4px 12px rgba(0,0,0,0.1)" : "none"
          }}
        >
          <span> {user.name}</span>
        </div>
      ) : (

        <button 
          onClick={() => navigate("/auth")}
          onMouseEnter={() => setHoverBtn(true)}
          onMouseLeave={() => setHoverBtn(false)}
          style={{
            padding: "12px 24px",
            background: hoverBtn ? "#ffffff" : "#f5f3ff", 
            color: "#5b21b6", 
            border: "none",
            borderRadius: "10px",
            fontWeight: "700",
            fontSize: "14px",
            cursor: "pointer",
            boxShadow: hoverBtn ? "0 4px 12px rgba(0,0,0,0.15)" : "0 2px 4px rgba(0,0,0,0.05)",
            transform: hoverBtn ? "scale(1.03)" : "scale(1)",
            transition: "all 0.2s cubic-bezier(0.4, 0, 0.2, 1)"
          }}
        >
          Личный кабинет
        </button>
      )}

    </header>
  );
}