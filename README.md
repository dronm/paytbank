# paytbank

A small Go client for the T-Bank pay API, intended for use in an e-commerce backend.

The package provides a lightweight transport layer around T-Bank payment methods and currently supports:

- `Init` — create a payment;
- `Cancel` — cancel a payment;
- `GetState` — get payment status.

API reference:
- T-Bank E-Acquiring API
- Non-PCI сценарий оплаты

## Features

- JSON `POST` requests to T-Bank API;
- context-aware HTTP calls;
- injectable HTTP client;
- configurable API base URL;
- package-level `DefaultClient` for simple usage;
- easy testing through `httptest.Server`.

## Installation

```bash
go get github.com/dronm/paytbank
