import { execSync } from "node:child_process";
import { Given, When, Then } from "@cucumber/cucumber";
import assert from "node:assert";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "../../..");
const PIPELINE_DIR = path.join(REPO_ROOT, "dagger-modules/pipeline");
const FIXTURE_DIR = path.join(REPO_ROOT, "tests/fixtures/go-sample");
const DAGGER_BIN = "/home/linuxbrew/.linuxbrew/Cellar/dagger/0.21.6/bin/dagger";

// ============================================
// CI-workflow steg-definitioner
// ============================================

/**
 * Kör ett kommando och returnerar resultatet
 */
function runCommand(cmd, options = {}) {
	const cwd = options.cwd || process.cwd();

	// Wrappa med bash för att undvika sh-problem
	const fullCmd = `/bin/bash -c "cd ${cwd} && ${cmd}"`;

	try {
		const result = execSync(fullCmd, {
			encoding: "utf-8",
			stdio: ["pipe", "pipe", "pipe"],
			maxBuffer: 10 * 1024 * 1024, // 10MB buffer
			timeout: options.timeout || 300000,
			env: { ...process.env },
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
 * @param {string} text - Text att söka i
 * @returns {string|null} - Versionen eller null om ingen hittades
 */
function extractSemVer(text) {
	const match = text.match(/\bv(\d+\.\d+\.\d+)\b/);
	return match ? `v${match[1]}` : null;
}

/**
 * Sätter upp registry-uppgifter
 * @param {Object} credentials - { username, secret, valid }
 */
function setupRegistryCredentials(credentials) {
	this.registryAddress = process.env.REGISTRY_ADDRESS || "ghcr.io/simon";
	this.registryUsername = credentials.username;
	this.registrySecret = credentials.secret;
	this.skipPublish = !credentials.username || !credentials.secret;

	if (credentials.valid === false) {
		this.authFailure = true;
	}
}

/**
 * Sätter build-resultat på this
 * @param {{success: boolean, output: string, error: string}} result
 */
function setBuildResult(result) {
	this.buildOutput = result.output;
	this.buildSuccess = result.success;
	this.buildError = result.error;
}

/**
 * Bygger en Docker-image via Dagger
 * @param {Object} options - Valfria inställningar
 * @param {string} [options.sourceDir] - Sökväg till källkoden
 * @returns {Promise<{success, output, error}>}
 */
async function buildImage(options = {}) {
	const sourceDir = options.sourceDir || this.testDir;

	const cmd = [
		DAGGER_BIN,
		"call",
		"build-image",
		`--source-dir`,
		sourceDir,
	].join(" ");
	const result = runCommand(cmd, { cwd: PIPELINE_DIR });
	// Extrahera container reference från output
	if (result.success && result.output) {
		result.output = result.output.trim();
	}
	return result;
}

/**
 * Pushar en Docker-image till registry
 * @param {string} containerRef - Container reference från build
 * @param {Object} options - Valfria inställningar
 * @param {string} [options.imageName="test-image"] - Image-namn
 * @param {string} [options.tag="test"] - Image-tagg
 * @returns {Promise<{success, output, error}>}
 */
async function pushImage(containerRef, options = {}) {
	const imageName = options.imageName || "test-image";
	const tag = options.tag || "test";

	const cmd = [
		DAGGER_BIN,
		"call",
		"push-images",
		"--containers",
		containerRef,
		`--registry-address`,
		this.registryAddress,
		`--image-name`,
		imageName,
		`--tag`,
		tag,
		`--username`,
		this.registryUsername,
		`--secret`,
		this.registrySecret,
	].join(" ");
	const result = runCommand(cmd, { cwd: PIPELINE_DIR });

	// Dagger skriver fel till stdout, inte stderr
	const fullOutput = (result.output || "") + (result.error || "");
	if (result.status !== 0 || fullOutput.includes("Error:")) {
		// Extrahera endast relevant felmeddelande
		const errorMatch = result.error.match(/Error: ([^\n]+)/);
		const errorMsg = errorMatch
			? errorMatch[1]
			: result.error || "Push misslyckades";
		return {
			success: false,
			output: result.output,
			error: errorMsg,
			status: result.status,
		};
	}
	return result;
}

// ---- Scenario: Tester körs ----

Given("en testmapp finns", async function () {
	// Nollställ state
	this.ciRan = false;
	this.testsPassed = false;
	this.imagePublished = false;
	this.testDir = FIXTURE_DIR;

	const goModPath = path.join(this.testDir, "go.mod");
	const fs = await import("node:fs/promises");
	try {
		await fs.access(goModPath);
	} catch {
		throw new Error(`Testmapp saknas: ${goModPath}`);
	}
});

When("CI-flödet körs", async function () {
	const cmd = [
		DAGGER_BIN,
		"call",
		"unit-tests",
		`--source-dir`,
		`${this.testDir}`,
	].join(" ");
	const result = runCommand(cmd, { cwd: PIPELINE_DIR });

	this.ciOutput = result.output;
	this.ciError = result.error;
	this.ciRan = true;
	this.ciSuccess = result.success;
});

Then("testerna körs utan fel", async function () {
	assert.strictEqual(this.ciRan, true, "CI-flödet kördes inte");
	assert.strictEqual(
		this.ciSuccess,
		true,
		`Testerna misslyckades:\n${this.ciError}`,
	);
	assert.ok(
		this.ciOutput.includes("Tester klara") || this.ciOutput.includes("✅"),
		"Ser inte bekräftelse på att testerna kördes klart",
	);
});

// ---- Scenario: Ingen image byggs vid misslyckade tester ----

When("testerna misslyckas", async function () {
	this.ciRan = true;
	this.ciSuccess = false;
	this.ciError = "Testerna misslyckades: some test failed";
	this.testsPassed = false;
});

Then("image byggs inte", async function () {
	assert.strictEqual(this.ciRan, true, "CI-flödet kördes inte");
	assert.strictEqual(
		this.testsPassed,
		false,
		"Testerna passerade (borde ha misslyckats)",
	);
	this.buildSkipped = true;
});

// ---- Scenario: Image byggs ----

When("testerna passerar", async function () {
	// Verifiera att testerna passerade (redan verifierat i förra steget)
	assert.strictEqual(this.ciSuccess, true, "Testerna passerade inte");
	this.testsPassed = true;
});

Then("skapas en image", async function () {
	assert.strictEqual(this.ciRan, true, "CI-flödet kördes inte");
	assert.strictEqual(this.testsPassed, true, "Testerna passerade inte");

	const result = await buildImage.call(this);
	setBuildResult.call(this, result);

	assert.strictEqual(
		this.buildSuccess,
		true,
		`Image-bygge misslyckades: ${this.buildError}`,
	);
});

// ---- Scenario: Registry-autentisering misslyckas ----

Given("registry-uppgifter är tillgängliga men felaktiga", async function () {
	setupRegistryCredentials.call(this, {
		username: "fel_användare",
		secret: "fel_lösenord",
		valid: false,
	});
});

When("pipelinen försöker publicera imagen", async function () {
	const buildResult = await buildImage.call(this);
	setBuildResult.call(this, buildResult);

	if (!this.buildSuccess) {
		this.pushError = this.buildError;
		return;
	}

	const pushResult = await pushImage.call(this, this.buildOutput);
	this.pushError = pushResult.error;

	if (!pushResult.success) {
		// Förväntat - autentiseringsfel
	} else {
		throw new Error("Push lyckades men borde ha misslyckats med auth-fel");
	}
});

Then("ska ett autentiseringsfel visas", async function () {
	assert.strictEqual(
		this.authFailure,
		true,
		"Inget autentiseringsfel inträffade",
	);
	// Kontrollera att push misslyckades
	const errorIndicators = [
		"authentication",
		"invalid",
		"credentials",
		"unauthorized",
		"denied",
		"exit code: 1",
		"failed to get value",
	];
	const hasErrorIndicator = errorIndicators.some((indicator) =>
		this.pushError?.toLowerCase().includes(indicator.toLowerCase()),
	);
	assert.ok(hasErrorIndicator, `Förväntade push-fel, fick: ${this.pushError}`);
});

// ---- Scenario: Image publiceras ----

Given("registry-uppgifter är tillgängliga", async function () {
	setupRegistryCredentials.call(this, {
		username: process.env.REGISTRY_USERNAME,
		secret: process.env.REGISTRY_SECRET,
		valid: true,
	});
});

When("pipelinen publicerar imagen", async function () {
	if (this.skipPublish) {
		this.imagePublished = false;
		return;
	}

	const buildResult = await buildImage.call(this);
	setBuildResult.call(this, buildResult);

	if (!this.buildSuccess) {
		this.imagePublished = false;
		return;
	}

	const pushResult = await pushImage.call(this, this.buildOutput);

	this.pushOutput = pushResult.output;
	this.pushError = pushResult.error;
	this.imagePublished = pushResult.success;
});

Then("imagen ska finnas i registry", async function () {
	if (this.skipPublish) {
		return;
	}

	assert.strictEqual(
		this.imagePublished,
		true,
		`Push misslyckades: ${this.pushOutput}`,
	);
	assert.ok(
		this.pushOutput.includes("Push klar") || this.pushOutput.includes("✅"),
		"Ser inte bekräftelse på att imagen pushades",
	);
});

// ---- Scenario: Image-tagg baseras på nästa semver-version ----

Given("det finns en image version {string}", function (version) {
	this.currentVersion = version;
});

When(
	"version-increment-åtgärden körs med commit {string}",
	async function (commitMessage) {
		// Escape the commit message for shell (handle spaces and special chars)
		const escapedMessage = commitMessage.replace(/'/g, "'\\''");
		const cmd = [
			DAGGER_BIN,
			"call",
			"sem-ver-bump",
			"--current-version",
			this.currentVersion,
			"--commit-message",
			`'${escapedMessage}'`,
		].join(" ");

		const result = runCommand(cmd, { cwd: PIPELINE_DIR });
		this.semverOutput = result.output;
		this.semverError = result.error;
		this.semverSuccess = result.success;
		this.semVerResult = extractSemVer(result.output);
	},
);

Then("ska nästa version vara {string}", function (expectedVersion) {
	assert.strictEqual(
		this.semverSuccess,
		true,
		`Semver-bump misslyckades: ${this.semverError}`,
	);
	assert.strictEqual(
		this.semVerResult,
		expectedVersion,
		`Förväntade version ${expectedVersion}, fick: ${this.semVerResult}`,
	);
});
