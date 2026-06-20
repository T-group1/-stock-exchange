const API_BASE_URL = "http://localhost:8080";

export interface FavoritePair {
  from: string;
  to: string;
}

export interface FavoritesResponse {
  favorite_pairs: string[];
}

// Получить токен из localStorage
const getToken = (): string | null => {
  const user = localStorage.getItem("user");
  if (!user) return null;
  const userData = JSON.parse(user);
  return userData.accessToken || null;
};

// Получить список избранных пар
export const fetchFavorites = async (): Promise<FavoritesResponse> => {
  const token = getToken();
  if (!token) {
    return { favorite_pairs: [] };
  }

  const response = await fetch(`${API_BASE_URL}/favorites`, {
    method: "GET",
    headers: {
      "Authorization": `Bearer ${token}`,
      "Content-Type": "application/json",
    },
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch favorites: ${response.statusText}`);
  }

  return response.json();
};

// Добавить пару в избранное
export const addFavorite = async (currencyPair: string): Promise<void> => {
  const token = getToken();
  if (!token) {
    throw new Error("User not authenticated");
  }

  const response = await fetch(`${API_BASE_URL}/favorites`, {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ currency_pair: currencyPair }),
  });

  if (!response.ok) {
    throw new Error(`Failed to add favorite: ${response.statusText}`);
  }
};

// Удалить пару из избранного
export const removeFavorite = async (currencyPair: string): Promise<void> => {
  const token = getToken();
  if (!token) {
    throw new Error("User not authenticated");
  }

  const response = await fetch(`${API_BASE_URL}/favorites/${currencyPair}`, {
    method: "DELETE",
    headers: {
      "Authorization": `Bearer ${token}`,
      "Content-Type": "application/json",
    },
  });

  if (!response.ok) {
    throw new Error(`Failed to remove favorite: ${response.statusText}`);
  }
};