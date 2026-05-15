from __future__ import annotations

import argparse
import json
import os
from datetime import datetime, timezone
from typing import Any, Dict, List

import pandas as pd
import psycopg


DEFAULT_DATABASE_URL = "postgres://aggregator:secret@localhost:5432/aggregator?sslmode=disable"


def read_table(connection: Any, table: str) -> pd.DataFrame:
    return pd.read_sql_query(f"SELECT * FROM {table}", connection)


def money(value: float) -> float:
    return round(float(value or 0), 2)


def status_counts(orders: pd.DataFrame) -> Dict[str, int]:
    if orders.empty or "status" not in orders:
        return {}
    return orders["status"].fillna("unknown").value_counts().sort_index().astype(int).to_dict()


def calculate_summary(orders: pd.DataFrame, customers: pd.DataFrame, operators: pd.DataFrame) -> Dict[str, Any]:
    if orders.empty:
        return {
            "generated_at": datetime.now(timezone.utc).isoformat(),
            "economic": {"gmv": 0, "average_order_value": 0, "platform_revenue": 0},
            "operational": {
                "total_orders": 0,
                "active_customers": 0,
                "active_operators": 0,
                "search_to_contract_conversion": 0,
                "orders_by_status": {},
            },
            "security": {"disputes": 0, "fraud_alerts": 0, "reliability_by_operator": []},
            "technical": {"peak_load_by_hour": [], "flight_heatmap_points": []},
            "anomalies": [],
            "forecast": {"next_hour_orders": 0},
        }

    orders = orders.copy()
    orders["budget"] = pd.to_numeric(orders.get("budget", 0), errors="coerce").fillna(0)
    orders["commission_amount"] = pd.to_numeric(orders.get("commission_amount", 0), errors="coerce").fillna(0)
    orders["created_at"] = pd.to_datetime(orders.get("created_at"), errors="coerce", utc=True)

    completed_mask = orders["status"].isin(["confirmed", "completed_pending_confirmation", "completed"])
    started_mask = orders["status"].isin(["searching", "matched", "confirmed", "completed_pending_confirmation", "completed", "dispute"])
    contracted_mask = orders["status"].isin(["confirmed", "completed_pending_confirmation", "completed"])
    conversion = float(contracted_mask.sum() / started_mask.sum()) if started_mask.sum() else 0

    hourly = (
        orders.dropna(subset=["created_at"])
        .assign(hour=lambda df: df["created_at"].dt.floor("h").dt.strftime("%Y-%m-%dT%H:00:00Z"))
        .groupby("hour")
        .size()
        .reset_index(name="orders")
        .sort_values("hour")
    )

    heatmap_columns = ["from_lat", "from_lon", "to_lat", "to_lon"]
    heatmap_points: List[Dict[str, Any]] = []
    if all(column in orders.columns for column in heatmap_columns):
        coords = orders[["id", *heatmap_columns]].copy()
        for column in heatmap_columns:
            coords[column] = pd.to_numeric(coords[column], errors="coerce")
        coords = coords.dropna(subset=heatmap_columns)
        heatmap_points = [
            {"order_id": row.id, "lat": row.from_lat, "lon": row.from_lon, "type": "from"}
            for row in coords.itertuples()
        ] + [
            {"order_id": row.id, "lat": row.to_lat, "lon": row.to_lon, "type": "to"}
            for row in coords.itertuples()
        ]

    anomalies = detect_anomalies(orders)

    reliability = []
    if "operator_id" in orders.columns:
        operator_orders = orders[orders["operator_id"].fillna("") != ""].copy()
        if not operator_orders.empty:
            grouped = operator_orders.groupby("operator_id")
            reliability = (
                grouped.agg(
                    total_orders=("id", "count"),
                    completed_orders=("status", lambda s: int(s.isin(["completed", "confirmed"]).sum())),
                    disputes=("status", lambda s: int((s == "dispute").sum())),
                )
                .assign(reliability_score=lambda df: ((df["completed_orders"] - df["disputes"]) / df["total_orders"]).clip(lower=0))
                .reset_index()
                .sort_values(["reliability_score", "total_orders"], ascending=[False, False])
                .to_dict(orient="records")
            )

    return {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "economic": {
            "gmv": money(orders["budget"].sum()),
            "average_order_value": money(orders["budget"].mean()),
            "platform_revenue": money(orders["commission_amount"].sum()),
        },
        "operational": {
            "total_orders": int(len(orders)),
            "active_customers": int(customers["id"].nunique()) if not customers.empty else int(orders["customer_id"].nunique()),
            "active_operators": int(operators["id"].nunique()) if not operators.empty else int(orders.get("operator_id", pd.Series(dtype=str)).nunique()),
            "search_to_contract_conversion": round(conversion, 4),
            "completed_or_confirmed_orders": int(completed_mask.sum()),
            "orders_by_status": status_counts(orders),
        },
        "security": {
            "disputes": int((orders["status"] == "dispute").sum()),
            "fraud_alerts": len(anomalies),
            "reliability_by_operator": reliability,
        },
        "technical": {
            "peak_load_by_hour": hourly.tail(24).to_dict(orient="records"),
            "flight_heatmap_points": heatmap_points[:200],
        },
        "anomalies": anomalies,
        "forecast": {
            "next_hour_orders": int(round(hourly["orders"].tail(6).mean())) if not hourly.empty else 0,
        },
    }


