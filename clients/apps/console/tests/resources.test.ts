import assert from "node:assert/strict";
import { test } from "node:test";
import { api } from "../lib/api";
import {
  listAgents,
  listTrunks,
  listIntegrations,
  listAgentTools,
  listAllTools,
} from "../lib/resources";

test("resource listing unwraps backend envelopes and sends organization scope", async () => {
  const originalFetch = globalThis.fetch;
  const calls: { url: string; org: string | null }[] = [];
  globalThis.fetch = async (url, init) => {
    const path = String(url);
    calls.push({ url: path, org: new Headers(init?.headers).get("X-Organization-ID") });
    const data = path.includes("/tools") ? { tools: [{id:"t1",voice_agent_id:"a1",name:"lookup"}] }
      : path.includes("/voice-agents") ? {voice_agents:[{id:"a1",name:"Agent One"}]}
      : path.includes("/trunks") ? {trunks:[{id:"tr1"}]}
      : {ai_integrations:[{id:"p1"}]};
    return Response.json({success:true,data});
  };
  try {
    assert.equal((await listAgents("org-one")).length,1);
    assert.equal((await listTrunks("org-one")).length,1);
    assert.equal((await listIntegrations("org-one")).length,1);
    assert.equal((await listAgentTools("org-one","a1")).length,1);
    assert.equal((await listAllTools("org-one"))[0].agent_name,"Agent One");
    assert.deepEqual(calls.map(c=>c.org),Array(6).fill("org-one"));
  } finally {
    globalThis.fetch=originalFetch;
  }
});

test("resource mutations use scoped POST PATCH DELETE and never add bearer credentials", async () => {
  const originalFetch=globalThis.fetch;
  const calls:RequestInit[]=[];
  globalThis.fetch=async (_url,init)=>{
    calls.push(init??{});
    return Response.json({success:true,data:{id:"one"}});
  };
  try {
    await api("/voice-agents",{method:"POST",organizationId:"org-one",body:{name:"Agent"}});
    await api("/trunks/one",{method:"PATCH",organizationId:"org-one",body:{name:"Trunk"}});
    await api("/ai-integrations/one/rotate",{method:"POST",organizationId:"org-one",body:{secret:"example"}});
    await api("/voice-agents/a1/tools/one",{method:"DELETE",organizationId:"org-one"});
    assert.deepEqual(calls.map(c=>c.method),["POST","PATCH","POST","DELETE"]);
    for(const call of calls){
      const headers=new Headers(call.headers);
      assert.equal(headers.get("X-Organization-ID"),"org-one");
      assert.equal(headers.get("Authorization"),null);
      assert.equal(call.credentials,"same-origin");
    }
  } finally {globalThis.fetch=originalFetch;}
});
