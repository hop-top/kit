#!/usr/bin/env node
"use strict";

const { execFileSync } = require("child_process");
const { existsSync, mkdirSync, createWriteStream } = require("fs");
const { join } = require("path");
const https = require("https");

const REPO = "hop-top/kit";
const BIN_DIR = join(__dirname, "..", "bin");
const BIN_NAME = process.platform === "win32" ? "kit.exe" : "kit";

function which(name) {
  try {
    const cmd = process.platform === "win32" ? "where" : "which";
    execFileSync(cmd, [name], { stdio: "ignore" });
    return true;
  } catch {
    return false;
  }
}

function platformKey() {
  const os = { darwin: "darwin", linux: "linux", win32: "windows" }[process.platform];
  const arch = { x64: "amd64", arm64: "arm64" }[process.arch];
  if (!os || !arch) throw new Error(`Unsupported platform: ${process.platform}/${process.arch}`);
  return `${os}_${arch}`;
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const follow = (u) => {
      https.get(u, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          follow(res.headers.location);
          return;
        }
        if (res.statusCode !== 200) {
          reject(new Error(`Download failed: ${res.statusCode}`));
          return;
        }
        const file = createWriteStream(dest, { mode: 0o755 });
        res.pipe(file);
        file.on("finish", () => file.close(resolve));
      }).on("error", reject);
    };
    follow(url);
  });
}

async function main() {
  if (which("kit")) {
    console.log("kit-engine: found kit in PATH, skipping download");
    return;
  }

  const binPath = join(BIN_DIR, BIN_NAME);
  if (existsSync(binPath)) {
    console.log("kit-engine: binary already downloaded");
    return;
  }

  const key = platformKey();
  const url = `https://github.com/${REPO}/releases/latest/download/kit_${key}`;

  console.log(`kit-engine: downloading kit for ${key}...`);
  mkdirSync(BIN_DIR, { recursive: true });
  await download(url, binPath);
  console.log("kit-engine: download complete");
}

main().catch((err) => {
  console.error(`kit-engine postinstall: ${err.message}`);
  process.exit(1);
});
