import { Command } from "commander";

export function createProgram(): Command {
  const program = new Command()
    .name("{{app_name}}")
    .description("{{description}}")
    .version("0.0.0")
    .option("-f, --format <type>", "output format", "text")
    .option("-v, --verbose", "verbose output");

  return program;
}
