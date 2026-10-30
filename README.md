## Task Overview
Keyrail is a non-custodial crypto wallet: it shows people what their tokens are worth without ever holding their funds. A Go service gathers each wallet's balances and current market prices, and a small web view turns them into the portfolio screen users check several times a day. Since the portfolio screen went live, users, the support team and the mobile team building on the same service have each reported results they cannot rely on, even though the service builds and its basic checks run. Users judge a wallet by whether its numbers can be trusted, so a portfolio screen that misleads them costs more than a cosmetic bug.

## Objectives
- Keep the portfolio timely and dependable for every wallet.
- Keep the service reliable when many people check their wallets at once.
- Make failures clear and consistent for every client of the service.
- Show users exactly where their portfolio stands at every moment.

## Helpful Tips
- Run the API and the view, and follow one wallet from the request to what appears on screen before you change anything.
- Think about what a user should see when the market data behind their wallet does not cooperate.
- Consider what changes when many people open their wallets at the same moment.
- Whatever you change should still be checkable by hand afterwards.

## How to Verify
- Load a wallet whose market data does not all come back promptly, and see what the response contains and how long it takes.
- Send a burst of requests for the same wallet at once, and look at how the service holds up.
- Ask for wallets the service cannot serve for different reasons, and compare what each client receives.
- Open the portfolio view while each of those responses comes back, and watch what the user sees from start to finish.
- The tests that ship with the service and the view exercise these behaviours.
