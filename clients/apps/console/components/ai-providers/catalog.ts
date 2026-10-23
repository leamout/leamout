import type { ProviderRole } from "./types";

// Display metadata for the providers registered in server/internal/ai/catalog/builtins.go.
// Compatibility remains enforced by the Go provider registry and its validators.
export const providerCatalog: {
  id: string;
  name: string;
  roles: ProviderRole[];
}[] = [
  { id: "openai", name: "OpenAI", roles: ["llm", "realtime"] },
  { id: "gemini", name: "Gemini", roles: ["realtime"] },
  { id: "deepgram", name: "Deepgram", roles: ["stt"] },
  { id: "assemblyai", name: "AssemblyAI", roles: ["stt"] },
  { id: "groq", name: "Groq", roles: ["llm"] },
  { id: "cartesia", name: "Cartesia", roles: ["tts"] },
  { id: "elevenlabs", name: "ElevenLabs", roles: ["tts"] },
];

export function providerName(id: string): string {
  return providerCatalog.find((provider) => provider.id === id)?.name ?? id;
}
