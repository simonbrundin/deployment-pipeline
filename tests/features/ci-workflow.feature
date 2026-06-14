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
