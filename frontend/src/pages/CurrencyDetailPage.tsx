import { useParams, useNavigate } from "react-router-dom";

export default function CurrencyDetailPage() {
  const { pair } = useParams();
  const navigate = useNavigate();

  return (
    <div style={{ padding: 40, maxWidth: 800, margin: "0 auto" }}>
      <h2>Детали валютной пары {pair?.replace("-", "/")}</h2>
      <div className="card" style={{ height: 300, display: "flex", alignItems: "center", justifyItems: "center", justifyContent: "center", background: "#f1f5f9" }}>
        <p style={{ color: "#64748b" }}></p>
      </div>
      <button onClick={() => navigate("/")} style={{marginTop: 20, background: "none", border: "none", color: "#64748b", cursor: "pointer"}}>← На главную</button>
    </div>
  );
}