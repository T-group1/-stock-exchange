export const rates: Record<string, number> = {
    USD: 1,
    EUR: 0.92,
    RUB: 90,
};

export function convertCurrency(amount: number, from: string, to: string) {
    const inUSD = amount / rates[from];
    return inUSD * rates[to];
}