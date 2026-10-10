export type ResourceOption = { id: string; name: string };

export type PhoneNumber = {
  id: string;
  number: string;
  countryCode: string;
  status: string;
  voiceEnabled: boolean;
  trunk?: ResourceOption;
  agent?: ResourceOption;
};
