import type { MDXComponents } from "mdx/types";

export function useMDXComponents(components: MDXComponents): MDXComponents {
  return {
    h1: (props) => (
      <h1
        {...props}
        className="font-heading text-4xl font-semibold tracking-tight text-foreground sm:text-5xl"
      />
    ),
    h2: (props) => (
      <h2
        {...props}
        className="mt-12 mb-5 font-heading text-2xl font-semibold tracking-tight text-foreground"
      />
    ),
    h3: (props) => (
      <h3
        {...props}
        className="mt-8 mb-3 font-heading text-xl font-semibold text-foreground"
      />
    ),
    p: (props) => (
      <p
        {...props}
        className="my-5 leading-8 text-muted-foreground"
      />
    ),
    a: (props) => (
      <a
        {...props}
        className="text-brand-orange underline underline-offset-4 hover:no-underline"
      />
    ),
    ul: (props) => <ul {...props} className="my-5 list-disc space-y-2 pl-6" />,
    ol: (props) => <ol {...props} className="my-5 list-decimal space-y-2 pl-6" />,
    li: (props) => <li {...props} className="leading-7 text-muted-foreground" />,
    blockquote: (props) => (
      <blockquote
        {...props}
        className="my-8 border-l-2 border-brand-orange pl-5 text-muted-foreground"
      />
    ),
    pre: (props) => (
      <pre
        {...props}
        className="my-6 overflow-x-auto rounded-lg border border-border bg-card p-5 font-mono text-sm leading-7 [&_code]:bg-transparent [&_code]:p-0"
      />
    ),
    code: (props) => (
      <code
        {...props}
        className="rounded bg-card px-1.5 py-0.5 font-mono text-sm text-foreground"
      />
    ),
    ...components,
  };
}
