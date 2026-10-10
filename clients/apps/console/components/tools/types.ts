export type Tool = {
  id: string;
  name: string;
  description: string;
  type: "builtin" | "webhook";
  parameters: string;
  endpointUrl: string;
  timeoutMs: number;
  enabled: boolean;
  agent: { id: string; name: string };
};
