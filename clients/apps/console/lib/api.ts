export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public code: string,
    public retryAfter: number | null = null,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

type Envelope<T> = {
  success: boolean;
  data?: T;
  error?: { code: string; message: string };
};

export async function api<T>(
  path: string,
  options: {
    method?: "GET" | "POST" | "DELETE";
    body?: unknown;
    signal?: AbortSignal;
  } = {},
): Promise<T> {
  const response = await fetch(`/api/v1${path}`, {
    method: options.method ?? "GET",
    credentials: "same-origin",
    cache: "no-store",
    signal: options.signal,
    headers:
      options.body === undefined
        ? undefined
        : { "Content-Type": "application/json" },
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
  });

  if (response.status === 204) return undefined as T;

  const envelope = (await response
    .json()
    .catch(() => null)) as Envelope<T> | null;
  if (!response.ok || !envelope?.success) {
    const retry = Number(response.headers.get("Retry-After"));
    throw new ApiError(
      envelope?.error?.message ??
        "Unable to complete the request. Please try again.",
      response.status,
      envelope?.error?.code ?? "REQUEST_FAILED",
      retry > 0 ? retry : null,
    );
  }

  return envelope.data as T;
}

export function errorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return error.status === 429
      ? `Too many attempts. ${error.retryAfter ? `Try again in ${error.retryAfter} seconds.` : "Please wait before trying again."}`
      : error.message;
  }

  return "Could not reach Leamout. Check your connection and try again.";
}
