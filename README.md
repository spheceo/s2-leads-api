# s2-leads-api

Google Maps lead scraper built in Go.

## Routes

### `GET /getIP`
Returns the current outbound IP and geo details from the configured proxy.

### `POST /search`
Fetches lead data for a business type in a city/country.

#### Headers
- `Authorization` (required): `Bearer <UNKEY_API_KEY>`

#### Request body
```json
{
  "business_type": "dentist",
  "city": "Los Angeles",
  "country_code": "us",
  "limit": 10
}
```

#### Body fields
- `business_type` (string, required): Search keyword/category, e.g. `dentist`, `plumber`, `restaurant`.
- `city` (string, required): City to search in.
- `country_code` (string, required): 2-letter country code, e.g. `us`, `za`, `gb`.
- `limit` (number, required): Number of results to request (1 to 500).

## ENV

Create a `.env` file in the project root with:

```env
RAPIDAPI_KEY=your_rapidapi_key
PROXY_URL=http://username:password@host:port
UNKEY_ROOT_KEY=your_unkey_root_key
```

- `RAPIDAPI_KEY` (required): API key used for the RapidAPI Google Maps data endpoint.
- `PROXY_URL` (required): Proxy URL used by outbound requests (for both geocoding and leads fetch).
- `UNKEY_ROOT_KEY` (required): Root key used to verify incoming API keys for protected routes.
