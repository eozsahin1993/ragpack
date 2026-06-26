#!/usr/bin/env node
import { Command } from "commander";
import { spawnSync } from "child_process";
import { existsSync, copyFileSync, readFileSync, appendFileSync } from "fs";
import { join, dirname, basename } from "path";
import { fileURLToPath } from "url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const TEMPLATES = join(__dirname, "..", "templates");
const COMPOSE_FILE = join(TEMPLATES, "docker-compose.yml");
const { version } = JSON.parse(readFileSync(join(__dirname, "../package.json"), "utf8"));

const ENV_FILE = join(process.cwd(), ".env.ragpack");
const PROJECT_NAME = basename(process.cwd());

const BANNER = `
\x1b[96m██████╗  █████╗  ██████╗ ██████╗  █████╗  ██████╗██╗  ██╗\x1b[0m
\x1b[96m██╔══██╗██╔══██╗██╔════╝ ██╔══██╗██╔══██╗██╔════╝██║ ██╔╝\x1b[0m
\x1b[36m██████╔╝███████║██║  ███╗██████╔╝███████║██║     █████╔╝ \x1b[0m
\x1b[36m██╔══██╗██╔══██║██║   ██║██╔═══╝ ██╔══██║██║     ██╔═██╗ \x1b[0m
\x1b[34m██║  ██║██║  ██║╚██████╔╝██║     ██║  ██║╚██████╗██║  ██╗\x1b[0m
\x1b[34m╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝     ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝\x1b[0m

\x1b[2m  Self-hosted RAG & semantic search\x1b[0m  \x1b[36m·\x1b[0m  \x1b[1mv${version}\x1b[0m
`;

function compose(args) {
  const result = spawnSync(
    "docker",
    ["compose", "--project-name", PROJECT_NAME, "-f", COMPOSE_FILE, "--env-file", ENV_FILE, ...args],
    {
      stdio: "inherit",
      env: { ...process.env, RAGPACK_ENV_FILE: ENV_FILE },
    }
  );
  if (result.status !== 0) process.exit(result.status ?? 1);
}

function requireInit() {
  if (!existsSync(ENV_FILE)) {
    console.error("No .env.ragpack found. Run `ragpack init` first.");
    process.exit(1);
  }
}

const program = new Command();

program
  .name("ragpack")
  .addHelpText("beforeAll", BANNER)
  .version(version);

program
  .command("init")
  .description("Create .env.ragpack in the current directory")
  .action(() => {
    if (existsSync(ENV_FILE)) {
      console.log(".env.ragpack already exists, skipping.");
    } else {
      copyFileSync(join(TEMPLATES, ".env.ragpack.example"), ENV_FILE);
      console.log("✓ Created .env.ragpack — edit it with your settings before starting.");
    }
    if (existsSync(".gitignore")) {
      const content = readFileSync(".gitignore", "utf8");
      if (!content.includes(".env.ragpack")) {
        appendFileSync(".gitignore", "\n.env.ragpack\n");
        console.log("✓ Added .env.ragpack to .gitignore");
      }
    }
  });

program
  .command("start")
  .description("Start the RagPack stack")
  .option("--profile <profile>", "Embedding provider profile (ollama, tei)")
  .option("--build", "Force rebuild images")
  .action((opts) => {
    requireInit();
    const preArgs = opts.profile ? ["--profile", opts.profile] : [];
    const args = [...preArgs, "up", "-d"];
    if (opts.build) args.push("--build");
    compose(args);
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

program
  .command("eject")
  .description("Copy docker-compose.yml into the current directory for customization")
  .action(() => {
    if (existsSync("docker-compose.yml")) {
      console.error("docker-compose.yml already exists.");
      process.exit(1);
    }
    copyFileSync(COMPOSE_FILE, "docker-compose.yml");
    console.log("✓ Created docker-compose.yml — ragpack will use this file from now on.");
  });

program.parse();
