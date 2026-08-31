import { Inviter, Registerer, SessionState, UserAgent } from "sip.js";

type AcceptanceConfig = {
  websocketUrl: string;
  sipUri: string;
  authorizationUsername: string;
  authorizationPassword: string;
  destinationUri: string;
  iceServers: RTCIceServer[];
};

declare global {
  interface Window {
    runLeamoutWebRTCAcceptance(config: AcceptanceConfig): Promise<void>;
  }
}

window.runLeamoutWebRTCAcceptance = async (config) => {
  const uri = UserAgent.makeURI(config.sipUri);
  const destination = UserAgent.makeURI(config.destinationUri);
  if (!uri || !destination) throw new Error("invalid SIP URI");

  const userAgent = new UserAgent({
    uri,
    authorizationUsername: config.authorizationUsername,
    authorizationPassword: config.authorizationPassword,
    transportOptions: { server: config.websocketUrl },
    sessionDescriptionHandlerFactoryOptions: {
      peerConnectionConfiguration: {
        iceServers: config.iceServers,
        iceTransportPolicy: "relay",
      },
    },
  });
  const registerer = new Registerer(userAgent);

  try {
    await userAgent.start();
    await registerer.register();

    const inviter = new Inviter(userAgent, destination, {
      sessionDescriptionHandlerOptions: { constraints: { audio: true, video: false } },
    });
    await inviter.invite();
    await new Promise<void>((resolve, reject) => {
      const timer = setTimeout(() => reject(new Error("SIP call did not establish")), 30_000);
      inviter.stateChange.addListener((state) => {
        if (state === SessionState.Established) {
          clearTimeout(timer);
          resolve();
        } else if (state === SessionState.Terminated) {
          clearTimeout(timer);
          reject(new Error("SIP call terminated before establishment"));
        }
      });
    });

    const handler = inviter.sessionDescriptionHandler as unknown as { peerConnection: RTCPeerConnection };
    const peerConnection = handler.peerConnection;
    const remote = document.querySelector<HTMLAudioElement>("#remote");
    if (!remote) throw new Error("remote audio element unavailable");
    remote.srcObject = new MediaStream(peerConnection.getReceivers().flatMap((receiver) => receiver.track ? [receiver.track] : []));

    await new Promise((resolve) => setTimeout(resolve, 2_000));
    const stats = await peerConnection.getStats();
    const selectedPairs = [...stats.values()].filter((report) =>
      report.type === "candidate-pair" && report.state === "succeeded" && report.nominated,
    );
    if (selectedPairs.length !== 1) throw new Error(`expected one selected ICE pair, got ${selectedPairs.length}`);
    const localCandidate = stats.get(selectedPairs[0].localCandidateId);
    if (!localCandidate || localCandidate.candidateType !== "relay") {
      throw new Error(`expected forced TURN relay, got ${localCandidate?.candidateType ?? "none"}`);
    }

    const inboundAudio = [...stats.values()].find((report) =>
      report.type === "inbound-rtp" && report.kind === "audio" && report.bytesReceived > 0,
    );
    if (!inboundAudio) throw new Error("no inbound audio received through TURN/RTPengine");
    await inviter.bye();
  } finally {
    await registerer.unregister().catch(() => undefined);
    await userAgent.stop();
  }
};
