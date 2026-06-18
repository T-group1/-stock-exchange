import { useParams, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import { fetchChartData } from "../Api/currencyApi";
import { ResponsiveContainer, AreaChart, Area, XAxis, YAxis, Tooltip, CartesianGrid } from "recharts";

interface ChartDataItem {
  date: string;
  rate: number;
}

interface CurrencyDetailPageProps {
  rates: Record<string, number>;
}

export default function CurrencyDetailPage({ rates }: CurrencyDetailPageProps) {
  const { pair } = useParams<{ pair: string }>();
  const navigate = useNavigate();
  
  const [chartData, setChartData] = useState<ChartDataItem[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>("");

  const [base, quote] = pair ? pair.split("-") : ["USD", "RUB"];

  useEffect(() => {
    async function loadData() {
      try {
        setLoading(true);
        const data = await fetchChartData(base, quote);
        setChartData(data);
      } catch (err) {
        setError("Не удалось загрузить историю курсов");
      } finally {
        setLoading(false);
      }
    }
    loadData();
  }, [base, quote]);

  const currentRate = rates[base] && rates[quote] ? (rates[quote] / rates[base]).toFixed(4) : null;

  return (
    <div style={{ padding: "30px 20px", maxWidth: "800px", margin: "0 auto", fontFamily: "sans-serif" }}>
      <button 
        onClick={() => navigate("/")} 
        style={{ background: "none", border: "none", color: "#64748b", cursor: "pointer", fontWeight: "600", marginBottom: "20px", display: "flex", alignItems: "center", gap: "8px" }}
      >
        ← На главную
      </button>

      <div style={{ background: "#fff", padding: "30px", borderRadius: "16px", border: "1px solid #e2e8f0", boxShadow: "0 4px 6px -1px rgba(0,0,0,0.05)" }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline", marginBottom: "30px" }}>
          <h2 style={{ color: "#0f172a", margin: 0 }}>История курса {base} / {quote}</h2>
          {currentRate && <span style={{ fontSize: "20px", fontWeight: "bold", color: "#7c3aed" }}>1 {base} = {currentRate} {quote}</span>}
        </div>

        {loading && (
          <div style={{ height: "300px", display: "flex", alignItems: "center", justifyContent: "center", color: "#64748b" }}>
            Загрузка данных графика...
          </div>
        )}

        {error && (
          <div style={{ height: "300px", display: "flex", alignItems: "center", justifyContent: "center", color: "#ef4444" }}>
            {error}
          </div>
        )}

        {!loading && !error && (
          <div style={{ height: "300px", width: "100%" }}>
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={chartData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorRate" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#7c3aed" stopOpacity={0.2}/>
                    <stop offset="95%" stopColor="#7c3aed" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#f1f5f9" />
                <XAxis dataKey="date" stroke="#94a3b8" fontSize={12} tickLine={false} />
                <YAxis stroke="#94a3b8" fontSize={12} domain={['auto', 'auto']} tickLine={false} />
                <Tooltip contentStyle={{ background: "#fff", borderRadius: "8px", border: "1px solid #e2e8f0" }} />
                <Area type="monotone" dataKey="rate" stroke="#7c3aed" strokeWidth={2} fillOpacity={1} fill="url(#colorRate)" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        )}
      </div>
    </div>
  );
}