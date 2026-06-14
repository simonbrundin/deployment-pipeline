Feature: Acceptance Phase

  Som utvecklare
  Vill jag att Acceptance Phase-flödet kör acceptance-tester
  För att verifiera att koden fungerar korrekt innan release

  Scenario: Acceptance-tester körs
    Given en testmapp finns med acceptance-tester
    When acceptance-phase körs
    Then köra alla tester utan @commit
