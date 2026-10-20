"use client";

import { ArrowUpRight01Icon, Tick02Icon } from "@hugeicons/core-free-icons";
import { HugeiconsIcon } from "@hugeicons/react";
import { Badge } from "@leamout/ui/components/badge";
import { Button } from "@leamout/ui/components/button";
import { Card, CardContent } from "@leamout/ui/components/card";
import { Separator } from "@leamout/ui/components/separator";
import { cn } from "@leamout/ui/lib/utils";
import { motion, useReducedMotion } from "motion/react";
import { plans } from "@/lib/plans";

const presentation = {
  Free: {
    background: "bg-card",
    description:
      "Build and test your first voice agent with your own providers.",
  },
  Developer: {
    background: "bg-card",
    description:
      "For developers bringing their first voice applications into production.",
  },
  Pro: {
    background: "bg-brand-orange/10",
    description: "More capacity and controls for growing production teams.",
  },
  Scale: {
    background: "bg-card",
    description:
      "Room for higher-volume workloads and a larger fleet of agents.",
  },
};

export default function Pricing({ signupHref }: { signupHref: string }) {
  const reducedMotion = useReducedMotion();

  return (
    <section
      className="bg-background py-12 sm:py-16 lg:py-20"
      aria-labelledby="pricing-title"
    >
      <div className="mx-auto max-w-7xl px-5 sm:px-8 lg:px-12">
        <div className="mb-10 flex flex-col items-center gap-4 text-center sm:mb-14">
          <Badge
            variant="outline"
            className="h-7 w-fit px-3 py-1 text-sm font-normal"
          >
            Pricing
          </Badge>
          <h1
            id="pricing-title"
            className="max-w-2xl text-balance text-4xl font-medium leading-tight tracking-tight sm:text-5xl"
          >
            The right plan for your next conversation.
          </h1>
          <p className="max-w-xl text-pretty text-base leading-7 text-muted-foreground">
            One platform subscription. Bring your own carrier and AI providers,
            and pay them directly for usage.
          </p>
        </div>

        <div className="grid gap-6 xl:grid-cols-2">
          {plans.map((plan, index) => {
            const style = presentation[plan.name as keyof typeof presentation];
            const features = [
              `${plan.minutes.toLocaleString("en-US")} runtime minutes / month`,
              `${plan.concurrency} concurrent ${plan.concurrency === 1 ? "call" : "calls"}`,
              `${plan.agents} voice ${plan.agents === 1 ? "agent" : "agents"}`,
              `${plan.trunks} SIP ${plan.trunks === 1 ? "trunk" : "trunks"}`,
              `${plan.retention}-day retention`,
              ...(plan.name === "Pro" || plan.name === "Scale"
                ? ["Advanced roles", "Configurable retention policies"]
                : ["Bring your own carrier & AI"]),
            ];

            return (
              <motion.div
                key={plan.name}
                initial={false}
                whileInView={
                  reducedMotion ? undefined : { opacity: [0.6, 1], y: [24, 0] }
                }
                viewport={{ once: true, amount: 0.15 }}
                transition={{
                  delay: (index % 2) * 0.12,
                  duration: 0.5,
                  ease: "easeInOut",
                }}
                className="min-w-0"
              >
                <Card
                  className={cn(
                    style.background,
                    "h-full rounded-2xl border-0 p-7 ring-0 sm:p-10 xl:p-8",
                  )}
                >
                  <CardContent className="flex h-full flex-col items-stretch gap-7 px-0 sm:flex-row sm:gap-8">
                    <div className="flex min-w-0 flex-1 flex-col items-start justify-between gap-8">
                      <div className="space-y-3">
                        <h2>
                          <Badge className="h-7 rounded-full px-3 py-1 text-sm font-normal">
                            {plan.name}
                          </Badge>
                        </h2>
                        <p className="max-w-56 text-sm leading-6 text-muted-foreground">
                          {style.description}
                        </p>
                      </div>
                      <div className="space-y-5">
                        <p className="flex flex-wrap items-baseline gap-x-1">
                          <span className="text-5xl font-semibold tracking-tight text-card-foreground">
                            ${plan.price}
                          </span>
                          <span className="text-sm text-muted-foreground">
                            USD / month
                          </span>
                        </p>
                        <Button
                          nativeButton={false}
                          render={
                            <a
                              href={signupHref}
                              aria-label={`Get started with ${plan.name}`}
                            />
                          }
                          className="group relative h-12 w-fit overflow-hidden rounded-full bg-brand-orange py-1 ps-6 pe-14 text-sm font-medium text-[#101316] transition-all duration-300 hover:bg-brand-orange/90 hover:ps-14 hover:pe-6 hover:text-black focus-visible:ring-brand-orange/40 motion-reduce:transition-none motion-reduce:hover:ps-6 motion-reduce:hover:pe-14"
                        >
                          <span className="relative z-10">Get started</span>
                          <span
                            aria-hidden="true"
                            className="absolute right-1 flex size-10 items-center justify-center rounded-full bg-[#101316] text-[#F5F6F4] transition-all duration-300 group-hover:right-[calc(100%-44px)] group-hover:rotate-45 motion-reduce:transition-none motion-reduce:group-hover:right-1 motion-reduce:group-hover:rotate-0"
                          >
                            <HugeiconsIcon
                              icon={ArrowUpRight01Icon}
                              size={16}
                              strokeWidth={1.5}
                            />
                          </span>
                        </Button>
                      </div>
                    </div>
                    <Separator
                      orientation="vertical"
                      className="hidden bg-foreground/15 sm:block"
                    />
                    <Separator className="bg-foreground/15 sm:hidden" />
                    <div className="min-w-0 flex-1 space-y-5">
                      <h3 className="text-lg font-medium text-card-foreground">
                        What’s included
                      </h3>
                      <ul className="space-y-3">
                        {features.map((feature) => (
                          <li
                            key={feature}
                            className="flex items-start gap-3 text-sm leading-6 text-card-foreground"
                          >
                            <HugeiconsIcon
                              icon={Tick02Icon}
                              size={16}
                              strokeWidth={1.5}
                              className="mt-1 shrink-0"
                              aria-hidden="true"
                            />
                            <span>{feature}</span>
                          </li>
                        ))}
                      </ul>
                    </div>
                  </CardContent>
                </Card>
              </motion.div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
