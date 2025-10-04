export const ORDER_SCHEMA_V1 = [
  "order_id",
  "customer_id",
  "product",
  "quantity",
  "unit_price",
  "order_date",
] as const;

export type OrderSchemaKey = (typeof ORDER_SCHEMA_V1)[number];
