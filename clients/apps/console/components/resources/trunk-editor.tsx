"use client";

import { Button } from "@leamout/ui/components/button";
import { Input } from "@leamout/ui/components/input";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api,errorMessage } from "@/lib/api";
import { resourceKeys,type SipTrunk,type TrunkEndpoint } from "@/lib/resources";
import { RequestError,RequestLoading } from "./request-state";

export function TrunkEditor({id}:{id?:string}) {
  const {selected}=useOrganization();
  const orgId=selected?.id??"";
  const client=useQueryClient();
  const router=useRouter();
  const detail=useQuery({queryKey:[...resourceKeys.trunks(orgId),id],queryFn:({signal})=>api<SipTrunk>(`/trunks/${id}`,{organizationId:orgId,signal}),enabled:Boolean(id&&orgId)});
  if(id&&detail.isPending)return <RequestLoading what="SIP trunk"/>;
  if(id&&detail.isError)return <RequestError error={detail.error} retry={()=>void detail.refetch()}/>;
  return <TrunkFields key={id??"new"} id={id} existing={detail.data} orgId={orgId} onDone={async (trunk)=>{
    await client.invalidateQueries({queryKey:resourceKeys.trunks(orgId)});
    router.push(`/sip-trunks/${trunk.id}`);
  }} onDelete={async ()=>{
    await client.invalidateQueries({queryKey:resourceKeys.trunks(orgId)});
    router.push("/sip-trunks");
  }}/>;
}
function TrunkFields({id,existing,orgId,onDone,onDelete}:{id?:string;existing?:SipTrunk;orgId:string;onDone:(trunk:SipTrunk)=>void;onDelete:()=>void}) {
  const [name,setName]=useState(existing?.name??"");
  const [direction,setDirection]=useState(existing?.direction??"bidirectional");
  const [status,setStatus]=useState(existing?.status??"disabled");
  const [maxCps,setMaxCps]=useState(existing?.max_cps?String(existing.max_cps):"");
  const [maxCalls,setMaxCalls]=useState(existing?.max_concurrent_calls?String(existing.max_concurrent_calls):"");
  const [host,setHost]=useState("");
  const [port,setPort]=useState("5060");
  const [transport,setTransport]=useState("udp");
  const [endpointError,setEndpointError]=useState("");
  const client=useQueryClient();
  const endpointKey=["resources","trunk-endpoints",orgId,id];
  const endpoints=useQuery({queryKey:endpointKey,queryFn:({signal})=>api<{endpoints:TrunkEndpoint[]}>(`/trunks/${id}/endpoints`,{organizationId:orgId,signal}),enabled:Boolean(id)});
  const save=useMutation({
    mutationFn:()=>api<SipTrunk>(id?`/trunks/${id}`:"/trunks",{organizationId:orgId,method:id?"PATCH":"POST",body:{
      name:name.trim(),direction,status,
      ...(maxCps?{max_cps:Number(maxCps)}:{}),
      ...(maxCalls?{max_concurrent_calls:Number(maxCalls)}:{}),
    }}),
    onSuccess:onDone,
  });
  const remove=useMutation({mutationFn:()=>api<void>(`/trunks/${id}`,{method:"DELETE",organizationId:orgId}),onSuccess:onDelete});
  const addEndpoint=useMutation({
    mutationFn:()=>api<TrunkEndpoint>(`/trunks/${id}/endpoints`,{method:"POST",organizationId:orgId,body:{host:host.trim(),port:Number(port),transport}}),
    onSuccess:async()=>{setHost("");setEndpointError("");await client.invalidateQueries({queryKey:endpointKey});},
  });
  const deleteEndpoint=useMutation({
    mutationFn:(endpointId:string)=>api<void>(`/trunks/${id}/endpoints/${endpointId}`,{method:"DELETE",organizationId:orgId}),
    onSuccess:()=>void client.invalidateQueries({queryKey:endpointKey}),
  });
  return <div className="max-w-2xl space-y-8">
    <Link href="/sip-trunks" className="text-sm underline">Back to SIP trunks</Link>
    <form className="space-y-5" onSubmit={e=>{e.preventDefault();if(name.trim())save.mutate();}}>
      <h1 className="text-3xl font-semibold">{id?"SIP trunk configuration":"Create SIP trunk"}</h1>
      <label className="block space-y-2 text-sm font-medium">Name<Input required maxLength={255} value={name} onChange={e=>setName(e.target.value)}/></label>
      <label className="block space-y-2 text-sm font-medium">Direction<select className="h-10 w-full rounded-md border bg-background px-3" value={direction} onChange={e=>setDirection(e.target.value as typeof direction)}><option value="bidirectional">Bidirectional</option><option value="inbound">Inbound</option><option value="outbound">Outbound</option></select></label>
      <label className="block space-y-2 text-sm font-medium">Status<select className="h-10 w-full rounded-md border bg-background px-3" value={status} onChange={e=>setStatus(e.target.value as typeof status)}><option value="disabled">Disabled</option><option value="active">Active</option></select></label>
      <label className="block space-y-2 text-sm font-medium">Max calls per second<Input type="number" min={1} value={maxCps} onChange={e=>setMaxCps(e.target.value)}/></label>
      <label className="block space-y-2 text-sm font-medium">Max concurrent calls<Input type="number" min={1} value={maxCalls} onChange={e=>setMaxCalls(e.target.value)}/></label>
      {save.isError&&<p role="alert" className="text-sm text-destructive">{errorMessage(save.error)}</p>}
      {remove.isError&&<p role="alert" className="text-sm text-destructive">{errorMessage(remove.error)}</p>}
      <div className="flex gap-3"><Button disabled={save.isPending||!name.trim()}>{id?"Save trunk":"Create trunk"}</Button>{id&&<Button type="button" variant="destructive" disabled={remove.isPending} onClick={()=>{if(window.confirm("Disable this trunk?"))remove.mutate();}}>Disable trunk</Button>}</div>
    </form>
    {id&&<section className="space-y-4 border-t pt-6">
      <h2 className="text-xl font-semibold">SIP endpoints</h2>
      {endpoints.isPending?<RequestLoading what="endpoints"/>:endpoints.isError?<RequestError error={endpoints.error} retry={()=>void endpoints.refetch()}/>:<div className="space-y-2">{endpoints.data?.endpoints.map(ep=><div className="flex items-center justify-between rounded-lg border p-3" key={ep.id}><span>{ep.host}:{ep.port} · {ep.transport}</span><Button variant="outline" disabled={deleteEndpoint.isPending} onClick={()=>{if(window.confirm("Delete this endpoint?"))deleteEndpoint.mutate(ep.id);}}>Remove</Button></div>)}{!endpoints.data?.endpoints.length&&<p className="text-sm text-muted-foreground">No endpoints configured.</p>}</div>}
      <form className="space-y-3" onSubmit={e=>{e.preventDefault();setEndpointError("");if(host.trim())addEndpoint.mutate();}}>
        <label className="block space-y-2 text-sm font-medium">SIP hostname<Input required value={host} onChange={e=>setHost(e.target.value)}/></label>
        <label className="block space-y-2 text-sm font-medium">Port<Input type="number" min={1} max={65535} required value={port} onChange={e=>setPort(e.target.value)}/></label>
        <label className="block space-y-2 text-sm font-medium">Transport<select className="h-10 w-full rounded-md border bg-background px-3" value={transport} onChange={e=>setTransport(e.target.value)}><option value="udp">UDP</option><option value="tcp">TCP</option><option value="tls">TLS</option></select></label>
        {(addEndpoint.isError||deleteEndpoint.isError||endpointError)&&<p role="alert" className="text-sm text-destructive">{endpointError||errorMessage(addEndpoint.error??deleteEndpoint.error)}</p>}
        <Button disabled={addEndpoint.isPending||!host.trim()}>Add endpoint</Button>
      </form>
      <p className="text-sm text-muted-foreground">Inbound/outbound authentication and routing policies are configured separately.</p>
    </section>}
  </div>;
}
