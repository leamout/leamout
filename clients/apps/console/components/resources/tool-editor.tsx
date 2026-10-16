"use client";

import { Button } from "@leamout/ui/components/button";
import { Input } from "@leamout/ui/components/input";
import { Textarea } from "@leamout/ui/components/textarea";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { type FormEvent, useEffect, useState } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api,errorMessage } from "@/lib/api";
import { resourceKeys,useAgents,useAllTools,type AgentTool } from "@/lib/resources";
import { RequestError,RequestLoading } from "./request-state";

const builtinNames=["hangup_call","transfer_call","hold_call","resume_call","send_dtmf"];

export function ToolEditor({id}:{id?:string}) {
  const {selected}=useOrganization();
  const orgId=selected?.id??"";
  const agents=useAgents();
  const tools=useAllTools();
  const tool=id?tools.data?.find(item=>item.id===id):undefined;
  const client=useQueryClient();
  const router=useRouter();
  const [agentId,setAgentId]=useState("");
  const [type,setType]=useState<"builtin"|"webhook">("webhook");
  const [name,setName]=useState("");
  const [description,setDescription]=useState("");
  const [parameters,setParameters]=useState('{"type":"object","properties":{}}');
  const [endpointUrl,setEndpointUrl]=useState("");
  const [timeoutMs,setTimeoutMs]=useState("10000");
  const [enabled,setEnabled]=useState(false);
  const [secret,setSecret]=useState("");
  const [createdId,setCreatedId]=useState("");
  const [validation,setValidation]=useState("");
  useEffect(()=>{
    if(!tool)return;
    setAgentId(tool.voice_agent_id);
    setType(tool.type);
    setName(tool.name);
    setDescription(tool.description);
    setParameters(JSON.stringify(tool.parameters,null,2));
    setEndpointUrl(tool.endpoint_url??"");
    setTimeoutMs(String(tool.timeout_ms));
    setEnabled(tool.enabled);
  },[tool]);
  const owner=tool?.voice_agent_id??agentId;
  const invalidate=async()=>{
    await client.invalidateQueries({queryKey:resourceKeys.tools(orgId)});
    if(owner)await client.invalidateQueries({queryKey:resourceKeys.agentTools(orgId,owner)});
  };
  const save=useMutation({mutationFn:()=>api<AgentTool>(id?`/voice-agents/${owner}/tools/${id}`:`/voice-agents/${agentId}/tools`,{
    method:id?"PATCH":"POST",organizationId:orgId,
    body:{...(!id?{type}:{}),name:name.trim(),description:description.trim(),parameters:JSON.parse(parameters),
      ...(type==="webhook"?{endpoint_url:endpointUrl.trim()}:{}),timeout_ms:Number(timeoutMs),enabled},
  }),onSuccess:async(value)=>{
    await invalidate();
    if(!id && value.signing_secret){
      setSecret(value.signing_secret);
      setCreatedId(value.id);
    } else if(!id) {
      router.push(`/tools/${value.id}`);
    }
  }});
  const remove=useMutation({mutationFn:()=>api<void>(`/voice-agents/${owner}/tools/${id}`,{method:"DELETE",organizationId:orgId}),onSuccess:async()=>{await invalidate();router.push("/tools");}});
  if(agents.isPending||(id&&tools.isPending))return <RequestLoading what="agent tools"/>;
  if(agents.isError)return <RequestError error={agents.error} retry={()=>void agents.refetch()}/>;
  if(id&&tools.isError)return <RequestError error={tools.error} retry={()=>void tools.refetch()}/>;
  if(id&&!tool)return <div role="alert">Tool not found in this organization.</div>;
  function submit(event:FormEvent<HTMLFormElement>){
    event.preventDefault();setValidation("");
    if(!owner||!name.trim()||!description.trim())return setValidation("Complete the required fields.");
    try{const v=JSON.parse(parameters);if(!v||typeof v!=="object"||Array.isArray(v))throw Error();}catch{return setValidation("Parameters must be a JSON object.");}
    if(type==="webhook"&&!/^https:\/\//i.test(endpointUrl.trim()))return setValidation("A secure HTTPS URL is required.");
    if(!Number.isInteger(Number(timeoutMs))||Number(timeoutMs)<100||Number(timeoutMs)>30000)return setValidation("Timeout must be between 100 and 30000 ms.");
    save.mutate();
  }
  return <form onSubmit={submit} className="max-w-2xl space-y-5">
    <Link href="/tools" className="text-sm underline">Back to tools</Link>
    <h1 className="text-3xl font-semibold">{id?"Tool configuration":"Create tool"}</h1>
    <label className="block space-y-2 text-sm font-medium">Agent<select required disabled={Boolean(id)} className="h-10 w-full rounded-md border bg-background px-3" value={owner} onChange={e=>setAgentId(e.target.value)}><option value="">Select agent</option>{agents.data?.map(a=><option key={a.id} value={a.id}>{a.name}</option>)}</select></label>
    <label className="block space-y-2 text-sm font-medium">Tool type<select disabled={Boolean(id)} className="h-10 w-full rounded-md border bg-background px-3" value={type} onChange={e=>setType(e.target.value as typeof type)}><option value="webhook">Webhook</option><option value="builtin">Built-in</option></select></label>
    <label className="block space-y-2 text-sm font-medium">Name{type==="builtin"?<select required className="h-10 w-full rounded-md border bg-background px-3" value={name} onChange={e=>setName(e.target.value)}><option value="">Choose an action</option>{builtinNames.map(v=><option key={v} value={v}>{v}</option>)}</select>:<Input required maxLength={128} value={name} onChange={e=>setName(e.target.value)}/>}</label>
    <label className="block space-y-2 text-sm font-medium">Description<Textarea required maxLength={2000} value={description} onChange={e=>setDescription(e.target.value)}/></label>
    <label className="block space-y-2 text-sm font-medium">Parameters (JSON)<Textarea required rows={6} className="font-mono" value={parameters} onChange={e=>setParameters(e.target.value)}/></label>
    {type==="webhook"&&<label className="block space-y-2 text-sm font-medium">Webhook URL<Input type="url" required value={endpointUrl} onChange={e=>setEndpointUrl(e.target.value)}/></label>}
    <label className="block space-y-2 text-sm font-medium">Timeout (ms)<Input required type="number" min={100} max={30000} value={timeoutMs} onChange={e=>setTimeoutMs(e.target.value)}/></label>
    <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={enabled} onChange={e=>setEnabled(e.target.checked)}/> Enabled</label>
    {validation&&<p role="alert" className="text-sm text-destructive">{validation}</p>}
    {save.isError&&<p role="alert" className="text-sm text-destructive">{errorMessage(save.error)}</p>}
    {remove.isError&&<p role="alert" className="text-sm text-destructive">{errorMessage(remove.error)}</p>}
    <div className="flex gap-3"><Button disabled={save.isPending}>{id?"Save tool":"Create tool"}</Button>{id&&<Button type="button" variant="destructive" disabled={remove.isPending} onClick={()=>{if(window.confirm("Delete this tool?"))remove.mutate();}}>Delete tool</Button>}</div>
    {secret&&<div role="status" className="space-y-3 rounded-lg border p-4">
      <p className="text-sm font-medium">Signing secret (copy now; shown once)</p>
      <code className="block break-all text-xs">{secret}</code>
      <div className="flex gap-2">
        <Button type="button" variant="outline" onClick={()=>void navigator.clipboard.writeText(secret)}>Copy secret</Button>
        <Button type="button" onClick={()=>router.push(`/tools/${createdId}`)}>View tool</Button>
      </div>
    </div>}
  </form>;
}
