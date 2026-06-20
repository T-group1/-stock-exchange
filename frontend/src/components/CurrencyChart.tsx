import { useEffect, useState, useRef } from "react";
import { ResponsiveContainer, AreaChart, Area, XAxis, YAxis, Tooltip, CartesianGrid } from "recharts";
import { fetchChartData } from "../Api/currencyApi";

interface ChartProps {
  baseCurrency: string;
  quoteCurrency: string;
}

interface Line {
  id: number;
  x1: number;
  y1: number;
  x2: number;
  y2: number;
  length: number;
}

// Маппинг количества дней в период для API
const getPeriodFromDays = (d: number): string => {
  if (d <= 7) return "1w";
  if (d <= 30) return "1m";
  return "3m";
};

export default function CurrencyChart({ baseCurrency, quoteCurrency }: ChartProps) {
  const [data, setData] = useState<any[]>([]);
  const [days, setDays] = useState<number>(30);

  const [lines, setLines] = useState<Line[]>([]);
  const [tool, setTool] = useState<"none" | "draw" | "move" | "rotate">("none");
  const [activeLineId, setActiveLineId] = useState<number | null>(null);
  const [drawingStart, setDrawingStart] = useState<{ x: number; y: number } | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const loadData = async () => {
      try {
        // Передаём период на бэкенд, чтобы получить нужный диапазон данных
        const period = getPeriodFromDays(days);
        const response = await fetchChartData(baseCurrency, quoteCurrency, period);
        
        let rawPoints: any[] = [];
        if (Array.isArray(response)) {
          rawPoints = response;
        } else if (response && typeof response === "object") {
          const keys = Object.keys(response);
          const foundKey = keys.find(k => Array.isArray(response[k]));
          rawPoints = foundKey ? response[foundKey] : (response.data || []);
        }

        if (rawPoints.length === 0) {
          console.warn("Данные для графика отсутствуют или пустой массив");
          setData([]);
          return;
        }

        const formattedData = rawPoints.map((p: any, i: number) => {
          const val = typeof p === 'object' && p !== null 
            ? (p.rate || p.value || p.price || 0) 
            : p;
            
          const extractedRate = Number(val) || 0;

          const dateVal = p && p.date ? p.date : null;
          const dateObj = dateVal ? new Date(dateVal) : null;
          const displayDate = (dateObj && !isNaN(dateObj.getTime()))
              ? dateObj.toLocaleDateString("ru-RU", { day: "numeric", month: "short" })
              : `День ${i + 1}`;

          return { rate: extractedRate, date: displayDate };
        });

        // Данные уже идут от старых к новым (ASC) — просто берём последние N записей
        // БЕЗ reverse(), чтобы график шёл слева направо: прошлое → настоящее
        setData(formattedData.slice(-days));

      } catch (err) {
        console.error("Ошибка при подготовке данных графика:", err);
        setData([]);
      }
    };
    
    loadData();
  }, [baseCurrency, quoteCurrency, days]);

  const getMouseCoords = (e: React.MouseEvent) => {
    if (!containerRef.current) return { x: 0, y: 0 };
    const rect = containerRef.current.getBoundingClientRect();
    const svg = containerRef.current.querySelector("svg");
    const svgRect = svg ? svg.getBoundingClientRect() : rect;
    return {
      x: e.clientX - svgRect.left,
      y: e.clientY - svgRect.top
    };
  };

  const handleChartClick = (e: React.MouseEvent) => {
    const coords = getMouseCoords(e);

    if ((tool === "move" || tool === "rotate") && activeLineId !== null) {
      setActiveLineId(null);
      setTool("none");
      return;
    }

    if (tool === "draw") {
      if (!drawingStart) {
        setDrawingStart(coords);
      } else {
        const dx = coords.x - drawingStart.x;
        const dy = coords.y - drawingStart.y;
        const length = Math.sqrt(dx * dx + dy * dy);

        const newLine: Line = {
          id: Date.now(),
          x1: drawingStart.x,
          y1: drawingStart.y,
          x2: coords.x,
          y2: coords.y,
          length
        };
        setLines([...lines, newLine]);
        setDrawingStart(null);
        setTool("none");
      }
    }
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    if (tool === "none" || activeLineId === null) return;
    const coords = getMouseCoords(e);

    setLines(lines.map(line => {
      if (line.id !== activeLineId) return line;

      if (tool === "move") {
        const dx = line.x2 - line.x1;
        const dy = line.y2 - line.y1;
        return {
          ...line,
          x1: coords.x - dx / 2,
          y1: coords.y - dy / 2,
          x2: coords.x + dx / 2,
          y2: coords.y + dy / 2
        };
      }

      if (tool === "rotate") {
        const angle = Math.atan2(coords.y - line.y1, coords.x - line.x1);
        return {
          ...line,
          x2: line.x1 + line.length * Math.cos(angle),
          y2: line.y1 + line.length * Math.sin(angle)
        };
      }

      return line;
    }));
  };

  if (baseCurrency === quoteCurrency) {
    return <div style={{ textAlign: "center", padding: "20px", color: "#94a3b8" }}>Выберите разные валюты 📈</div>;
  }

  return (
    <div
      ref={containerRef}
      style={{
        background: "#f5f3ff",
        padding: "20px",
        borderRadius: "16px",
        border: "1px solid #ddd6fe",
        position: "relative",
        boxShadow: "0 4px 6px -1px rgba(0, 0, 0, 0.05)"
      }}
    >
      <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "15px" }}>
        <div style={{ display: "flex", gap: "5px", background: "#f1f5f9", padding: "3px", borderRadius: "6px" }}>
          {([[7, "7 дней"], [30, "1 мес"], [90, "3 мес"]] as const).map(([v, label]) => (
            <button key={v} onClick={() => setDays(v)} style={{ padding: "4px 10px", border: "none", borderRadius: "4px", background: days === v ? "#fff" : "transparent", boxShadow: days === v ? "0 1px 3px rgba(0,0,0,0.1)" : "none", cursor: "pointer", fontSize: "12px", fontWeight: days === v ? "bold" : "normal" }}>
              {label}
            </button>
          ))}
        </div>

        <div style={{ display: "flex", gap: "5px" }}>
          <button onClick={() => setTool(tool === "draw" ? "none" : "draw")} style={{ padding: "4px 10px", background: tool === "draw" ? "#7c3aed" : "#f1f5f9", color: tool === "draw" ? "#fff" : "#475569", border: "none", borderRadius: "6px", cursor: "pointer", fontSize: "12px" }}>
            Рисовать линии
          </button>
          <button onClick={() => setTool(tool === "move" ? "none" : "move")} style={{ padding: "4px 10px", background: tool === "move" ? "#0284c7" : "#f1f5f9", color: tool === "move" ? "#fff" : "#475569", border: "none", borderRadius: "6px", cursor: "pointer", fontSize: "12px" }}>
            Перенос линий
          </button>
          <button onClick={() => setTool(tool === "rotate" ? "none" : "rotate")} style={{ padding: "4px 10px", background: tool === "rotate" ? "#f59e0b" : "#f1f5f9", color: tool === "rotate" ? "#fff" : "#475569", border: "none", borderRadius: "6px", cursor: "pointer", fontSize: "12px" }}>
            Вращение линий
          </button>
          {lines.length > 0 && (
            <button onClick={() => { setLines([]); setActiveLineId(null); }} style={{ padding: "4px 10px", background: "#fee2e2", color: "#ef4444", border: "none", borderRadius: "6px", cursor: "pointer", fontSize: "12px" }}>
              Очистить
            </button>
          )}
        </div>
      </div>

      <div style={{ fontSize: "11px", color: "#94a3b8", marginBottom: "10px" }}>
        {tool === "draw" && (drawingStart ? " Кликните второй раз, чтобы завершить линию" : " Кликните на график, чтобы начать рисовать")}
        {tool === "move" && " Выберите линию ниже (кликните на неё) и двигайте мышь для переноса"}
        {tool === "rotate" && " Выберите линию ниже и двигайте мышь для вращения вокруг начальной точки"}
      </div>

      <div
        onClick={handleChartClick}
        onMouseMove={handleMouseMove}
        style={{ position: "relative" }}
      >
        <ResponsiveContainer width="100%" height={250}>
          <AreaChart data={data} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#f1f5f9" />
            <XAxis dataKey="date" stroke="#94a3b8" fontSize={11} tickLine={false} />
            <YAxis stroke="#94a3b8" fontSize={11} tickLine={false} domain={["auto", "auto"]} />
            <Tooltip contentStyle={{ background: "#1e293b", borderRadius: "8px", color: "#fff", border: "none" }} labelStyle={{ color: "#94a3b8" }} />
            <Area type="monotone" dataKey="rate" stroke="#7c3aed" fillOpacity={1} fill="url(#colorRate)" />
            <defs>
              <linearGradient id="colorRate" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#7c3aed" stopOpacity={0.1} />
                <stop offset="95%" stopColor="#7c3aed" stopOpacity={0} />
              </linearGradient>
            </defs>
          </AreaChart>
        </ResponsiveContainer>

        <svg
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            width: "100%",
            height: "100%",
            pointerEvents: tool === "none" ? "none" : "auto"
          }}
        >
          {lines.map((line) => (
            <g key={line.id}>
              <line
                x1={line.x1} y1={line.y1} x2={line.x2} y2={line.y2}
                stroke="transparent" strokeWidth={15} style={{ cursor: "pointer" }}
                onClick={(e) => { e.stopPropagation(); setActiveLineId(line.id); }}
              />

              <line
                x1={line.x1} y1={line.y1} x2={line.x2} y2={line.y2}
                stroke={activeLineId === line.id ? "#ef4444" : "#06b6d4"}
                strokeWidth={activeLineId === line.id ? 3 : 2}
                strokeDasharray={activeLineId === line.id ? "4 4" : "none"}
              />

              <circle
                cx={line.x1}
                cy={line.y1}
                r={5}
                fill="#3b82f6"
                stroke="#fff"
                strokeWidth={1.5}
                style={{ filter: "drop-shadow(0px 1px 2px rgba(0,0,0,0.2))" }}
              />

              <circle
                cx={line.x2}
                cy={line.y2}
                r={5}
                fill="#3b82f6"
                stroke="#fff"
                strokeWidth={1.5}
                style={{ filter: "drop-shadow(0px 1px 2px rgba(0,0,0,0.2))" }}
              />
            </g>
          ))}
        </svg>
      </div>
    </div>
  );
}