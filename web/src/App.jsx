import { useEffect, useState } from 'react';

function money(value) {
  return `$${Number(value).toFixed(2)}`;
}

export default function App({ address = '0xabc123' }) {
  const [portfolio, setPortfolio] = useState(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const controller = new AbortController();
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError('');
      setPortfolio(null);
      try {
        const response = await fetch(`/api/wallets/${address}/portfolio`, {
          signal: controller.signal
        });
        const body = await response.json().catch(() => null);
        if (!response.ok) {
          throw new Error(body?.error?.message || 'Unable to load portfolio');
        }
        if (!body || typeof body.address !== 'string' || !Array.isArray(body.tokens)) {
          throw new Error('Invalid portfolio response');
        }
        if (!cancelled) {
          setPortfolio(body);
        }
      } catch (err) {
        if (!cancelled && err.name !== 'AbortError') {
          setError(err.message);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    load();
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [address]);

  if (loading || (portfolio && portfolio.address !== address.trim().toLowerCase())) {
    return <main><p>Loading portfolio...</p></main>;
  }

  if (error) {
    return <main role="alert">{error}</main>;
  }

  if (!portfolio) {
    return <main><p>Loading portfolio...</p></main>;
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
              <td>{token.price_available ? money(token.price_usd) : 'Price unavailable'}</td>
              <td>{token.price_available ? money(token.value_usd) : 'Price unavailable'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </main>
  );
}
