export interface Version { id: string; entityId: string; timestamp: string; data: unknown }

export interface CollectionQuery { limit?: number; offset?: number; sort?: string; search?: string }

export class Collection<T extends { id: string }> {
  constructor(private baseURL: string, private type: string) {}

  private url(path = ""): string {
    return `${this.baseURL}/${this.type}${path}`;
  }

  private async request<R>(method: string, path: string, body?: unknown): Promise<R> {
    const res = await fetch(this.url(path), {
      method,
      headers: { "Content-Type": "application/json" },
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
    if (!res.ok) throw new Error(`${method} ${this.type}${path}: ${res.status}`);
    if (res.status === 204) return undefined as R;
    return (await res.json()) as R;
  }

  async create(data: Omit<T, "id"> | T): Promise<T> {
    return this.request<T>("POST", "/", data);
  }

  async get(id: string): Promise<T> {
    return this.request<T>("GET", `/${encodeURIComponent(id)}`);
  }

  async list(q?: CollectionQuery): Promise<T[]> {
    const p = new URLSearchParams();
    if (q?.limit !== undefined) p.set("limit", String(q.limit));
    if (q?.offset !== undefined) p.set("offset", String(q.offset));
    if (q?.sort) p.set("sort", q.sort);
    if (q?.search) p.set("search", q.search);
    const qs = p.toString();
    return this.request<T[]>("GET", qs ? `/?${qs}` : "/");
  }

  async update(id: string, data: T): Promise<T> {
    return this.request<T>("PUT", `/${encodeURIComponent(id)}`, data);
  }

  async delete(id: string): Promise<void> {
    return this.request<void>("DELETE", `/${encodeURIComponent(id)}`);
  }

  async history(id: string): Promise<Version[]> {
    return this.request<Version[]>("GET", `/${encodeURIComponent(id)}/versions`);
  }

  async revert(id: string, versionId: string): Promise<T> {
    return this.request<T>(
      "POST",
      `/${encodeURIComponent(id)}/revert/${encodeURIComponent(versionId)}`,
    );
  }
}
