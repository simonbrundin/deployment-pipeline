# Deployment Pipeline

Paketerar koden och verifierar att den är bra nog att lansera

## Flöde

1. Commit Phase - Trigger (`git push`)
   1. Källkod hämtas - _cachas_
   2. Dependencies installeras - _cachas_
   3. Lint
   4. Enhetstester körs
   5. Artifact (image) byggs - _cachas_
   6. Image signeras
2. Acceptance Phase - Trigger (Commit Phase passed)
   1. Produktionslik miljö sätts upp
   2. Acceptanstester körs
   3. Artifact får tummen upp
      - Image - Ny tagg sätts
      -

## Användning i andra repos

```yaml
# .github/workflows/ci.yaml
name: CI

on: push

jobs:
  commit:
    uses: simonbrundin/deployment-pipeline/.github/workflows/commit-phase.yaml@v1
    with:
      source-dir: . # Valfri, default: .
      image-name: "" # Valfri, default: repo-namn

  acceptance:
    needs: commit
    uses: simonbrundin/deployment-pipeline/.github/workflows/acceptance-phase.yaml@v1
    with:
      source-dir: . # Valfri, default: .
```

## Krävs i anropande repo

**Secrets:**

- `DAGGER_CLOUD_TOKEN`

**Variables:**

- `REGISTRY_ADDRESS`
- `REGISTRY_USERNAME`
- `REGISTRY_PASSWORD`

## Struktur

```
.github/workflows/
├── ci.yaml              # Kombinerad pipeline (för test)
├── commit-phase.yaml    # Reusable: bygg + push
└── acceptance-phase.yaml # Reusable: test + deploy
```

## Versionering

Använd taggar (`@v1`, `@v2`) för att låsa version. Se till att taggen finns i
detta repo innan du uppdaterar referenser i produktionsrepos. detta repo innan
du uppdaterar referenser i produktionsrepos. detta repo innan du uppdaterar
referenser i produktionsrepos.
