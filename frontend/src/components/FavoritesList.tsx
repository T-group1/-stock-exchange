import { useState } from "react";
import { apiFetch } from "../Api/apiClient";

export default function FavoritesList({ favorites = [], setFavorites, setFrom, setTo }: any) {
  // Меняем тип состояния с number на string, так как ключи будут вида "USD_RUB"
  const [hoveredCardId, setHoveredCardId] = useState<string | null>(null);
  const [hoveredDeleteId, setHoveredDeleteId] = useState<string | null>(null);
  const [hoveredTextId, setHoveredTextId] = useState<string | null>(null);

  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);

  // Обновленная функция удаления (теперь стучится на бэкенд)
  const removeFavorite = async (from: string, to: string) => {
    const pairString = `${from}_${to}`;
    try {
      // 1. Отправляем запрос на удаление
      const response = await apiFetch(`/api/favorites/${pairString}`, {
        method: "DELETE",
      });

      // 2. Если бэкенд ответил "ОК", удаляем карточку с экрана
      if (response.ok) {
        setFavorites(favorites.filter((f: any) => `${f.from}_${f.to}` !== pairString));
      } else {
        console.error("Сервер вернул ошибку при удалении");
      }
    } catch (error) {
      console.error("Ошибка сети при удалении:", error);
    }
  };

  const handleDragStart = (index: number) => {
    setDraggedIndex(index);
  };

  const handleDragOver = (index: number) => {
    if (draggedIndex === null || draggedIndex === index) return;

    const updatedList = [...favorites];
    const draggedItem = updatedList[draggedIndex];

    updatedList.splice(draggedIndex, 1);
    updatedList.splice(index, 0, draggedItem);

    setDraggedIndex(index);
    setFavorites(updatedList);
  };

  const handleDragEnd = () => {
    setDraggedIndex(null);
  };

  return (
    <div style={{
      background: "#f5f3ff",
      padding: "25px",
      borderRadius: "16px",
      border: "1px solid #ddd6fe",
      marginTop: "25px",
      boxShadow: "0 4px 6px -1px rgba(0, 0, 0, 0.05)",
      fontFamily: "sans-serif"
    }}>
      <h3 style={{ margin: "0 0 5px 0", fontSize: "16px", color: "#6d28d9", fontWeight: "600" }}>
        ⭐ Избранные валютные пары ({favorites.length}/10)
      </h3>
      <p style={{ margin: "0 0 15px 0", fontSize: "12px", color: "#94a3b8" }}>
        Зажмите `⋮⋮` для перетаскивания или кликните на название пары, чтобы открыть её на графике
      </p>

      {favorites.length === 0 ? (
        <div style={{ textAlign: "center", padding: "30px", color: "#94a3b8", border: "2px dashed #ddd6fe", borderRadius: "12px", background: "#fff" }}>
          Список избранных пар пока пуст.
        </div>
      ) : (
        <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
          {favorites.map((f: any, index: number) => {
            // Создаем уникальный текстовый ключ для каждой пары
            const pairKey = `${f.from}_${f.to}`;

            return (
              <div
                key={pairKey}
                draggable={true}
                onDragStart={() => handleDragStart(index)}
                onDragOver={(e) => { e.preventDefault(); handleDragOver(index); }}
                onDragEnd={handleDragEnd}
                onMouseEnter={() => setHoveredCardId(pairKey)}
                onMouseLeave={() => { setHoveredCardId(null); setHoveredTextId(null); }}
                style={{
                  background: "#fff",
                  border: hoveredCardId === pairKey ? "1px solid #7c3aed" : "1px solid #e2e8f0",
                  borderRadius: "10px",
                  padding: "10px 16px",
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                  cursor: "grab",
                  opacity: draggedIndex === index ? 0.5 : 1,
                  boxShadow: hoveredCardId === pairKey ? "0 4px 10px rgba(124, 58, 237, 0.06)" : "none",
                  transition: "border 0.2s ease, box-shadow 0.2s ease, opacity 0.1s"
                }}
              >
                <div style={{ display: "flex", alignItems: "center", gap: "4px", flex: 1 }}>
                  <span style={{ color: "#cbd5e1", fontSize: "16px", userSelect: "none", padding: "0 8px 0 4px" }}>
                    ⋮⋮
                  </span>

                  <div
                    onClick={() => {
                      if (typeof setFrom === "function" && typeof setTo === "function") {
                        setFrom(f.from);
                        setTo(f.to);
                        window.scrollTo({ top: 0, behavior: "smooth" });
                      }
                    }}
                    onMouseEnter={() => setHoveredTextId(pairKey)}
                    onMouseLeave={() => setHoveredTextId(null)}
                    style={{
                      display: "flex",
                      alignItems: "center",
                      padding: "6px 12px",
                      borderRadius: "6px",
                      background: hoveredTextId === pairKey ? "#f3e8ff" : "transparent",
                      cursor: "pointer",
                      transition: "all 0.15s ease",
                      flex: 1
                    }}
                  >
                    <span style={{
                      fontWeight: "600",
                      color: hoveredTextId === pairKey ? "#7c3aed" : "#1e293b",
                      fontSize: "15px"
                    }}>
                      {f.from} ➔ {f.to}
                    </span>
                  </div>
                </div>

                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    removeFavorite(f.from, f.to); // Передаем from и to для корректного удаления
                  }}
                  onMouseEnter={() => setHoveredDeleteId(pairKey)}
                  onMouseLeave={() => setHoveredDeleteId(null)}
                  style={{
                    background: hoveredDeleteId === pairKey ? "#fee2e2" : "transparent",
                    border: "none",
                    color: hoveredDeleteId === pairKey ? "#ef4444" : "#94a3b8",
                    borderRadius: "6px",
                    width: "28px",
                    height: "28px",
                    cursor: "pointer",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    fontSize: "12px",
                    transition: "all 0.15s ease",
                    marginLeft: "10px"
                  }}
                >
                  ✕
                </button>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}