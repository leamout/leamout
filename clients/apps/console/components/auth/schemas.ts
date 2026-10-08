import { z } from "zod";

export const emailSchema = z.object({
  email: z.string().trim().email("Enter a valid email address."),
});

export const loginPasswordSchema = z.object({
  password: z.string().min(1, "Enter your password."),
});

export const verificationSchema = z.object({
  code: z.string().regex(/^[0-9]{6}$/, "Enter the six-digit code."),
});

export const newPasswordSchema = z
  .object({
    password: z.string().min(1, "Enter a new password."),
    confirmation: z.string().min(1, "Confirm your password."),
  })
  .refine((values) => values.password === values.confirmation, {
    message: "The passwords do not match.",
    path: ["confirmation"],
  });
