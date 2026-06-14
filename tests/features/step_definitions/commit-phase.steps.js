import assert from "node:assert";
import { Given, When, Then } from "@cucumber/cucumber";
import {
	PIPELINE_DIR,
	DAGGER_BIN,
	runCommand,
	extractSemVer,
} from "./shared.steps.js";

// ============================================
// Commit Phase steg-definitioner
// ============================================

When("CI-flödet körs", async function () {
	const cmd = [
		DAGGER_BIN,
		"call",
		"run-tests",
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

// ============================================
// Hjälpfunktioner
// ============================================

/**
 * Sätter upp registry-uppgifter
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
 */
function setBuildResult(result) {
	this.buildOutput = result.output;
	this.buildSuccess = result.success;
	this.buildError = result.error;
}

/**
 * Bygger en Docker-image via Dagger
 */
async function buildImage(options = {}) {
	const sourceDir = options.sourceDir || this.testDir;
	const { PIPELINE_DIR: PIPELINE, DAGGER_BIN: BIN } = await import(
		"./shared.steps.js"
	);

	const cmd = [BIN, "call", "build-image", `--source-dir`, sourceDir].join(" ");
	const result = runCommand(cmd, { cwd: PIPELINE });
	// Extrahera container reference från output
	if (result.success && result.output) {
		result.output = result.output.trim();
	}
	return result;
}

/**
 * Pushar en Docker-image till registry
 */
async function pushImage(containerRef, options = {}) {
	const imageName = options.imageName || "test-image";
	const tag = options.tag || "test";
	const { PIPELINE_DIR: PIPELINE, DAGGER_BIN: BIN } = await import(
		"./shared.steps.js"
	);

	const cmd = [
		BIN,
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
	const result = runCommand(cmd, { cwd: PIPELINE });

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
