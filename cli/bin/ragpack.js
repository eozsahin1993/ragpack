#!/usr/bin/env node
import { Command } from "commander";
import { spawnSync } from "child_process";
import { existsSync, copyFileSync } from "fs";
import { join, dirname } from "path";
import { fileURLToPath } from "url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const TEMPLATES = join(__dirname, "..", "templates");

function compose(args) {
  const result = spawnSync("docker", ["compose", ...args], { stdio: "inherit" });
  if (result.status !== 0) process.exit(result.status ?? 1);
}

function requireInit() {
  if (!existsSync("docker-compose.yml")) {
    console.error("No docker-compose.yml found. Run `ragpack init` first.");
    process.exit(1);
  }
}

const program = new Command();

program
  .name("ragpack")
  .description("Self-hosted RAG and semantic search engine")
  .version("0.1.0");

program
  .command("init")
  .description("Scaffold docker-compose.yml and .env in the current directory")
  .action(() => {
    if (existsSync("docker-compose.yml")) {
      console.log("docker-compose.yml already exists, skipping.");
    } else {
      copyFileSync(join(TEMPLATES, "docker-compose.yml"), "docker-compose.yml");
      console.log("✓ Created docker-compose.yml");
    }
    if (!existsSync(".env")) {
      copyFileSync(join(TEMPLATES, ".env.example"), ".env");
      console.log("✓ Created .env — edit it with your settings before starting.");
    }
  });

program
  .command("start")
  .description("Start the RagPack stack")
  .option("--profile <profile>", "Embedding provider profile (ollama, tei)")
  .option("--dev", "Start in dev mode with hot reload (requires cloned repo)")
  .option("--build", "Force rebuild images")
  .action((opts) => {
    requireInit();
    const args = ["up", "-d"];
    if (opts.build) args.push("--build");
    if (opts.profile) args.push("--profile", opts.profile);
    if (opts.dev) {
      compose(["-f", "docker-compose.yml", "-f", "docker-compose.dev.yml", ...args]);
    } else {
      compose(args);
    }
  });

program
  .command("stop")
  .description("Stop the RagPack stack")
  .option("-v, --volumes", "Also remove volumes (deletes all data)")
  .action((opts) => {
    requireInit();
    compose(opts.volumes ? ["down", "-v"] : ["down"]);
  });

program
  .command("logs")
  .description("Tail logs from the stack")
  .argument("[service]", "Service to tail (backend, web-admin, ollama)")
  .option("-n, --lines <n>", "Number of lines to show", "50")
  .action((service, opts) => {
    requireInit();
    const args = ["logs", "-f", "--tail", opts.lines];
    if (service) args.push(service);
    compose(args);
  });

program
  .command("update")
  .description("Pull latest images and restart")
  .action(() => {
    requireInit();
    compose(["pull"]);
    compose(["up", "-d"]);
  });

program.parse();
