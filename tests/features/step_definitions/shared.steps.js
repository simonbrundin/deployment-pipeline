import { execSync } from "node:child_process";
import { Given } from "@cucumber/cucumber";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "../../..");
const PIPELINE_DIR = path.join(REPO_ROOT, "dagger-modules/pipeline");
const FIXTURE_DIR = path.join(REPO_ROOT, "tests/fixtures/go-sample");

// Hitta dagger-binär - kolla miljövariabel först, annars sök i vanliga platser
function findDaggerBinary() {
	const possiblePaths = [
		"/home/linuxbrew/.linuxbrew/bin/dagger",
		"/usr/local/bin/dagger",
		"/usr/bin/dagger",
		"dagger", // I PATH
	];

	for (const p of possiblePaths) {
		try {
			execSync(`test -x ${p}`, { stdio: "ignore" });
			return p;
		} catch {
			// Fortsätt till nästa
		}
	}

	// Fallback till PATH
	return "dagger";
}

const DAGGER_BIN = process.env.DAGGER_BIN || findDaggerBinary();

// ============================================
// Delade hjälpfunktioner
// ============================================

/**
 * Kör ett kommando och returnerar resultatet
 */
function runCommand(cmd, options = {}) {
	const cwd = options.cwd || process.cwd();

	// Wrappa med bash för att undvika sh-problem
	const fullCmd = `/bin/bash -c "cd ${cwd} && ${cmd}"`;

	// Lägg till extra env-variabler om angivet
	const env = { ...process.env };
	if (options.env) {
		Object.assign(env, options.env);
	}

	try {
		const result = execSync(fullCmd, {
			encoding: "utf-8",
			stdio: ["pipe", "pipe", "pipe"],
			maxBuffer: 10 * 1024 * 1024, // 10MB buffer
			timeout: options.timeout || 300000,
			env,
		});
		return { success: true, output: result, error: "" };
	} catch (error) {
		return {
			success: false,
			output: error.stdout || "",
			error: error.stderr || error.message || "",
			status: error.status || -1,
		};
	}
}

/**
 * Extraherar semver-version (t.ex. v1.2.3) från text
 */
function extractSemVer(text) {
	const match = text.match(/\bv(\d+\.\d+\.\d+)\b/);
	return match ? `v${match[1]}` : null;
}

// Exportera för användning i andra filer
export {
	REPO_ROOT,
	PIPELINE_DIR,
	FIXTURE_DIR,
	DAGGER_BIN,
	runCommand,
	extractSemVer,
};

// ============================================
// Delade steg
// ============================================

Given("en testmapp finns", async function () {
	this.testDir = FIXTURE_DIR;

	const goModPath = path.join(this.testDir, "go.mod");
	const fs = await import("node:fs/promises");
	try {
		await fs.access(goModPath);
	} catch {
		throw new Error(`Testmapp saknas: ${goModPath}`);
	}
});
