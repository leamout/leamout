import { PhoneNumberDetails } from "@/components/phone-numbers/phone-number-details";

export default async function Page({
  params,
}: PageProps<"/phone-numbers/[numberId]">) {
  const { numberId } = await params;

  return <PhoneNumberDetails numberId={numberId} />;
}
