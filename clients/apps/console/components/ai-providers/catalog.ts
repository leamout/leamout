export const providerCatalog = [
  { id: "openai", name: "OpenAI" },
  { id: "deepgram", name: "Deepgram" },
  { id: "assemblyai", name: "AssemblyAI" },
  { id: "groq", name: "Groq" },
  { id: "cartesia", name: "Cartesia" },
  { id: "elevenlabs", name: "ElevenLabs" },
];

export type ProviderAvailability = {
  providerId: string;
  capabilities?: string[];
  organizationCredentialCount?: number;
  platformAvailable?: boolean;
};

export type ProviderCredential = {
  id: string;
  name: string;
  connectionState: string;
  verifiedAt?: string;
};
