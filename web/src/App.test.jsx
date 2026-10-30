import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest';
import App from './App.jsx';

function mockFetch(response) {
  global.fetch = vi.fn().mockResolvedValue(response);
}

describe('App', () => {
  beforeEach(() => {
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  test('shows a loading state while the portfolio request is pending', () => {
    global.fetch = vi.fn(() => new Promise(() => {}));
    render(<App address="0xabc123" />);
    expect(screen.getByText(/loading portfolio/i)).toBeInTheDocument();
  });

  test('shows the API error message', async () => {
    mockFetch({
      ok: false,
      status: 404,
      json: async () => ({ error: { code: 'wallet_not_found', message: 'Wallet not found' } })
    });

    render(<App address="0xmissing" />);

    expect(await screen.findByRole('alert')).toHaveTextContent('Wallet not found');
  });

  test('shows an empty state for wallets with no tokens', async () => {
    mockFetch({
      ok: true,
      status: 200,
      json: async () => ({ address: '0xempty', tokens: [], total_usd: 0 })
    });

    render(<App address="0xempty" />);

    expect(await screen.findByText(/no tokens found/i)).toBeInTheDocument();
  });

  test('labels unavailable token prices', async () => {
    mockFetch({
      ok: true,
      status: 200,
      json: async () => ({
        address: '0xabc123',
        total_usd: 3200,
        tokens: [
          { symbol: 'ETH', chain: 'ethereum', amount: 1, price_usd: 3200, value_usd: 3200, price_available: true },
          { symbol: 'KEY', chain: 'base', amount: 500, price_available: false }
        ]
      })
    });

    render(<App address="0xabc123" />);

    expect(await screen.findByText('ETH')).toBeInTheDocument();
    expect(screen.getByText(/price unavailable/i)).toBeInTheDocument();
  });
});
