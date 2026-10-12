import { z } from "zod";

export const flowKey = "leamout.auth-flow";
export const flowSchema = z.object({
  transactionId: z.string().uuid(),
  email: z.string().email(),
  intent: z.enum(["log-in", "create-account", "recovery"]),
  stage: z.enum(["password", "verification", "enroll", "reset"]),
  expiresAt: z.number(),
  resendAt: z.number(),
});
export type AuthFlow = z.infer<typeof flowSchema>;

export function restoreFlow(
  value: string | null,
  now = Date.now(),
): AuthFlow | null {
  try {
    const parsed = flowSchema.safeParse(JSON.parse(value ?? "null"));
    return parsed.success && parsed.data.expiresAt > now ? parsed.data : null;
  } catch {
    return null;
  }
}
