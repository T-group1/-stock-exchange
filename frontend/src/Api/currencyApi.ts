export const mockRates: Record<string, number> = {
  USD: 92.45,
  EUR: 102.30,
  CNY: 12.85,
  GBP: 117.50,
  JPY: 0.62,
  CAD: 67.80,
  AUD: 61.25,
  CHF: 105.90,
  HKD: 11.85,
  SGD: 68.45,
  RUB: 1.00
};

export const mockChanges: Record<string, number> = {
  USD: 1.2, EUR: -0.8, CNY: 0.5, GBP: 0.9, JPY: -0.3, CAD: 1.5, AUD: -1.1, CHF: 0.4, HKD: 0.7, SGD: -0.6
};


export async function fetchRates(): Promise<any> {
  try {
    const response = await fetch('/api/rates');

    if (!response.ok) {
      throw new Error(`Ошибка сервера: ${response.status}`);
    }

    const responseData = await response.json();
    const ratesRecord: Record<string, number> = {};

    if (responseData && responseData.rates && !Array.isArray(responseData.rates)) {
      Object.keys(responseData.rates).forEach((code) => {
        const item = responseData.rates[code];
 
        const value = typeof item === 'object' && item !== null 
          ? (item.rate || item.value || item.price) 
          : item;

        if (value) {
          ratesRecord[code] = Number(value);
        }
      });
      
      if (!ratesRecord["RUB"]) ratesRecord["RUB"] = 1.0;
      
      return ratesRecord;
    }

    if (responseData.rates && Array.isArray(responseData.rates)) {
      responseData.rates.forEach((item: any) => {
        const code = item.currency || item.code || item.id;
        const value = item.rate || item.value;
        if (code && value) {
          ratesRecord[code] = Number(value);
        }
      });
      return ratesRecord;
    }

    return responseData;

  } catch (error) {
    console.error("Сервер недоступен, показываем моки:", error);
    return mockRates;
  }
}

export async function fetchChartData(base: string, quote: string) {
  try {
    const pair = `${base}_${quote}`.toUpperCase();

    const response = await fetch(`/api/rates/${pair}`);

    if (!response.ok) {
      throw new Error(`Ошибка сервера: ${response.status}`);
    }

    const responseData = await response.json();

    return responseData.data || responseData;

  } catch (error) {
    console.error("Не удалось загрузить график, генерируем моки:", error);

    const baseRate = mockRates[base] / (mockRates[quote] || 1);
    const mockData = [];

    for (let i = 12; i <= 30; i += 2) {
      mockData.push({ date: `${i} мая`, rate: +(baseRate * (1 + (Math.random() * 0.04 - 0.02))).toFixed(4) });
    }
    for (let i = 1; i <= 9; i += 2) {
      mockData.push({ date: `${i} июня`, rate: +(baseRate * (1 + (Math.random() * 0.04 - 0.02))).toFixed(4) });
    }

    return mockData;
  }
}
