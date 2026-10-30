import { useEffect, useState } from 'react';

function money(value) {
  return `$${Number(value).toFixed(2)}`;
}

export default function App({ address = '0xabc123' }) {
  const [portfolio, setPortfolio] = useState(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const response = await fetch(`/api/wallets/${address}/portfolio`);
        const body = await response.json();
        if (!response.ok) {
          throw new Error('Unable to load portfolio');
        }
        if (!cancelled) {
          setPortfolio(body);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err.message);
        }
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, [address]);

  if (loading) {
    return <main><p>Loading portfolio...</p></main>;
  }

  if (error) {
    return <main role="alert">{error}</main>;
  }

  if (!portfolio) {
    return <main />;
  }

  if (portfolio.tokens.length === 0) {
    return <main><p>No tokens found for this wallet.</p></main>;
  }

  return (
    <main>
      <h1>Wallet Portfolio</h1>
      <p>Total value: {money(portfolio.total_usd)}</p>
      <table>
        <thead>
          <tr>
            <th>Token</th>
            <th>Amount</th>
            <th>Price</th>
            <th>Value</th>
          </tr>
        </thead>
        <tbody>
          {portfolio.tokens.map((token) => (
            <tr key={`${token.chain}-${token.symbol}`}>
              <td>{token.symbol}</td>
              <td>{token.amount}</td>
              <td>{money(token.price_usd)}</td>
              <td>{money(token.value_usd)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </main>
  );
}
