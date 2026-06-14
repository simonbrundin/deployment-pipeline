Feature: CI-workflow

  Som utvecklare
  Vill jag att CI-flödet kör tester automatiskt
  För att få snabb återkoppling om koden fungerar

  @ci @unit-tests
  Scenario: Tester körs
    Given en testmapp finns
    When CI-flödet körs
    Then testerna körs utan fel

  @ci @build
  Scenario: Image byggs
    When CI-flödet körs
    And testerna passerar
    Then skapas en image

  @ci @build
  Scenario: Ingen image byggs vid misslyckade tester
    When testerna misslyckas
    Then image byggs inte

  @ci @publish
  Scenario: Image publiceras
    Given registry-uppgifter är tillgängliga
    When pipelinen publicerar imagen
    Then imagen ska finnas i registry

  @ci @publish
  Scenario: Registry-autentisering misslyckas
    Given registry-uppgifter är tillgängliga men felaktiga
    When pipelinen försöker publicera imagen
    Then ska ett autentiseringsfel visas

  @ci @version @semver
  Scenario: Image-tagg baseras på nästa semver-version
    Given det finns en image version "v1.0.0"
    When version-increment-åtgärden körs med commit "feat: new feature"
    Then ska nästa version vara "v1.1.0"

  @ci @version @semver
  Scenario: Patch-bump vid buggfix
    Given det finns en image version "v1.0.0"
    When version-increment-åtgärden körs med commit "fix: bug fix"
    Then ska nästa version vara "v1.0.1"

  @ci @version @semver
  Scenario: Major-bump vid breaking change
    Given det finns en image version "v1.0.0"
    When version-increment-åtgärden körs med commit "BREAKING CHANGE: new API"
    Then ska nästa version vara "v2.0.0"
