"use client";

import { Button } from "@leamout/ui/components/button";
import { Input } from "@leamout/ui/components/input";
import { Textarea } from "@leamout/ui/components/textarea";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import { resourceKeys, type VoiceAgent } from "@/lib/resources";
import { RequestError, RequestLoading } from "./request-state";

export function AgentEditor({ id }: { id?: string }) {
  const { selected } = useOrganization();
  const orgId = selected?.id ?? "";
  const router = useRouter();
  const client = useQueryClient();
  const detail = useQuery({
    queryKey: [...resourceKeys.agents(orgId), id],
    queryFn: ({ signal }) => api<VoiceAgent>(`/voice-agents/${id}`, {organizationId:orgId, signal}),
    enabled: Boolean(id && orgId),
  });
  if (id && detail.isPending) return <RequestLoading what="agent" />;
  if (id && detail.isError) return <RequestError error={detail.error} retry={() => void detail.refetch()} />;
  return <AgentFields key={id ?? "new"} id={id} existing={detail.data} orgId={orgId} onDone={async (agent) => {
    await client.invalidateQueries({queryKey:resourceKeys.agents(orgId)});
    router.push(`/agents/${agent.id}`);
  }} onDelete={async () => {
    await client.invalidateQueries({queryKey:resourceKeys.agents(orgId)});
    router.push("/agents");
  }} />;
}

function AgentFields({id,existing,orgId,onDone,onDelete}:{id?:string;existing?:VoiceAgent;orgId:string;onDone:(agent:VoiceAgent)=>void;onDelete:()=>void}) {
  const [name,setName]=useState(existing?.name ?? "");
  const [engine,setEngine]=useState<"realtime"|"composable">(existing?.engine ?? "realtime");
  const [instructions,setInstructions]=useState(existing?.instructions ?? "");
  const [language,setLanguage]=useState(existing?.language ?? "en");
  const [voice,setVoice]=useState(existing?.voice ?? "");
  const save = useMutation({
    mutationFn:()=>api<VoiceAgent>(id ? `/voice-agents/${id}` : "/voice-agents", {
      organizationId:orgId,method:id?"PATCH":"POST",
      body:id?{name:name.trim(),engine,instructions:instructions.trim(),language,voice:voice||undefined}:{name:name.trim(),engine,instructions:instructions.trim(),language,voice:voice||undefined,engine_config:{}},
    }),
    onSuccess:onDone,
  });
  const remove = useMutation({
    mutationFn:()=>api<void>(`/voice-agents/${id}`,{organizationId:orgId,method:"DELETE"}),
    onSuccess:onDelete,
  });
  return <form className="max-w-2xl space-y-5" onSubmit={event=>{event.preventDefault();if(name.trim()&&instructions.trim())save.mutate();}}>
    <Link href="/agents" className="text-sm underline">Back to agents</Link>
    <h1 className="text-3xl font-semibold">{id?"Agent configuration":"Create agent"}</h1>
    <label className="block space-y-2 text-sm font-medium">Name<Input required maxLength={255} value={name} onChange={e=>setName(e.target.value)}/></label>
    <label className="block space-y-2 text-sm font-medium">Engine<select value={engine} onChange={e=>setEngine(e.target.value as typeof engine)} className="h-10 w-full rounded-md border bg-background px-3"><option value="realtime">Realtime</option><option value="composable">Composable STT / LLM / TTS</option></select></label>
    <label className="block space-y-2 text-sm font-medium">Instructions<Textarea required rows={8} maxLength={20000} value={instructions} onChange={e=>setInstructions(e.target.value)}/></label>
    <label className="block space-y-2 text-sm font-medium">Language<Input maxLength={64} value={language} onChange={e=>setLanguage(e.target.value)}/></label>
    <label className="block space-y-2 text-sm font-medium">Voice (optional)<Input maxLength={255} value={voice} onChange={e=>setVoice(e.target.value)}/></label>
    <p className="text-sm text-muted-foreground">Provider bindings must be configured before this agent can handle calls.</p>
    {save.isError&&<p role="alert" className="text-sm text-destructive">{errorMessage(save.error)}</p>}
    {remove.isError&&<p role="alert" className="text-sm text-destructive">{errorMessage(remove.error)}</p>}
    <div className="flex gap-3"><Button type="submit" disabled={save.isPending||!name.trim()||!instructions.trim()}>{save.isPending?"Saving…":id?"Save agent":"Create agent"}</Button>
    {id&&<Button variant="destructive" type="button" disabled={remove.isPending} onClick={()=>{if(window.confirm("Disable this agent?"))remove.mutate();}}>Disable agent</Button>}</div>
  </form>;
}
