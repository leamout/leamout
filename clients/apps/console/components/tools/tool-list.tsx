"use client";

import { Button } from "@leamout/ui/components/button";
import { Input } from "@leamout/ui/components/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@leamout/ui/components/table";
import Link from "next/link";
import { useState } from "react";
import type { Tool } from "@/components/tools/types";

export function ToolList({ tools = [] }: { tools?: Tool[] }) {
  const [search, setSearch] = useState("");
  const [type, setType] = useState("all");
  const filtered = tools.filter(
    (tool) =>
      `${tool.name} ${tool.description} ${tool.agent.name}`
        .toLowerCase()
        .includes(search.toLowerCase()) &&
      (type === "all" || tool.type === type),
  );
  return (
    <div className="space-y-6">
      <div className="flex flex-wrap justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">Tools</h1>
          <p className="text-muted-foreground">
            Manage webhook actions and built-in call controls for your agents.
          </p>
        </div>
        <Button render={<Link href="/tools/new" />}>Create tool</Button>
      </div>
      <div className="flex flex-wrap gap-4">
        <div className="flex-1 space-y-2">
          <label htmlFor="tool-search" className="text-sm font-medium">
            Search tools
          </label>
          <Input
            id="tool-search"
            type="search"
            placeholder="Name, description, or agent"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
        <div className="space-y-2">
          <label
            htmlFor="tool-type-filter"
            className="block text-sm font-medium"
          >
            Type
          </label>
          <select
            id="tool-type-filter"
            className="h-9 rounded-md border bg-background px-3"
            value={type}
            onChange={(event) => setType(event.target.value)}
          >
            <option value="all">All types</option>
            <option value="webhook">Webhook</option>
            <option value="builtin">Built-in</option>
          </select>
        </div>
      </div>
      <div className="overflow-hidden rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              {["Name", "Type", "Agent", "Status", "Timeout"].map((label) => (
                <TableHead key={label}>{label}</TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.map((tool) => (
              <TableRow key={tool.id}>
                <TableCell>
                  <Link
                    className="font-medium hover:underline"
                    href={`/tools/${encodeURIComponent(tool.id)}`}
                  >
                    {tool.name}
                  </Link>
                  <p className="text-sm text-muted-foreground">
                    {tool.description}
                  </p>
                </TableCell>
                <TableCell>{tool.type}</TableCell>
                <TableCell>
                  <Link
                    className="hover:underline"
                    href={`/agents/${encodeURIComponent(tool.agent.id)}`}
                  >
                    {tool.agent.name}
                  </Link>
                </TableCell>
                <TableCell>{tool.enabled ? "Enabled" : "Disabled"}</TableCell>
                <TableCell>{tool.timeoutMs} ms</TableCell>
              </TableRow>
            ))}
            {filtered.length === 0 && (
              <TableRow>
                <TableCell colSpan={5} className="h-40 text-center">
                  <p className="font-medium">
                    {tools.length ? "No matching tools" : "No tools to display"}
                  </p>
                  <p className="mt-2 text-muted-foreground">
                    {tools.length
                      ? "Try another search or filter."
                      : "Agent tools will appear here when account integration is connected."}
                  </p>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <p className="text-sm text-muted-foreground">
        UI preview. Tool execution is not connected.
      </p>
    </div>
  );
}
