// frontend/src/Api/apiClient.ts

// Функция для обновления токена
async function refreshAuthToken(): Promise<string | null> {
    const refreshToken = localStorage.getItem("refresh_token");
    if (!refreshToken) return null;

    try {
        const response = await fetch("/api/auth/refresh", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ refresh_token: refreshToken }),
        });

        if (!response.ok) {
            throw new Error("Failed to refresh token");
        }

        const data = await response.json();

        // Сохраняем новые токены в localStorage
        localStorage.setItem("token", data.access_token);
        localStorage.setItem("refresh_token", data.refresh_token);

        return data.access_token;
    } catch (error) {
        console.error("Token refresh failed:", error);
        // Если рефреш не удался (например, он тоже истек), выкидываем пользователя
        localStorage.removeItem("token");
        localStorage.removeItem("refresh_token");
        localStorage.removeItem("user");
        // Перенаправляем на логин (лучше делать через React Router, но это запасной вариант)
        window.location.href = "/auth";
        return null;
    }
}

// Универсальная обертка над обычным fetch
export async function apiFetch(url: string, options: RequestInit = {}) {
    let token = localStorage.getItem("token");

    // Подставляем текущий токен в заголовки
    const headers = new Headers(options.headers || {});
    if (token) {
        headers.set("Authorization", `Bearer ${token}`);
    }

    const initialOptions = { ...options, headers };

    // 1. Выполняем оригинальный запрос
    let response = await fetch(url, initialOptions);

    // 2. Если получаем 401, пробуем обновить токен
    if (response.status === 401) {
        console.warn("Access token expired. Attempting to refresh...");

        const newToken = await refreshAuthToken();

        // 3. Если удалось получить новый токен, повторяем изначальный запрос
        if (newToken) {
            headers.set("Authorization", `Bearer ${newToken}`);
            response = await fetch(url, { ...options, headers });
        }
    }

    // Возвращаем итоговый ответ (успешный или с ошибкой, которую обработает компонент)
    return response;
}