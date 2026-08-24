import os
import openai
import requests
import csv
from dotenv import load_dotenv

load_dotenv()

Polygon_api_key = os.getenv("polygon_api_key")

print(Polygon_api_key)

#limit=5
url=f"https://api.polygon.io/v3/reference/tickers?market=stocks&active=true&order=asc&limit=1000&sort=ticker&apiKey={Polygon_api_key}"



response = requests.get(url)
data=response.json()

# Debug: Print the initial response structure
print("Initial response keys:", data.keys())
print("Initial response status:", response.status_code)

# Desired CSV schema as per example
fieldnames = [
    "ticker",
    "name",
    "market",
    "locale",
    "primary_exchange",
    "type",
    "active",
    "currency",
    "last_updated",
]

rows = []

def append_results(page_data):
    if 'results' not in page_data:
        return
    for item in page_data['results']:
        row = {
            "ticker": item.get("ticker"),
            "name": item.get("name"),
            "market": item.get("market"),
            "locale": item.get("locale"),
            "primary_exchange": item.get("primary_exchange"),
            "type": item.get("type"),
            "active": item.get("active"),
            # Polygon may use different keys for these
            "currency": item.get("currency") or item.get("currency_name"),
            "last_updated": (
                item.get("last_updated")
                or item.get("last_updated_utc")
                or item.get("updated")
                or item.get("updated_utc")
            ),
        }
        rows.append(row)

# Append first page
append_results(data)

# Handle pagination
while 'next_url' in data and data['next_url']:
    print('requesting next page',data['next_url'])
    next_url = data['next_url']
    if 'apiKey=' not in next_url:
        separator = '&' if '?' in next_url else '?'
        next_url = f"{next_url}{separator}apiKey={Polygon_api_key}"

    response = requests.get(next_url)
    data=response.json()

    # Debug: Print the response structure
    print("Response keys:", data.keys())
    print("Response status:", response.status_code)

    if 'results' in data:
        append_results(data)
    else:
        print("No 'results' key found. Full response:", data)
        break

# Write CSV with the exact schema order
output_csv = "tickers.csv"
with open(output_csv, mode="w", newline="", encoding="utf-8") as f:
    writer = csv.DictWriter(f, fieldnames=fieldnames, extrasaction="ignore")
    writer.writeheader()
    for r in rows:
        # Ensure all keys exist; fill missing with None
        for k in fieldnames:
            if k not in r:
                r[k] = None
        writer.writerow(r)

print(f"Wrote {len(rows)} records to {output_csv}")

example={"ticker":"AAPL","name":"Apple Inc.","market":"stocks","locale":"us","primary_exchange":"NASDAQ","type":"cs","active":True,"currency":"USD","last_updated":"2025-09-18T16:00:00Z"}
