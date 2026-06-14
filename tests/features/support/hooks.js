import { Before } from "@cucumber/cucumber";

Before(async function () {
	// Nollställ state före varje scenario
	this.ciRan = false;
	this.testsPassed = false;
	this.imagePublished = false;
	this.testDir = process.cwd();
});
