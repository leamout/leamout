// Public fixed-price plans mirrored from server/migrations/033_create_plans.sql.
// Update this catalog alongside the backend plan seed when offerings change.
export const plans = [
  {
    code: "free",
    name: "Free",
    price: 0,
    minutes: 100,
    concurrency: 1,
    agents: 1,
    trunks: 1,
    retention: 1,
  },
  {
    code: "developer",
    name: "Developer",
    price: 79,
    minutes: 2000,
    concurrency: 3,
    agents: 5,
    trunks: 3,
    retention: 14,
  },
  {
    code: "pro",
    name: "Pro",
    price: 299,
    minutes: 10000,
    concurrency: 15,
    agents: 25,
    trunks: 10,
    retention: 30,
  },
  {
    code: "scale",
    name: "Scale",
    price: 799,
    minutes: 30000,
    concurrency: 40,
    agents: 100,
    trunks: 30,
    retention: 90,
  },
] as const;

export type PublicPlan = (typeof plans)[number];