def detect_anomalies(orders: pd.DataFrame) -> List[Dict[str, Any]]:
    anomalies: List[Dict[str, Any]] = []
    if orders.empty:
        return anomalies

    if "budget" in orders:
        q1 = orders["budget"].quantile(0.25)
        q3 = orders["budget"].quantile(0.75)
        iqr = q3 - q1
        high_budget_limit = q3 + 1.5 * iqr if iqr > 0 else orders["budget"].max() * 2

        high_budget_orders = orders[orders["budget"] > high_budget_limit]
        for row in high_budget_orders[["id", "budget"]].itertuples():
            anomalies.append({"type": "high_budget", "order_id": row.id, "value": money(row.budget)})

        invalid_budget_orders = orders[orders["budget"] <= 0]
        for row in invalid_budget_orders[["id", "budget"]].itertuples():
            anomalies.append({"type": "invalid_budget", "order_id": row.id, "value": money(row.budget)})

    if "created_at" in orders:
        now = pd.Timestamp.now(tz="UTC")
        stale = orders[(orders["status"] == "searching") & (orders["created_at"].notna()) & ((now - orders["created_at"]) > pd.Timedelta(hours=24))]
        for row in stale[["id", "created_at"]].itertuples():
            anomalies.append({"type": "stale_searching_order", "order_id": row.id, "created_at": row.created_at.isoformat()})

    return anomalies


def build_report(database_url: str) -> Dict[str, Any]:
    connection = psycopg.connect(database_url, autocommit=True)
    connection.execute("SET default_transaction_read_only = on")
    try:
        orders = read_table(connection, "orders")
        customers = read_table(connection, "customers")
        operators = read_table(connection, "operators")
        return calculate_summary(orders, customers, operators)
    finally:
        connection.close()


def main() -> None:
    parser = argparse.ArgumentParser(description="Build read-only analytics report as JSON.")
    parser.add_argument("--database-url", default=os.getenv("ANALYTICS_DATABASE_URL") or os.getenv("DATABASE_URL") or DEFAULT_DATABASE_URL)
    parser.add_argument("--output", help="Optional path for JSON output. Stdout is used by default.")
    args = parser.parse_args()

    report = build_report(args.database_url)
    data = json.dumps(report, ensure_ascii=False, indent=2)
    if args.output:
        with open(args.output, "w", encoding="utf-8") as file:
            file.write(data + "\n")
    else:
        print(data)


if __name__ == "__main__":
    main()
