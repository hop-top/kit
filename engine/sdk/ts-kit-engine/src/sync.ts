export interface RemoteStatus { name: string; url: string; mode: "push" | "pull" | "both"; lastSync?: string }

export class SyncClient {
  constructor(private baseURL: string) {}

  async addRemote(name: string, url: string, mode: "push" | "pull" | "both" = "both"): Promise<void> {
    const res = await fetch(`${this.baseURL}/sync/remotes`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, url, mode }),
    });
    if (!res.ok) throw new Error(`addRemote: ${res.status}`);
  }

  async removeRemote(name: string): Promise<void> {
    const res = await fetch(`${this.baseURL}/sync/remotes/${encodeURIComponent(name)}`, {
      method: "DELETE",
    });
    if (!res.ok) throw new Error(`removeRemote: ${res.status}`);
  }

  async status(): Promise<RemoteStatus[]> {
    const res = await fetch(`${this.baseURL}/sync/remotes`);
    if (!res.ok) throw new Error(`sync status: ${res.status}`);
    return (await res.json()) as RemoteStatus[];
  }
}
