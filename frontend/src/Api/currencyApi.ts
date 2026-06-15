// Fake data base 
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
// Changes in percent
export const mockChanges: Record<string, number> = {
  USD: 1.2, EUR: -0.8, CNY: 0.5, GBP: 0.9, JPY: -0.3, CAD: 1.5, AUD: -1.1, CHF: 0.4, HKD: 0.7, SGD: -0.6
};

export async function fetchRates(): Promise<Record<string, number>> {
  return new Promise((resolve) => {
    setTimeout(() => resolve(mockRates), 300);
  });
}

// Graphycs generation (last 30d)
export function fetchChartData(base: string, quote: string) {
  const data = [];
  const baseRate = mockRates[base] / (mockRates[quote] || 1);
  
  for (let i = 12; i <= 30; i += 2) {
    data.push({ date: `${i} мая`, rate: +(baseRate * (1 + (Math.random() * 0.04 - 0.02))).toFixed(4) });
  }
  for (let i = 1; i <= 9; i += 2) {
    data.push({ date: `${i} июня`, rate: +(baseRate * (1 + (Math.random() * 0.04 - 0.02))).toFixed(4) });
  }
  return data;
}