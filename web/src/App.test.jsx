import { act, render, screen } from '@testing-library/react';
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
    expect(screen.getAllByText(/price unavailable/i)).toHaveLength(2);
  });

  test('does not replace the current wallet with a stale response', async () => {
    let resolveFirst;
    let resolveSecond;
    global.fetch = vi.fn()
      .mockImplementationOnce(() => new Promise((resolve) => { resolveFirst = resolve; }))
      .mockImplementationOnce(() => new Promise((resolve) => { resolveSecond = resolve; }));

    const { rerender } = render(<App address="0xfirst" />);
    rerender(<App address="0xsecond" />);

    await act(async () => {
      resolveSecond({
        ok: true,
        json: async () => ({ address: '0xsecond', tokens: [{ symbol: 'NEW', chain: 'test', amount: 1, price_available: false }], total_usd: 0 })
      });
    });
    await act(async () => {
      resolveFirst({
        ok: true,
        json: async () => ({ address: '0xfirst', tokens: [{ symbol: 'OLD', chain: 'test', amount: 1, price_available: false }], total_usd: 0 })
      });
    });

    expect(await screen.findByText('NEW')).toBeInTheDocument();
    expect(screen.queryByText('OLD')).not.toBeInTheDocument();
  });

  test('starts a new wallet request in the loading state and clears previous errors', async () => {
    mockFetch({
      ok: false,
      status: 404,
      json: async () => ({ error: { code: 'wallet_not_found', message: 'Wallet not found' } })
    });
    const { rerender } = render(<App address="0xmissing" />);
    expect(await screen.findByRole('alert')).toHaveTextContent('Wallet not found');

    global.fetch = vi.fn(() => new Promise(() => {}));
    rerender(<App address="0xabc123" />);

    expect(await screen.findByText(/loading portfolio/i)).toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });
});
