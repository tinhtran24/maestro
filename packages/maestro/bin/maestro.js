#!/usr/bin/env node
// Pure-Node shim: resolve the per-platform optionalDependency that holds the
// prebuilt Go `to` binary for this host, then exec it transparently.
// Zero install scripts; zero third-party deps. The binary is delivered by npm
// installing only the matching `@tinhtran24/maestro-<platform>-<arch>` package (its
// os/cpu fields gate the rest out).

"use strict";

const { spawnSync } = require("node:child_process");
const path = require("node:path");

// npm cpu names match process.arch (x64/arm64); npm os names match
// process.platform (darwin/win32/linux). Our platform packages are named
// `@tinhtran24/maestro-<platform>-<arch>` to mirror that exactly.
const platform = process.platform;
const arch = process.arch;
const pkg = `@tinhtran24/maestro-${platform}-${arch}`;
const binName = platform === "win32" ? "maestro.exe" : "maestro";

// Deprecation notice when invoked via the legacy `to` command name.
if (path.basename(process.argv[1] || "").replace(/\.js$/, "") === "to" || process.env.MAESTRO_INVOKED_AS === "to") {
  process.stderr.write("`to` is deprecated and will be removed; use `maestro` (Thanos is now Maestro).\n");
}

function resolveBinary() {
  // require.resolve the platform package's package.json to find its install
  // dir (works whether hoisted to a parent node_modules or nested), then join
  // the binary path. The platform package ships the binary under bin/.
  let pkgJsonPath;
  try {
    pkgJsonPath = require.resolve(`${pkg}/package.json`);
  } catch {
    return null;
  }
  return path.join(path.dirname(pkgJsonPath), "bin", binName);
}

const binary = resolveBinary();

if (!binary) {
  process.stderr.write(
    `@tinhtran24/maestro: no prebuilt binary for ${platform}-${arch}.\n` +
      `The optional dependency ${pkg} is not installed, which usually means\n` +
      `this platform is unsupported. Supported: darwin-arm64, darwin-x64,\n` +
      `win32-x64, linux-x64.\n`,
  );
  process.exit(1);
}

const result = spawnSync(binary, process.argv.slice(2), { stdio: "inherit" });

if (result.error) {
  if (result.error.code === "ENOENT") {
    process.stderr.write(
      `@tinhtran24/maestro: binary not found at ${binary}.\n` +
        `Reinstall @tinhtran24/maestro to restore the platform package.\n`,
    );
  } else {
    process.stderr.write(`@tinhtran24/maestro: failed to run binary: ${result.error.message}\n`);
  }
  process.exit(1);
}

// Propagate signal-terminations as a conventional 128+signal code, else the
// child's own exit code.
if (result.signal) {
  process.exit(1);
}
process.exit(result.status === null ? 1 : result.status);
