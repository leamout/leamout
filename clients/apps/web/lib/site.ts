export const consoleURL = (
  process.env.CONSOLE_URL ??
  (process.env.NODE_ENV === "development"
    ? "http://localhost:3001"
    : "https://console.leamout.com")
).replace(/\/+$/, "");
