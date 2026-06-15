import Header from "../components/Header";
import CurrencyConverter from "../components/CurrencyConverter";
import FavoritesList from "../components/FavoritesList";
import { useEffect, useState } from "react";
import { fetchRates } from "../Api/currencyApi";

export default function HomePage({ user, favorites, setFavorites }: any) {
  const [rates, setRates] = useState<Record<string, number>>({});
  const [ , setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      setLoading(true);
      const data = await fetchRates();
      setRates(data);
      setLoading(false);
    }
    load();
  }, []);

  return (
    <div style={{ fontFamily: "sans-serif", padding: "0 20px" }}>
      <Header user={user} />
      
      <div style={{ maxWidth: 800, margin: "20px auto" }}>
      
        <CurrencyConverter 
          rates={rates} 
          favorites={favorites} 
          setFavorites={setFavorites} 
          user={user}
        />

        <FavoritesList favorites={favorites} setFavorites={setFavorites} />
      </div>
    </div>
  );
}