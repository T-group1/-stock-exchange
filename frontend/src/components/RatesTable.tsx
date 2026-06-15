import { useNavigate } from "react-router-dom";

type Props = {
  rates: Record<string, number>;
};

const mockChanges: Record<string, string> = {
  USD: "+1.2%", EUR: "-0.4%", CNY: "+0.8%", GBP: "+1.5%", JPY: "-0.2%"
};

export default function RatesTable({ rates }: Props) {
  const navigate = useNavigate();

  return (
    <div className="card" style={{ maxWidth: 600, margin: "20px auto" }}>
      <h3 style={{ margin: "0 0 20px 0" }}>Курсы валют (к RUB)</h3>

      <table className="rates-table">
        <thead>
          <tr>
            <th>Валютная пара</th>
            <th>Курс</th>
            <th>Изменение</th>
          </tr>
        </thead>

        <tbody>
          {Object.entries(rates || {})
            .filter(([currency]) => currency !== "RUB") 
            .map(([currency, value]) => {
              const rateToRub = rates["RUB"] ? rates["RUB"] / value : value;
              const change = mockChanges[currency] || "+0.1%";
              const isPositive = change.startsWith("+");

              return (
                <tr 
                  key={currency} 
                  className="clickable-row"
                  onClick={() => navigate(`/currency/${currency}-RUB`)}
                >
                  <td><strong>{currency}/RUB</strong></td>
                  <td className="rate-bold">
                    {typeof rateToRub === "number" ? rateToRub.toFixed(2) : "—"}
                  </td>
                  <td className={isPositive ? "text-green" : "text-red"}>
                    {change}
                  </td>
                </tr>
              );
            })}
        </tbody>
      </table>
    </div>
  );
}