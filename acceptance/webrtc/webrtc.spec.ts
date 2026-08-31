import { expect, test } from "@playwright/test";

const required = (name: string): string => {
  const value = process.env[name];
  if (!value) throw new Error(`${name} is required`);
  return value;
};

test("browser call uses a forced TURN relay through RTPengine", async ({ page, request }) => {
  const apiURL = required("LEAMOUT_API_URL");
  const token = required("LEAMOUT_API_TOKEN");
  const organizationID = required("LEAMOUT_ORGANIZATION_ID");
  const response = await request.post(`${apiURL}/v1/realtime/ice-credentials`, {
    headers: { Authorization: `Bearer ${token}`, "X-Organization-ID": organizationID },
  });
  expect(response.ok(), await response.text()).toBeTruthy();
  const payload = await response.json();
  const credentials = payload.data ?? payload;

  await page.goto("/");
  await page.evaluate(async ({ credentials, environment }) => {
    await window.runLeamoutWebRTCAcceptance({
      websocketUrl: environment.websocketUrl,
      sipUri: environment.sipUri,
      authorizationUsername: environment.username,
      authorizationPassword: environment.password,
      destinationUri: environment.destinationUri,
      iceServers: credentials.ice_servers,
    });
  }, {
    credentials,
    environment: {
      websocketUrl: required("LEAMOUT_WSS_URL"),
      sipUri: required("LEAMOUT_SIP_URI"),
      username: required("LEAMOUT_SIP_USERNAME"),
      password: required("LEAMOUT_SIP_PASSWORD"),
      destinationUri: process.env.LEAMOUT_DESTINATION_URI ?? "sip:9196@leamout.local",
    },
  });
});
