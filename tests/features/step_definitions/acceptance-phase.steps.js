import assert from "node:assert";
import { Given, When, Then } from "@cucumber/cucumber";
import path from "node:path";
import {
	PIPELINE_DIR,
	DAGGER_BIN,
	runCommand,
	FIXTURE_DIR,
} from "./shared.steps.js";

// ============================================
// Acceptance Phase steg-definitioner
// ============================================

Given("en testmapp finns med acceptance-tester", async function () {
	this.testDir = FIXTURE_DIR;

	const goModPath = path.join(this.testDir, "go.mod");
	const fs = await import("node:fs/promises");
	try {
		await fs.access(goModPath);
	} catch {
		throw new Error(`Testmapp saknas: ${goModPath}`);
	}
});

When("acceptance-phase körs", async function () {
	const cmd = [
		DAGGER_BIN,
		"call",
		"acceptance-phase",
		`--source-dir`,
		`${this.testDir}`,
	].join(" ");
	const result = runCommand(cmd, { cwd: PIPELINE_DIR });

	this.acceptanceOutput = result.output;
	this.acceptanceError = result.error;
	this.acceptanceSuccess = result.success;
});

Then("köra alla tester utan @commit", async function () {
	assert.strictEqual(
		this.acceptanceSuccess,
		true,
		`Acceptance-phase misslyckades:
${this.acceptanceError}`,
	);
});
