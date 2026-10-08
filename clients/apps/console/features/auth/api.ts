type Envelope<T> = {
  success: boolean;
  data?: T;
  error?: { code: string; message: string };
};

export async function authRequest<T>(path: string, body: object): Promise<T> {
  const response = await fetch(`/api/v1/auth${path}`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const result = (await response.json()) as Envelope<T>;
  if (!response.ok || !result.success) {
    throw new Error(
      result.error?.message ?? "Unable to continue. Please try again.",
    );
  }
  return result.data as T;
}
