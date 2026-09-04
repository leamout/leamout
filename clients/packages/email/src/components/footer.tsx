import { Link, Section, Text } from "react-email";

export function EmailFooter() {
  return (
    <Section className="border-t border-stroke px-6 py-10 sm:px-10 sm:py-12">
      <Text className="m-0 max-w-[390px] font-sans text-[12px] leading-[18px] text-fg-3">
        Leamout is a self-hosted programmable communications control plane for
        voice and messaging infrastructure.
      </Text>
      <Text className="m-0 mt-5 font-sans text-[11px] leading-[18px] text-fg-3">
        4R59+MW, Akatsi South
        <br />
        Volta Region, Ghana
      </Text>
      <Text className="m-0 mt-2 font-sans text-[11px] leading-[18px] text-fg-3">
        <Link href="https://leamout.com" className="text-fg-2">
          leamout.com
        </Link>
      </Text>
    </Section>
  );
}
