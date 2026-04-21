export interface PublicKeyInfo { publicKey: string; fingerprint: string }

export class IdentityClient {
  constructor(private baseURL: string) {}

  async publicKey(): Promise<PublicKeyInfo> {
    const res = await fetch(`${this.baseURL}/identity/key`);
    if (!res.ok) throw new Error(`identity publicKey: ${res.status}`);
    return (await res.json()) as PublicKeyInfo;
  }

  async verify(token: string): Promise<unknown> {
    const res = await fetch(`${this.baseURL}/identity/verify`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token }),
    });
    if (!res.ok) throw new Error(`identity verify: ${res.status}`);
    return await res.json();
  }
}
