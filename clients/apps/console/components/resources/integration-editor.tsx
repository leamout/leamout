"use client";

import { Button } from "@leamout/ui/components/button";
import { Input } from "@leamout/ui/components/input";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import { resourceKeys, useIntegrations, type AiIntegration } from "@/lib/resources";
import { RequestError, RequestLoading } from "./request-state";

const supported = ["openai","gemini","deepgram","assemblyai","groq","cartesia","elevenlabs"];

export function IntegrationEditor({provider}:{provider:string}) {
  const {selected}=useOrganization();
  const orgId=selected?.id??"";
  const client=useQueryClient();
  const integrations=useIntegrations();
  const [name,setName]=useState("");
  const [secret,setSecret]=useState("");
  const [rotating,setRotating]=useState<string|null>(null);
  const [newSecret,setNewSecret]=useState("");
  const invalidate=()=>client.invalidateQueries({queryKey:resourceKeys.integrations(orgId)});
  const create=useMutation({mutationFn:()=>api<AiIntegration>("/ai-integrations",{method:"POST",organizationId:orgId,body:{provider,name:name.trim(),secret}}),onSuccess:async()=>{setName("");setSecret("");await invalidate();}});
  const verify=useMutation({mutationFn:(id:string)=>api(`/ai-integrations/${id}/verify`,{method:"POST",organizationId:orgId}),onSuccess:()=>void invalidate()});
  const rotate=useMutation({mutationFn:(id:string)=>api(`/ai-integrations/${id}/rotate`,{method:"POST",organizationId:orgId,body:{secret:newSecret}}),onSuccess:async()=>{setRotating(null);setNewSecret("");await invalidate();}});
  const remove=useMutation({mutationFn:(id:string)=>api<void>(`/ai-integrations/${id}`,{method:"DELETE",organizationId:orgId}),onSuccess:()=>void invalidate()});
  if(!supported.includes(provider))return <div className="space-y-4"><p>Unknown provider.</p><Link href="/ai-providers" className="underline">Back</Link></div>;
  const items=integrations.data?.filter(item=>item.provider===provider)??[];
  return <div className="max-w-2xl space-y-6">
    <Link href="/ai-providers" className="text-sm underline">Back to AI providers</Link>
    <h1 className="text-3xl font-semibold capitalize">{provider}</h1>
    <p className="text-sm text-muted-foreground">Organization-owned credentials are encrypted by Leamout and are not shown again.</p>
    {integrations.isPending?<RequestLoading what="credentials"/>:integrations.isError?<RequestError error={integrations.error} retry={()=>void integrations.refetch()}/>:<div className="space-y-3">
      {items.map(item=><section key={item.id} className="space-y-3 rounded-lg border p-4">
        <div className="flex justify-between gap-3"><div><h2 className="font-medium">{item.name}</h2><p className="text-sm text-muted-foreground">Connection: {item.connection_state}</p></div>
        <Button type="button" variant="outline" disabled={verify.isPending} onClick={()=>verify.mutate(item.id)}>Verify</Button></div>
        {rotating===item.id?<form className="flex flex-wrap gap-2" onSubmit={e=>{e.preventDefault();if(newSecret.trim())rotate.mutate(item.id);}}>
          <Input aria-label="Replacement API key" type="password" autoComplete="new-password" required value={newSecret} onChange={e=>setNewSecret(e.target.value)}/>
          <Button disabled={rotate.isPending}>Rotate secret</Button><Button type="button" variant="outline" onClick={()=>{setRotating(null);setNewSecret("");}}>Cancel</Button>
        </form>:<div className="flex gap-2"><Button variant="outline" onClick={()=>{setRotating(item.id);setNewSecret("");}}>Rotate key</Button><Button variant="destructive" disabled={remove.isPending} onClick={()=>{if(window.confirm("Delete this integration? Agents using it may stop working."))remove.mutate(item.id);}}>Delete</Button></div>}
      </section>)}
      {!items.length&&<p className="text-sm text-muted-foreground">No credentials for this provider.</p>}
    </div>}
    <form className="space-y-4 rounded-lg border p-5" onSubmit={e=>{e.preventDefault();if(name.trim()&&secret.trim())create.mutate();}}>
      <h2 className="font-semibold">Add credential</h2>
      <label className="block space-y-2 text-sm font-medium">Credential name<Input required value={name} onChange={e=>setName(e.target.value)}/></label>
      <label className="block space-y-2 text-sm font-medium">API key<Input required type="password" autoComplete="new-password" value={secret} onChange={e=>setSecret(e.target.value)}/></label>
      {create.isError&&<p role="alert" className="text-sm text-destructive">{errorMessage(create.error)}</p>}
      {(verify.isError||rotate.isError||remove.isError)&&<p role="alert" className="text-sm text-destructive">{errorMessage(verify.error??rotate.error??remove.error)}</p>}
      <Button disabled={create.isPending||!name.trim()||!secret.trim()}>Save credential</Button>
    </form>
  </div>;
}
